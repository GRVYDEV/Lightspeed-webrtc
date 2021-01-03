// +build !js

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"flag"

	"github.com/GRVYDEV/lightspeed-webrtc/internal/signal"
	"github.com/gorilla/websocket"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/samplebuilder"
)

var (
	videoBuilder *samplebuilder.SampleBuilder
	addr         = flag.String("addr", "localhost:8080", "http service address")
	iAddr        = flag.String("i-addr", "127.0.0.1", "Address that ingest server should listen on")
	upgrader     = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	videoTrack *webrtc.TrackLocalStaticSample

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
	fmt.Printf("WS Hosted on: %#v\n", *addr)
	log.SetFlags(0)
	trackLocals = map[string]*webrtc.TrackLocalStaticRTP{}

	port := 65535

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

	// Unmarshal the incoming packet
	packet := &rtp.Packet{}
	if err = packet.Unmarshal(inboundRTPPacket[:n]); err != nil {
		panic(err)
	}

	// Create a video track
	videoTrack, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/h264"}, "video", "pion")
	if err != nil {
		panic(err)
	}

	// start HTTP server
	go func() {
		http.HandleFunc("/websocket", websocketHandler)

		log.Fatal(http.ListenAndServe(*addr, nil))
	}()

	videoBuilder = samplebuilder.New(10, &codecs.H264Packet{}, 90000)

	// Read RTP packets forever and send them to the WebRTC Client
	spsAndPpsCache := []byte{}
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
		videoBuilder.Push(packet)

		sample := videoBuilder.Pop()
		if sample == nil {
			break
		}
		nal := signal.NewNal(sample.Data)
		nal.ParseHeader()
		fmt.Printf("NAL Unit Type: %s\n", nal.UnitType.String())

		nal.Data = append([]byte{0x00, 0x00, 0x00, 0x01}, nal.Data...)

		if nal.UnitType == signal.NalUnitTypeSPS || nal.UnitType == signal.NalUnitTypePPS {
			spsAndPpsCache = append(spsAndPpsCache, nal.Data...)
			continue
		} else if nal.UnitType == signal.NalUnitTypeCodedSliceIdr {
			nal.Data = append(spsAndPpsCache, nal.Data...)
			spsAndPpsCache = []byte{}
		}

		// packet := &rtp.Packet{}
		// if err = packet.Unmarshal(inboundRTPPacket[:n]); err != nil {
		// 	panic(err)
		// }

		if writeErr := videoTrack.WriteSample(media.Sample{Data: nal.Data}); writeErr != nil {
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
