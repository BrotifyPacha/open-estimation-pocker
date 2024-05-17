package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

const bottedRooms = 100
const botsPerRoom = 10

func main() {

	var wg sync.WaitGroup

	wg.Add(bottedRooms * botsPerRoom)

	for i := range bottedRooms {
		roomName := fmt.Sprintf("room-%d", i+1)
		if i == 0 {
			roomName = "ci-cd"
		}
		for j := range botsPerRoom {
			botName := fmt.Sprintf("bot-%d", j+1)
			go func() {
				runBotInRoom(roomName, botName)
				wg.Done()
			}()
		}
	}

	wg.Wait()
}

func runBotInRoom(roomID string, username string) {

	ws, err := openConnection(roomID, username)
	if err != nil {
		log.Println(err.Error())
		return
	}

	var wg sync.WaitGroup

	wg.Add(1)

	go sendHeartbeats(ws)

	go func() {
		var i uint8
		msg := `{ "type": 4, "data": { "tasks": [{"url": "hahahhah"}] }}`
		websocket.Message.Send(ws, msg)
		for {
			i++
			msg := fmt.Sprintf(`{ "type": 5, "data": { "estimation-value": %d } }`, i)
			websocket.Message.Send(ws, msg)
			time.Sleep(time.Second * 1)
		}
	}()

	go func() {
		for {

			var msg string
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				log.Printf("room(%s) bot(%s) error: %s\n", roomID, username, err.Error())
				break
			}

			if msg == "pong" {
				continue
			}

			fmt.Println(msg)

		}
		wg.Done()
	}()

	wg.Wait()
}

func sendHeartbeats(ws *websocket.Conn) {
	for {
		websocket.Message.Send(ws, "ping")
		time.Sleep(2 * time.Second)
	}
}

func openConnection(roomID, username string) (*websocket.Conn, error) {
	url := fmt.Sprintf("ws://localhost:8080/%s/ws?username=%s", roomID, username)
	return websocket.Dial(url, "ws", url)
}
