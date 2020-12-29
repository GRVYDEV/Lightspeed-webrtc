// +build !js

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"flag"

	"github.com/gorilla/websocket"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media/samplebuilder"
)

var (
	videoBuilder *samplebuilder.SampleBuilder
	addr         = flag.String("addr", "localhost:8080", "http service address")
	iAddr         = flag.String("i-addr", "127.0.0.1", "Address that ingest server should listen on")
	upgrader     = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	videoTrack *webrtc.TrackLocalStaticRTP

	// lock for peerConnections and trackLocals
	listLock        sync.RWMutex
	peerConnections []peerConnectionState
	trackLocals     map[string]*webrtc.TrackLocalStaticRTP
)

type websocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

type peerConnectionState struct {
	peerConnection *webrtc.PeerConnection
	websocket      *threadSafeWriter
}

func main() {
	flag.Parse()
	fmt.Printf("Addr: %#v", *addr)
	log.SetFlags(0)
	trackLocals = map[string]*webrtc.TrackLocalStaticRTP{}

	// peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
	// 	ICEServers: []webrtc.ICEServer{
	// 		{
	// 			URLs: []string{"stun:stun.l.google.com:19302"},
	// 		},
	// 	},
	// })

	// if err != nil {
	// 	panic(err)
	// }

	
	port, err := strconv.Atoi(os.Getenv("INGEST_PORT"))

	if err != nil {
		port = 65535
	}
	// Open a UDP Listener for RTP Packets on port 5004
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(*iAddr), Port: port})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = listener.Close(); err != nil {
			panic(err)
		}
	}()

	fmt.Println("Waiting for RTP Packets, please run GStreamer or ffmpeg now")

	// Listen for a single RTP Packet, we need this to determine the SSRC
	inboundRTPPacket := make([]byte, 4096) // UDP MTU
	n, _, err := listener.ReadFromUDP(inboundRTPPacket)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Read packet")

	// Unmarshal the incoming packet
	packet := &rtp.Packet{}
	if err = packet.Unmarshal(inboundRTPPacket[:n]); err != nil {
		panic(err)
	}

	// Create a video track
	videoTrack, err = webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: "video/h264"}, "video", "pion")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Track created")

	// start HTTP server
	go func() {
		fmt.Printf("Server started?")
		http.HandleFunc("/websocket", websocketHandler)

		log.Fatal(http.ListenAndServe(*addr, nil))
	}()

	// transceiver, err := peerConnection.AddTransceiverFromTrack(videoTrack,
	// 	webrtc.RtpTransceiverInit{
	// 		Direction: webrtc.RTPTransceiverDirectionSendonly,
	// 	},
	// )
	// // rtpSender, err := peerConnection.AddTrack(videoTrack)
	// if err != nil {
	// 	panic(err)
	// }

	// Read incoming RTCP packets
	// Before these packets are retuned they are processed by interceptors. For things
	// like NACK this needs to be called.

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	// peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
	// 	fmt.Printf("Connection State has changed %s \n", connectionState.String())
	// })

	// // Wait for the offer to be pasted
	// offer := webrtc.SessionDescription{}

	// signal.Decode(signal.MustReadStdin(), &offer)

	// // Set the remote SessionDescription
	// if err = peerConnection.SetRemoteDescription(offer); err != nil {
	// 	panic(err)
	// }

	// // Create answer
	// answer, err := peerConnection.CreateAnswer(nil)
	// if err != nil {
	// 	panic(err)
	// }

	// // Create channel that is blocked until ICE Gathering is complete
	// gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// // Sets the LocalDescription, and starts our UDP listeners
	// if err = peerConnection.SetLocalDescription(answer); err != nil {
	// 	panic(err)
	// }

	// // Block until ICE Gathering is complete, disabling trickle ICE
	// // we do this because we only can exchange one signaling message
	// // in a production application you should exchange ICE Candidates via OnICECandidate
	// <-gatherComplete

	// // Output the answer in base64 so we can paste it in browser
	// fmt.Println(signal.Encode(*peerConnection.LocalDescription()))

	// videoBuilder = samplebuilder.New(10, &codecs.H264Packet{}, 90000)

	// Read RTP packets forever and send them to the WebRTC Client
	for {

		n, _, err := listener.ReadFrom(inboundRTPPacket)

		if err != nil {
			fmt.Printf("error during read: %s", err)
			panic(err)
		}

		packet := &rtp.Packet{}
		if err = packet.Unmarshal(inboundRTPPacket[:n]); err != nil {
			panic(err)
		}

		// videoBuilder.Push(packet)
		// for {
		// 	sample := videoBuilder.Pop()
		// 	if sample == nil {
		// 		break
		// 	}
		// 	nal := signal.NewNal(sample.Data)
		// 	nal.ParseHeader()
		// 	fmt.Printf("NAL Unit Type: %s\n", nal.UnitType.String())

		// }

		if _, writeErr := videoTrack.Write(inboundRTPPacket[:n]); writeErr != nil {
			panic(writeErr)
		}
	}

}

