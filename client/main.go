package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

// Messages represent the structure of the websocket messages
type Message struct {
	MessageType int
	Data        []byte
}

func main() {
	u := url.URL{Scheme: "ws", Host: "localhost:3000", Path: "/ws"}

	// 1 - connect to the server
	fmt.Printf("connecting to %s\n", u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	// Channels for managing the messages
	send := make(chan Message)
	done := make(chan struct{})

	// Goroutine for reading messages
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			fmt.Printf("Received: %s\n", message)
		}
	}()

	// Goroutine for sending messages
	go func() {
		defer close(done)
		for {
			select {
			case msg := <-send:
				// write the message to the websocket connection
				err := conn.WriteMessage(msg.MessageType, msg.Data)
				if err != nil {
					log.Println("write:", err)
					return
				}
			case <-done:
				return
			}
		}
	}()

	// Read input from the terminal and send it to the websocket server
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter a message to send to the server")
	for scanner.Scan() {
		text := scanner.Text()

		// send the text to the channel
		send <- Message{websocket.TextMessage, []byte(text)}

		if err := scanner.Err(); err != nil {
			log.Println("scanner error:", err)
		}
	}
}
