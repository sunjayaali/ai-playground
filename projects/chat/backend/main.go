package main

import (
	"encoding/json"
	"flag"
	"log"
	"sync"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

// Envelope is the wire format every frame uses.
type Envelope struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	TS    int64  `json:"ts,omitempty"`
	Count int    `json:"count,omitempty"`
	Time  string `json:"time,omitempty"`
	// From echoes the sender's client id, so clients can
	// tell their own messages apart from the rest.
	From string `json:"from,omitempty"`
}

// ClientManager tracks the live connections. Each conn owns a
// write mutex: the hub broadcasts and the per-connection tickers
// both write, so without it those two goroutines could interleave
// frames on the same socket.
type ClientManager struct {
	mu      sync.Mutex
	clients map[*client]struct{}
}

type client struct {
	conn    *websocket.Conn
	writeMu sync.Mutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[*client]struct{}),
	}
}

func (m *ClientManager) Add(c *client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[c] = struct{}{}
}

func (m *ClientManager) Remove(c *client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, c)
}

// Count snapshots the live client count.
func (m *ClientManager) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.clients)
}

// FanOut sends env to every client, including origin. Each write
// is serialized per-connection; a client that fails is removed so
// a dead socket can't stall the fan-out for everyone else.
func (m *ClientManager) FanOut(env Envelope) {
	m.mu.Lock()
	targets := make([]*client, 0, len(m.clients))
	for c := range m.clients {
		targets = append(targets, c)
	}
	m.mu.Unlock()

	for _, c := range targets {
		if err := c.writeJSON(env); err != nil {
			log.Printf("write: %v", err)
			_ = c.conn.Close()
		}
	}
}

// AnnounceCount pushes the hub-wide client count to everyone.
func (m *ClientManager) AnnounceCount() {
	m.FanOut(Envelope{
		Type:  "connected_clients",
		Count: m.Count(),
	})
}

// writeJSON guards the connection's write side. fasthttp/websocket
// does not allow concurrent writers, and BroadcastCount-style
// fan-outs race with the per-client ticker.
func (c *client) writeJSON(env Envelope) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteJSON(env)
}

func main() {
	addr := flag.String("addr", ":3001", "listen address")
	tick := flag.Duration("tick", 0, "server_time interval (0 disables)")
	flag.Parse()

	app := fiber.New()
	clients := NewClientManager()

	// Only upgrade actual websocket handshakes on /ws; anything
	// else gets a 426.
	app.Use("/ws", func(c fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		me := &client{conn: c}

		// ReadMessage blocks until the peer sends a frame or the
		// socket dies; a close frame surfaces as an error here,
		// which is how we learn the client left.
		disconnected := make(chan struct{})
		go func() {
			for {
				mt, p, err := c.ReadMessage()
				if err != nil {
					close(disconnected)
					return
				}
				handleClientFrame(me, p, mt, clients)
			}
		}()

		clients.Add(me)
		clients.AnnounceCount()
		clients.FanOut(Envelope{Type: "system", Text: "a client joined the hub", TS: time.Now().UnixMilli()})
		defer func() {
			clients.Remove(me)
			clients.AnnounceCount()
			clients.FanOut(Envelope{Type: "system", Text: "a client left the hub", TS: time.Now().UnixMilli()})
		}()

		// A nil channel blocks forever in a select, so
		// -tick 0 cleanly disables the server_time feed.
		var tickCh <-chan time.Time
		if *tick > 0 {
			t := time.NewTicker(*tick)
			defer t.Stop()
			tickCh = t.C
		}

		for {
			select {
			case <-disconnected:
				return
			case <-tickCh:
				// Ticker writes go through the same per-conn
				// mutex as fan-outs, so the two never collide.
				if err := me.writeJSON(Envelope{
					Type: "server_time",
					Time: time.Now().UTC().Format(time.RFC3339),
				}); err != nil {
					log.Printf("server_time: %v", err)
					return
				}
			}
		}
	}))

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true, "clients": clients.Count()})
	})

	log.Printf("listening on %s (ws://localhost%s/ws)", *addr, *addr)
	log.Fatal(app.Listen(*addr))
}

// handleClientFrame interprets what a browser sent us.
//
//   - {"type":"ping"}        -> "pong", round-tripped for RTT
//   - {"text":"..."} or
//     {"type":"echo",...}    -> echoed to every client (this is
//     the Send button's path)
//   - non-JSON text          -> plain-text echo
func handleClientFrame(me *client, p []byte, messageType int, clients *ClientManager) {
	// Ignore control frames; only text frames carry protocol
	// traffic from our client.
	if messageType != websocket.TextMessage {
		return
	}

	var msg struct {
		Type string `json:"type"`
		Text string `json:"text"`
		TS   int64  `json:"ts"`
		From string `json:"from"`
	}
	if err := json.Unmarshal(p, &msg); err != nil {
		// Plain text, not our envelope: echo it verbatim so the
		// sender sees it in the feed.
		clients.FanOut(Envelope{Type: "echo", Text: string(p), TS: time.Now().UnixMilli()})
		return
	}

	switch msg.Type {
	case "ping":
		if err := me.writeJSON(Envelope{Type: "pong", TS: time.Now().UnixMilli()}); err != nil {
			log.Printf("pong: %v", err)
		}
	case "echo", "":
		clients.FanOut(Envelope{
			Type: "echo",
			Text: msg.Text,
			TS:   time.Now().UnixMilli(),
			From: msg.From,
		})
	default:
		// Unknown type: pass it through so the room can see it.
		clients.FanOut(Envelope{Type: msg.Type, Text: msg.Text, TS: time.Now().UnixMilli(), From: msg.From})
	}
}