func cleanConnections() {
	listLock.Lock()
	defer func() {
		listLock.Unlock()
	}()
	attemptClean := func() (tryAgain bool) {
		for i := range peerConnections {
			if peerConnections[i].peerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
				peerConnections = append(peerConnections[:i], peerConnections[i+1:]...)
				return true // We modified the slice, start from the beginning
			}
		}
		return
	}

	for cleanAttempt := 0; ; cleanAttempt++ {
		if cleanAttempt == 25 {
			// Release the lock and attempt a sync in 3 seconds. We might be blocking a RemoveTrack or AddTrack
			go func() {
				time.Sleep(time.Second * 3)
				cleanConnections()
			}()
			return
		}

		if !attemptClean() {
			break
		}
	}

}

// Handle incoming websockets
func websocketHandler(w http.ResponseWriter, r *http.Request) {

	
	// Upgrade HTTP request to Websocket
	unsafeConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	c := &threadSafeWriter{unsafeConn, sync.Mutex{}}

	// When this frame returns close the Websocket
	defer c.Close() //nolint

	// Create new PeerConnection
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Print(err)
		return
	}

	// When this frame returns close the PeerConnection
	defer peerConnection.Close() //nolint

	// Accept one audio and one video track Outgoing
	transceiver, err := peerConnection.AddTransceiverFromTrack(videoTrack,
		webrtc.RtpTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionSendonly,
		},
	)
	if err != nil {
		log.Print(err)
		return
	}
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := transceiver.Sender().Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()
	// for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
	// 	if _, err := peerConnection.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
	// 		Direction: webrtc.RTPTransceiverDirectionSendonly,
	// 	}); err != nil {
	// 		log.Print(err)
	// 		return
	// 	}
	// }

	// Add our new PeerConnection to global list
	listLock.Lock()
	peerConnections = append(peerConnections, peerConnectionState{peerConnection, c})
	fmt.Printf("Connections: %d\n", len(peerConnections))
	listLock.Unlock()

	// Trickle ICE. Emit server candidate to client
	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}

		candidateString, err := json.Marshal(i.ToJSON())
		if err != nil {
			log.Println(err)
			return
		}

		if writeErr := c.WriteJSON(&websocketMessage{
			Event: "candidate",
			Data:  string(candidateString),
		}); writeErr != nil {
			log.Println(writeErr)
		}
	})

	// If PeerConnection is closed remove it from global list
	peerConnection.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		switch p {
		case webrtc.PeerConnectionStateFailed:
			if err := peerConnection.Close(); err != nil {
				log.Print(err)
			}
		case webrtc.PeerConnectionStateClosed:
			cleanConnections()
			
		}
	})

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		log.Print(err)
	}

	if err = peerConnection.SetLocalDescription(offer); err != nil {
		log.Print(err)
	}

	offerString, err := json.Marshal(offer)
	if err != nil {
		log.Print(err)
	}

	if err = c.WriteJSON(&websocketMessage{
		Event: "offer",
		Data:  string(offerString),
	}); err != nil {
		log.Print(err)
	}

	message := &websocketMessage{}
	for {
		_, raw, err := c.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		} else if err := json.Unmarshal(raw, &message); err != nil {
			log.Println(err)
			return
		}

		switch message.Event {
		case "candidate":
		
			candidate := webrtc.ICECandidateInit{}
			if err := json.Unmarshal([]byte(message.Data), &candidate); err != nil {
				log.Println(err)
				return
			}

			if err := peerConnection.AddICECandidate(candidate); err != nil {
				log.Println(err)
				return
			}
		case "answer":
		
			answer := webrtc.SessionDescription{}
			if err := json.Unmarshal([]byte(message.Data), &answer); err != nil {
				log.Println(err)
				return
			}

			if err := peerConnection.SetRemoteDescription(answer); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

type threadSafeWriter struct {
	*websocket.Conn
	sync.Mutex
}

func (t *threadSafeWriter) WriteJSON(v interface{}) error {
	t.Lock()
	defer t.Unlock()

	return t.Conn.WriteJSON(v)
}
