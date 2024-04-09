package domain

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var EstimationValuePresets = [][]float64{
	{0.5, 1, 2, 3, 4, 5, 6},
	{1, 2, 3, 4, 5, 6},
}

var DefaultEstimationPreset = EstimationValuePresets[0]

type Room struct {
	ID               string   `json:"id"`
	HostID           string   `json:"host-id"`
	Users            []User   `json:"users"`
	EstimationValues []string `json:"estimation-values"`
	EstimationTasks  []string `json:"estimation-tasks"`
}

type EstimationTask struct {
	Url         string
	Estimations map[string]float64 // userID -> estimation
}

type EventQueue interface {
	Publish(event Event) error
	Subscribe() (<-chan Event, error)
}
