package main

import (
	"log"
	"os"
	"time"

	"github.com/timberio/tcp_server"
)

type Server struct {
	server          *tcp_server.Server
	ConnectionCount int64
	MessageCount    int64
	File            *os.File
}

func (s *Server) Listen() {
	s.server.Listen()
}

func NewServer(address string, filePath string) *Server {
	internal_server := tcp_server.New(address)

	server := &Server{server: internal_server, ConnectionCount: 0, MessageCount: 0}

	if filePath != "" {
		log.Printf("Ensuring file %v is deleted", filePath)

		os.Remove(filePath)

		log.Printf("Opening file %v", filePath)

		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		server.File = f
	}

	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func(server *Server) {
		for {
			select {
			case <-ticker.C:
				log.Printf("Received %v messages across %v connections", server.MessageCount, server.ConnectionCount)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}(server)

	internal_server.OnNewClient(func(c *tcp_server.Client) {
		log.Print("New connection established")
		server.ConnectionCount++
	})

	internal_server.OnNewMessage(func(c *tcp_server.Client, message string) {
		server.MessageCount++

		if server.File != nil {
			_, err := server.File.WriteString(message)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	internal_server.OnClientConnectionClosed(func(c *tcp_server.Client, err error) {
		log.Print("Connection lost")
	})

	return server
}
