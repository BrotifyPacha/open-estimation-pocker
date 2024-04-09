package domain

import "encoding/json"

type EventType int

const (
	EventTypeUserInit EventType = iota
	EventTypeUserJoined
	EventTypeUserLeft
	EventTypeRoomState
	EventTypeTaskListChanged
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

func TaskListDataFromInterface(i any) (*TaskListData, bool) {
	eventDataI, ok := i.(map[string]any)
	if !ok {
		return nil, false
	}
	tasksI, ok := eventDataI["tasks"].([]map[string]any)
	if !ok {
		return nil, false
	}

	result := TaskListData{
		Tasks: make([]EstimationTask, len(tasksI)),
	}

	for _, taskI := range tasksI {
		result.Tasks = append(result.Tasks, EstimationTask{
			Url: taskI["url"].(string),
		})
	}

	return nil, false
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
