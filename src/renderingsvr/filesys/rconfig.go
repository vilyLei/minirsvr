package filesys

import (
	"encoding/json"
	"fmt"
)

/*
sizes := param.Resolution
	fileContent := `{
		"renderer-proc":"` + rendererPath + `",
		"renderer-instance":
			{
				"name":"high-image-renderer",
				"status":"stop"
			},
		"sys": {
			"rootDir":"` + param.RootDir + `"
		},
		"resource":
			{
				"type": "` + param.ResourceType + `",
				"models": ` + param.Models + `
			},
		"task":
			{
				"taskID": ` + strconv.FormatInt(param.TaskID, 10) + `,
				"times": ` + strconv.FormatInt(param.Times, 10) + `,
				"outputPath": "` + param.OutputPath + `",
				"outputResolution": [` + strconv.Itoa(sizes[0]) + `,` + strconv.Itoa(sizes[1]) + `]
			}
		}`

*/

type RenderingConfigParam struct {
	Uuid       string
	TaskID     int64
	Name       string
	OutputPath string
	Times      int64
	Progress   int
	Phase      string

	Resolution    [2]int
	Camdvs        [16]float64
	BGTransparent int

	RootDir string

	ResourceType string
	Models       []string `json:"models"`
}
type RenderingTask struct {
	Uuid     string `json:"uuid"`
	TaskID   int64  `json:"taskID"`
	Name     string `json:"name"`
	Phase    string `json:"phase"`
	Times    int64  `json:"times"`
	Progress int    `json:"progress"`
}
type RenderingIns struct {
	Rendering_ins    string        `json:"rendering-ins"`
	Rendering_task   RenderingTask `json:"rendering-task"`
	Rendering_status string        `json:"rendering-status"`
}

type RTC_Instance struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
type RTC_System struct {
	RootDir string `json:"rootDir"`
}
type RTC_Resource struct {
	Type   string   `json:"type"`
	Models []string `json:"models"`
}
type RTC_Task struct {
	TaskID           int64       `json:"taskID"`
	Times            int64       `json:"times"`
	OutputPath       string      `json:"outputPath"`
	OutputResolution [2]int      `json:"outputResolution"`
	Camdvs           [16]float64 `json:"camdvs"`
	BGTransparent    int
}
type RenderTaskConfig struct {
	RendererProc     string       `json:"renderer-proc"`
	RendererInstance RTC_Instance `json:"renderer-instance"`
	Sys              RTC_System   `json:"sys"`
	Resource         RTC_Resource `json:"resource"`
	Task             RTC_Task     `json:"task"`
}

func (self *RenderTaskConfig) Init() {
}
func (self *RenderTaskConfig) Reset() {

	self.RendererProc = ""
	ins := &self.RendererInstance
	ins.Name = "high-image-renderer"
	ins.Status = "stop"

	sys := self.Sys
	sys.RootDir = ""
}

func (self *RenderTaskConfig) SetValueFromParam(param *RenderingConfigParam) {

	self.Sys.RootDir = param.RootDir
	res := &self.Resource
	res.Type = param.ResourceType
	res.Models = param.Models
	task := &self.Task
	task.TaskID = param.TaskID
	task.Times = param.Times
	task.OutputPath = param.OutputPath
	task.OutputResolution = param.Resolution
	task.Camdvs = param.Camdvs
	task.BGTransparent = param.BGTransparent

}
func (self *RenderTaskConfig) GetJsonString() string {

	jsonBytes, err := json.Marshal(self)
	if err != nil {
		fmt.Println("error:", err)
		return ""
	}
	return string(jsonBytes)
}

type LocalSysConfigRenderer struct {
	MainProc     string `json:"mainProc"`
	RenerderProc string `json:"renererdProc"`
}
type LocalSysConfig struct {
	Renderer LocalSysConfigRenderer `json:"renderer"`
}

func (self *LocalSysConfig) GetRenderCMD(rtaskDir string) string {
	r := &self.Renderer
	cmd := r.MainProc + " -- " + r.RenerderProc + " rtaskDir=" + rtaskDir
	return cmd
}
