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
	ID               string           `json:"id"`
	HostID           string           `json:"host-id"`
	Users            []User           `json:"users"`
	EstimationValues []float64        `json:"estimation-values"`
	EstimationTasks  []EstimationTask `json:"estimation-tasks"`
	ActiveTask       string           `json:"active-task"`
}

type EstimationTask struct {
	Url                 string             `json:"url"`
	Estimations         map[string]float64 `json:"estimations"` // userID -> estimation
	EstimationsRevealed bool               `json:"estimations-revealed"`
}

type EventQueue interface {
	Publish(event Event) error
	Subscribe() (<-chan Event, error)
}
