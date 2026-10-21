package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
)

type Msg struct {
	From string
	Body string
	Kind string
}

type ChatService struct {
	mu      sync.Mutex
	clients map[string][]Msg
}

func (s *ChatService) broadcast(msg Msg, sender string) {
	for user := range s.clients {
		if user != sender {
			s.clients[user] = append(s.clients[user], msg)
		}
	}
}

func (s *ChatService) Join(username *string, reply *bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[*username]; exists {
		*reply = false
		return nil
	}
	s.clients[*username] = []Msg{}
	*reply = true

	s.broadcast(Msg{
		From: *username,
		Kind: "join",
		Body: fmt.Sprintf("User %s joined the chat.", *username),
	}, *username)

	fmt.Printf("User connected: %s\n", *username)
	return nil
}

func (s *ChatService) Leave(username *string, reply *bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clients[*username]; exists {
		delete(s.clients, *username)
		s.broadcast(Msg{
			From: *username,
			Kind: "leave",
			Body: fmt.Sprintf("User %s left the chat.", *username),
		}, "")
		fmt.Printf("User disconnected: %s\n", *username)
	}
	*reply = true
	return nil
}

func (s *ChatService) SendMessage(msg *Msg, reply *bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.broadcast(*msg, msg.From)
	*reply = true
	return nil
}

func (s *ChatService) GetMessages(username *string, reply *[]Msg) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if msgs, exists := s.clients[*username]; exists {
		*reply = append([]Msg{}, msgs...)
		s.clients[*username] = []Msg{}
	}
	return nil
}

func main() {
	chat := &ChatService{
		clients: make(map[string][]Msg),
	}
	rpc.Register(chat)

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Listener error:", err)
	}
	fmt.Println("RPC Server is running on port 1234...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(conn)
	}
}
