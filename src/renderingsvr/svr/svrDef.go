package svr

type RTaskJsonNode struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	ResUrl     string `json:"resUrl"`
	Resolution [2]int `json:"resolution"`
	Phase      string `json:"phase"`
	Action     string `json:"action"`
}
type RTasksJson struct {
	Tasks []RTaskJsonNode `json:"tasks"`
}
type RTaskJson struct {
	Phase  string        `json:"phase"`
	Task   RTaskJsonNode `json:"task"`
	Status int           `json:"status"`
}
