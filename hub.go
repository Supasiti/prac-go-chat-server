package main

type hub struct {
	clients    map[*client]bool
	nextId     int
	register   chan *client
	unregister chan *client
	notify     chan *chatMessage
}

func NewHub() *hub {

	return &hub{
		clients:    make(map[*client]bool),
		nextId:     1,
		register:   make(chan *client),
		unregister: make(chan *client),
		notify:     make(chan *chatMessage),
	}
}

func (h *hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.nextId += 1
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case chat := <-h.notify:
			for client := range h.clients {
				client.send <- chat
			}
		}
	}
}
