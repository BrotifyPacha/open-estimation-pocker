package domain

import "encoding/json"

type EventType int

const (
	EventTypeUserInit EventType = iota
	EventTypeUserJoined
	EventTypeUserLeft
	EventTypeRoomState
	EventTypeTaskListChanged
	EventTypeUserPickedEstimationValue
	EventTypeRevealEstimations
	EventTypeSelectedTaskChanged
)

type Event struct {
	RoomID string    `json:"room-id"`
	Type   EventType `json:"type"`
	UserID string    `json:"user-id"`
	Data   any       `json:"data"`
}

func (e *Event) String() string {
	bytes, _ := json.Marshal(e)
	return string(bytes)
}

type UserData struct {
	User User `json:"user"`
}

type TaskListData struct {
	Tasks []EstimationTask `tasks`
}

func UserInitializationEvent(roomID string, user User) Event {
	return Event{
		RoomID: roomID,
		Type:   EventTypeUserInit,
		Data:   UserData{User: user},
	}
}

func UserJoinedEvent(roomID string, user User) Event {
	return Event{
		RoomID: roomID,
		Type:   EventTypeUserJoined,
		Data:   UserData{User: user},
	}
}

func UserLeftEvent(roomID string, user User) Event {
	return Event{
		RoomID: roomID,
		Type:   EventTypeUserLeft,
		Data:   UserData{User: user},
	}
}

func RoomStateEvent(room Room) Event {
	return Event{
		RoomID: room.ID,
		Type:   EventTypeRoomState,
		Data:   room,
	}
}

func RevealEstimations(roomID string) Event {
	return Event{
		RoomID: roomID,
		Type:   EventTypeRevealEstimations,
	}
}
