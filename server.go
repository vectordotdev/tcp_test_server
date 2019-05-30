package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"time"

	"github.com/timberio/tcp_server"
)

type Server struct {
	server          *tcp_server.Server
	ConnectionCount int64  `json:"connection_count"`
	FirstMessage    string `json:"first_message"`
	LastMessage     string `json:"last_message"`
	MessageCount    int64  `json:"message_count"`
	sampleCadence   float64
	sampleMessage   string
}

func (s *Server) Listen() {
	s.server.Listen()
}

func (s *Server) WriteSummary() {
	sBytes, err := json.Marshal(s)
	if err != nil {
		log.Fatal(err)
	}

	filePath := "/tmp/tcp_test_server_summary.json"

	err = ioutil.WriteFile(filePath, sBytes, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Wrote activity summary to %s", filePath)
}

func NewServer(address string) *Server {
	internal_server := tcp_server.New(address)

	server := &Server{server: internal_server, ConnectionCount: 0, MessageCount: 0, sampleCadence: 5000.0}

	internal_server.OnNewClient(func(c *tcp_server.Client) {
		log.Print("New connection established")
		server.ConnectionCount++
	})

	internal_server.OnNewMessage(func(c *tcp_server.Client, message string) {
		server.MessageCount++

		if server.MessageCount == 1 {
			server.FirstMessage = message
		}

		server.LastMessage = message

		if math.Mod(float64(server.MessageCount), server.sampleCadence) == 0 {
			server.sampleMessage = message
		}
	})

	internal_server.OnClientConnectionClosed(func(c *tcp_server.Client, err error) {
		log.Print("Connection lost")
	})

	// Print debug output on an interval. This helps with providing insight
	// into activity without saturating IO.
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Printf("Received %v messages across %v connections", server.MessageCount, server.ConnectionCount)

				if server.sampleMessage != "" {
					log.Printf("Sample: %s", server.sampleMessage)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return server
}
