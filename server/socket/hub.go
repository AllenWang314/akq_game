package socket

import (
	"log"
	"encoding"
	"encoding/json"

	"github.com/AllenWang314/akq-game/socket/packet"
)

// Hub maintains the set of active clients
type Hub struct {
	// Registered clients.
	clients map[string]*Client

	// Inbound messages from the clients.
	broadcast chan *SocketMessage

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func (h *Hub) Init() *Hub {
	h.broadcast = make(chan *SocketMessage)
	h.register = make(chan *Client)
	h.unregister = make(chan *Client)
	h.clients = map[string]*Client{}
	return h
}

// Listens for messages from websocket clients
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			log.Println("A client has joined")
			h.clients[client.id] = client
		case client := <-h.unregister:
			// When a client disconnects, remove them from our clients map
			log.Println("A client has disconnected")
			delete(h.clients, client.id)
			close(client.send)
		case message := <-h.broadcast:
			// Process incoming messages from clients
			log.Println("Received message from client")
			h.processMessage(message)
		}
	}
}

// Sends a message to all of our clients
func (h *Hub) Send(slug string, msg encoding.BinaryMarshaler) {
	data, _ := msg.MarshalBinary()
	h.SendBytes(data)
}

func (h *Hub) SendBytes(msg []byte) {
	log.Println("Sending message to all clients")
	for _ , client := range h.clients {
		client.send <- msg
	}
}

// Processes an incoming message
func (h *Hub) processMessage(m *SocketMessage) {
	res := packet.BasePacket{}

	if err := json.Unmarshal(m.msg, &res); err != nil {
		// TODO: Log to Sentry or something -- this should never happen
		log.Println("ERROR: Received invalid JSON message in processMessage!")
		return
	}

	switch res.Type {
	case "join":
		log.Println("Received join packet")
		res := packet.JoinPacket{}
		json.Unmarshal(m.msg, &res)

		h.Send(res.Slug, res)
	case "round":
		log.Println("Received round packet")
		res := packet.RoundPacket{}
		json.Unmarshal(m.msg, &res)
	
		res.GenerateCard()
		h.Send(res.Slug, res)
	case "turn":
		log.Println("Received turn packet")
		res := packet.TurnPacket{}
		json.Unmarshal(m.msg, &res)

		h.Send(res.Slug, res)
	case "finish":
		log.Println("Received finish packet")
		res := packet.FinishPacket{}
		json.Unmarshal(m.msg, &res)

		h.Send(res.Slug, res)
	}

}