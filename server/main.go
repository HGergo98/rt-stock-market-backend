package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handler(w http.ResponseWriter, r *http.Request) {
	// 1 - upgrade the incoming GET request to a websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// 2 - read messages from the websocket
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}

		// 3 - print the message to the console
		fmt.Printf("Received: %s\n", message)

		res := fmt.Sprintf("%s, at: %s", string(message), time.Now().String())

		// 4 - write the message back to the client
		if err := conn.WriteMessage(messageType, []byte(res)); err != nil {
			fmt.Println(err)
			return
		}
	}
}

func main() {
	http.HandleFunc("/ws", handler)

	fmt.Println("Starting server on :3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		fmt.Println(err)
		return
	}
}
