package ws

const (
	MessageTypeAnswer    = "answer"
	MessageTypeCandidate = "candidate"
	MessageTypeOffer     = "offer"
	MessageTypeInfo      = "info"
)

type WebsocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}
