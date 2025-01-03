package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type client struct {
	socket *websocket.Conn
	send chan []byte
	room *room
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		c.room.forward <- msg
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.send {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}

type room struct {
	forward chan []byte
	join chan *client
	leave chan *client
	clients map[*client]bool
}

func (r *room) run() {
	for {
		select {
		case client := <- r.join:
			r.clients[client] = true
		
		case client := <- r.leave:
			delete(r.clients, client)
			close(client.send)
		
		case msg := <- r.forward:
			for client := range r.clients {
				client.send <- msg
			}
		}
	}
}

const (
	socketBufferSize = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHttp:", err)
		return
	}
	client := &client {
		socket: socket,
		send: make(chan []byte, messageBufferSize),
		room: r,
	}
	r.join <- client
	defer func() {r.leave <- client}()
	go client.write()
	client.read()
}

func newRoom() *room {
	return &room {
		forward: make(chan []byte),
		join: make(chan *client),
		leave: make(chan *client),
		clients: make(map[*client]bool),
	}
}