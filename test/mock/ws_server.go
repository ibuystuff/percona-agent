package mock

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"net/http"
)

type WebsocketServer struct {
}

var SendChan chan interface{}
var RecvChan chan interface{}

// addr: http://127.0.0.1:8000
// endpoint: /agent
func (s *WebsocketServer) Run(addr string, endpoint string) {
	go run()
	http.Handle(endpoint, websocket.Handler(wsHandler))
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

type client struct {
	ws       *websocket.Conn
	origin   string
	SendChan chan interface{} // data to client
	RecvChan chan interface{} // data from client
}

func wsHandler(ws *websocket.Conn) {
	c := &client{
		ws:       ws,
		origin:   ws.Config().Origin.String(),
		SendChan: make(chan interface{}, 5),
		RecvChan: make(chan interface{}, 5),
	}
	ClientConnectChan <- c
	defer func() { ClientDisconnectChan <- c }()
	go c.send()
	c.recv()
}

func (c *client) recv() {
	defer c.ws.Close()
	for {
		var data interface{}
		err := websocket.JSON.Receive(c.ws, &data)
		if err != nil {
			break
		}
		// log.Printf("recv: %+v\n", data)
		c.RecvChan <- data
	}
}

func (c *client) send() {
	defer c.ws.Close()
	for data := range c.SendChan {
		// log.Printf("recv: %+v\n", data)
		err := websocket.JSON.Send(c.ws, data)
		if err != nil {
			break
		}
	}
}

var ClientConnectChan = make(chan *client)
var ClientDisconnectChan = make(chan *client)
var Clients = make(map[string]*client)

func run() {
	for {
		select {
		case c := <-ClientConnectChan:
			// todo: this is probably prone to deadlocks, not thread-safe
			Clients[c.origin] = c
			// log.Printf("connect: %+v\n", c)
		case c := <-ClientDisconnectChan:
			c, ok := Clients[c.origin]
			if ok {
				close(c.SendChan)
				c.ws.Close()
				//log.Printf("disconnect: %+v\n", c)
				delete(Clients, c.origin)
			}
		}
	}
}