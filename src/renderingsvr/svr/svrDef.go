package svr

import (
	"renderingsvr.com/rdata"
)

type RTaskJsonNode struct {
	Id            int64                 `json:"id"`
	Name          string                `json:"name"`
	ResUrl        string                `json:"resUrl"`
	Resolution    [2]int                `json:"resolution"`
	Camdvs        [16]float64           `json:"camdvs"`
	BGTransparent int                   `json:"bgTransparent"`
	Phase         string                `json:"phase"`
	Action        string                `json:"action"`
	RNode         rdata.RTRenderingNode `json:"rnode"`
}

func (self *RTaskJsonNode) Reset() {
}

type RTasksJson struct {
	Tasks []RTaskJsonNode `json:"tasks"`
}

func (self *RTasksJson) Reset() {
	// self.Tasks = nil
}

type RTaskJson struct {
	Phase  string        `json:"phase"`
	Task   RTaskJsonNode `json:"task"`
	Status int           `json:"status"`
}

func (self *RTaskJson) Reset() {
	// self.RNode = nil
}

var AutoCheckRTask = false
var RootDir = ""
