package websockets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"sync"
	"time"

	"estimation-poker/internal/domain"

	"golang.org/x/net/websocket"
)

type UserConnection struct {
	*domain.User
	RoomID        string
	LastHeartbeat time.Time
	Connection    *websocket.Conn
}

var (
	HeartBeatInterval       = time.Second * 15
	CheckHeartBeatsInterval = time.Second * 15
	heartbeatLogEnabled     = false
)

type Server struct {
	Rooms       map[string]domain.Room
	Connections map[string]UserConnection
	roomMx      sync.RWMutex
	connMx      sync.RWMutex
	Queue       domain.EventQueue
}

func NewServer(ctx context.Context, queue domain.EventQueue) *Server {

	server := &Server{
		Connections: make(map[string]UserConnection),
		Rooms:       make(map[string]domain.Room),
		roomMx:      sync.RWMutex{},
		connMx:      sync.RWMutex{},
		Queue:       queue,
	}

	go server.runHeartbeatChecker(ctx)

	return server
}

func logHeartBeats(format string, message ...any) {
	if !heartbeatLogEnabled {
		return
	}
	log.Printf("heartbeat: "+format+"\n", message...)
}

func (s *Server) runHeartbeatChecker(ctx context.Context) {
	ticker := time.Tick(CheckHeartBeatsInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			if len(s.Connections) == 0 {
				continue
			}

			logHeartBeats("checker: start")
			toClose := []string{}

			s.connMx.Lock()
			for userID, userConn := range s.Connections {
				if time.Since(userConn.LastHeartbeat) > HeartBeatInterval {
					logHeartBeats("checker: failed roomID = %s userID = %s name %s", userConn.RoomID, userID, userConn.Name)
					toClose = append(toClose, userID)
				}
			}
			s.connMx.Unlock()

			s.connMx.Lock()
			for _, userIDToClose := range toClose {
				userConn := s.Connections[userIDToClose]
				userConn.Connection.Close()
				delete(s.Connections, userIDToClose)

				s.Queue.Publish(domain.UserLeftEvent(userConn.RoomID, *userConn.User))

			}
			s.connMx.Unlock()
			logHeartBeats("checker: finished")
		}

	}
}

func (s *Server) WebsocketHandler(c *websocket.Conn) {
	log.Println("handling websocket url")

	roomID := c.Request().PathValue("room_id")
	username := c.Request().URL.Query().Get("username")

	if username == "" {
		log.Println("username is empty")
		c.Write([]byte("error: username missing"))
		c.WriteClose(400)
		return
	}

	s.roomMx.Lock()
	_, ok := s.Rooms[roomID]
	if !ok {
		log.Println("room didn't exist created one")
		s.Rooms[roomID] = domain.Room{
			ID:               roomID,
			Users:            []domain.User{},
			EstimationTasks:  []domain.EstimationTask{},
			EstimationValues: domain.DefaultEstimationPreset,
		}
	}
	s.roomMx.Unlock()

	s.handleConnections(username, roomID, c)
}

func (s *Server) handleConnections(username string, roomID string, c *websocket.Conn) {

	userID := fmt.Sprint(rand.Int63())
	user := domain.User{ID: userID, Name: username}

	s.connMx.Lock()
	s.Connections[userID] = UserConnection{
		User:          &user,
		RoomID:        roomID,
		LastHeartbeat: time.Now(),
		Connection:    c,
	}
	s.connMx.Unlock()

	err := websocket.JSON.Send(c, domain.UserInitializationEvent(roomID, user))
	if err != nil {
		logErrorf("init: %s", err.Error())
	}
	err = s.Queue.Publish(domain.UserJoinedEvent(roomID, user))
	if err != nil {
		logErrorf("join: %s", err.Error())
	}

	wg := sync.WaitGroup{}

	// handle connetions
	// queue events -> Connections
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			s.connMx.RLock()
			_, ok := s.Connections[userID]
			if !ok {
				s.connMx.RUnlock()
				break
			}
			s.connMx.RUnlock()

			message := ""
			err := websocket.Message.Receive(c, &message)
			if err != nil {
				if errors.Is(err, io.EOF) {

					log.Println("closed connection for", userID)

					s.connMx.Lock()
					delete(s.Connections, userID)
					s.connMx.Unlock()

					s.Queue.Publish(domain.UserLeftEvent(roomID, user))

					break
				}
				logErrorf("receive: %s", err.Error())
			}

			if message == "ping" {
				websocket.Message.Send(c, "pong")

				logHeartBeats("received: roomID = %s userID = %s name = %s", roomID, userID, username)

				s.connMx.Lock()
				conn := s.Connections[userID]
				conn.LastHeartbeat = time.Now()
				s.Connections[userID] = conn
				s.connMx.Unlock()
				continue
			}

			var event domain.Event
			err = json.Unmarshal([]byte(message), &event)
			if err != nil {
				logErrorf("receive: unmarshal: %s", err.Error())
				continue
			}

			// if event.Type == domain.EventTypeTaskListChanged {
			// 	var taskListChangedData domain.TaskListData
			// 	err = json.Unmarshal([]byte(message), &taskListChangedData)
			// 	if err != nil {
			// 		logErrorf("receive: unmarshal: %s", err.Error())
			// 		continue
			// 	}
			// 	event.Data = taskListChangedData.Tasks
			// }

			event.UserID = userID
			event.RoomID = roomID

			log.Println("published")
			s.Queue.Publish(event)

		}
	}()

	// handle connetions
	// connection events -> queue
	wg.Add(1)
	go func() {
		defer wg.Done()
		eventCh, err := s.Queue.Subscribe()
		if err != nil {
			logErrorf("%s", err.Error())
			return
		}
		for event := range eventCh {

			log.Println("received")

			switch event.Type {
			case domain.EventTypeUserJoined:
				{
					eventData, ok := event.Data.(domain.UserData)
					if !ok {
						continue
					}
					s.roomMx.Lock()

					room := s.Rooms[event.RoomID]
					room.Users = append(room.Users, eventData.User)

					if len(s.Rooms[event.RoomID].Users) == 0 {
						room.HostID = eventData.User.ID
					}

					s.Rooms[event.RoomID] = room

					s.roomMx.Unlock()
					go s.Queue.Publish(domain.RoomStateEvent(room))
					continue
				}
			case domain.EventTypeUserLeft:
				{
					eventData, ok := event.Data.(domain.UserData)
					if !ok {
						continue
					}

					s.roomMx.Lock()

					room := s.Rooms[event.RoomID]

					users := []domain.User{}
					for _, user := range room.Users {
						if user.ID != eventData.User.ID {
							users = append(users, user)
						}
					}
					room.Users = users

					hostLeft := room.HostID == eventData.User.ID
					if hostLeft && len(room.Users) > 0 {
						room.HostID = room.Users[0].ID
					}

					s.Rooms[event.RoomID] = room

					s.roomMx.Unlock()
					go s.Queue.Publish(domain.RoomStateEvent(room))
					continue
				}
			case domain.EventTypeTaskListChanged:
				{
					eventDataI, ok := event.Data.(map[string]any)
					fmt.Printf("norm %v\n", eventDataI)
					if !ok {
						continue
					}
					tasksI, ok := eventDataI["tasks"].([]any)
					if !ok {
						fmt.Printf("fail %v", eventDataI)
						continue
					}
					tasks := make([]domain.EstimationTask, 0, len(tasksI))
					for _, task := range tasksI {
						t, ok := task.(map[string]any)
						if !ok {
							continue
						}
						tasks = append(tasks, domain.EstimationTask{Url: t["url"].(string)})
					}

					s.roomMx.Lock()

					room := s.Rooms[event.RoomID]

					if event.UserID != room.HostID ||
						event.RoomID != room.ID {
						s.roomMx.Unlock()
						continue
					}

					room.EstimationTasks = tasks

					if len(room.EstimationTasks) == 0 {
						room.ActiveTask = ""
					} else {
						if room.ActiveTask == "" {
							room.ActiveTask = room.EstimationTasks[0].Url
						}
					}

					s.Rooms[event.RoomID] = room

					s.roomMx.Unlock()
					go s.Queue.Publish(domain.RoomStateEvent(room))
					continue
				}

			case domain.EventTypeUserPickedEstimationValue:
				{
					eventDataI, ok := event.Data.(map[string]any)
					if !ok {
						continue
					}
					estivationValue, ok := eventDataI["estimation-value"].(float64)
					if !ok {
						continue
					}

					s.roomMx.Lock()

					room := s.Rooms[event.RoomID]

					if room.ActiveTask == "" {
						s.roomMx.Unlock()
						continue
					}

					for i, task := range room.EstimationTasks {
						if task.Url != room.ActiveTask {
							continue
						}

						if task.Estimations == nil {
							task.Estimations = make(map[string]float64)
						}

						task.Estimations[event.UserID] = estivationValue
						room.EstimationTasks[i] = task
					}

					s.Rooms[event.RoomID] = room
					s.roomMx.Unlock()

					go s.Queue.Publish(domain.RoomStateEvent(room))

				}
			case domain.EventTypeRevealEstimations:
				{
					s.roomMx.Lock()

					room := s.Rooms[event.RoomID]

					if room.ActiveTask == "" {
						s.roomMx.Unlock()
						continue
					}

					for i, task := range room.EstimationTasks {
						if task.Url != room.ActiveTask {
							continue
						}
						room.EstimationTasks[i].EstimationsRevealed = true
					}

					s.Rooms[event.RoomID] = room

					s.roomMx.Unlock()

					go s.Queue.Publish(domain.RoomStateEvent(room))
				}
			case domain.EventTypeSelectedTaskChanged:
				{
					s.roomMx.Lock()

					eventDataI, ok := event.Data.(map[string]any)
					if !ok {
						s.roomMx.Unlock()
						continue
					}

					newActiveTask, ok := eventDataI["new-task"].(string)
					if !ok {
						s.roomMx.Unlock()
						continue
					}

					room := s.Rooms[event.RoomID]
					room.ActiveTask = newActiveTask
					s.Rooms[event.RoomID] = room

					s.roomMx.Unlock()

					go s.Queue.Publish(domain.RoomStateEvent(room))
				}
			}

			s.connMx.Lock()
			s.roomMx.Lock()
			for _, connection := range s.Connections {
				if connection.RoomID != event.RoomID {
					continue
				}

				err := websocket.JSON.Send(connection.Connection, event)
				if err != nil {
					fmt.Println("ws send error:", err.Error())
				}
			}
			s.roomMx.Unlock()
			s.connMx.Unlock()
		}
	}()

	wg.Wait()

	return

}

func logErrorf(format string, args ...string) {
	log.Printf("error: "+format+"\n", args)
}
