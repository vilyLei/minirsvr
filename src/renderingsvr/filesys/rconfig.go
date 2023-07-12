package filesys

import (
	"encoding/json"
	"fmt"

	"renderingsvr.com/rdata"
)

type RenderingConfigParam struct {
	Uuid       string
	TaskID     int64
	Name       string
	OutputPath string
	Times      int64
	Progress   int
	Phase      string

	RootDir string

	ResourceType string
	Models       []string `json:"models"`

	RNode rdata.RTRenderingNode `json:"rnode"`
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
type RTC_ActiveMesh struct {
	Name string `json:"name"`
}
type RTC_Task struct {
	TaskID     int64  `json:"taskID"`
	Times      int64  `json:"times"`
	OutputPath string `json:"outputPath"`

	RNode rdata.RTRenderingNode `json:"rnode"`
}

// func (self *RTC_Task) setRNodeFromJson(rnodeJsonStr string) {
// 	err := json.Unmarshal([]byte(rnodeJsonStr), self)
// 	if err != nil {
// 		fmt.Printf("RTC_Task::setRNodeFromJson() Unmarshal failed, err: %v\n", err)
// 	}
// 	fmt.Println("RTC_Task::setRNodeFromJson(), self: ", self)
// }

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
	task.RNode = param.RNode

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
	RenerderProc string `json:"rendererProc"`
}
type LocalSysConfigModelToDrc struct {
	MainProc   string `json:"mainProc"`
	ExportProc string `json:"exportProc"`
}
type LocalSysConfig struct {
	Renderer   LocalSysConfigRenderer   `json:"renderer"`
	ModelToDrc LocalSysConfigModelToDrc `json:"modelToDrc"`
}

func (self *LocalSysConfig) GetJsonString() string {

	jsonBytes, err := json.Marshal(self)
	if err != nil {
		fmt.Println("error:", err)
		return ""
	}
	return string(jsonBytes)
}
func (self *LocalSysConfig) GetRenderCMD(rtaskDir string) string {
	r := &self.Renderer
	cmd := r.MainProc + " -- " + r.RenerderProc + " rtaskDir=" + rtaskDir
	return cmd
}
func (self *LocalSysConfig) GetModelExportCMD(modelFilePath string) string {
	m := &self.ModelToDrc
	cmd := m.MainProc + " " + m.ExportProc + " modelFilePath=" + modelFilePath
	return cmd
}

var rendererCmdParam = "renderer=D:/programs/blender/blender.exe"

type SysStartupParam struct {
	PortStr        string
	RsvrType       string
	SvrRootUrl     string
	ProcType       string
	AutoCheckRTask bool
}

func (self *SysStartupParam) Reset() *SysStartupParam {
	self.AutoCheckRTask = true
	self.PortStr = "9092"
	self.ProcType = "local"
	self.RsvrType = "local"
	self.SvrRootUrl = "http://localhost:9091/"
	return self
}

var rcfgFilePath = "static/sys/bpyc/rcfg.json"

func (self *SysStartupParam) SetParam(dataMap map[string]string) *SysStartupParam {

	self.Reset()

	value, hasKey := dataMap["port"]
	if hasKey {
		self.PortStr = value
	}
	value, hasKey = dataMap["proc"]
	if hasKey {
		self.ProcType = value
	}

	value, hasKey = dataMap["auto"]
	if hasKey {
		self.AutoCheckRTask = value == "true"
	}

	value, hasKey = dataMap["rsvr"]
	if hasKey {
		self.RsvrType = value
	}
	switch self.RsvrType {
	case "remote-debug":
		self.SvrRootUrl = "http://www.artvily.com:9093/"
	case "remote-release", "remote":
		self.SvrRootUrl = "http://www.artvily.com/"
	default:
	}

	rendererPath := GetSysConfValueWithName("renderer")
	if rendererPath != "" {
		hasFilePath, _ := PathExists(rendererPath)
		if hasFilePath {
			fmt.Println("dsrdiffusion find the renderer program success !!!")
		} else {
			fmt.Println("dsrdiffusion occurred Error: can't find the renderer program !!!")
		}
		rendererCmdParam = "renderer=" + rendererPath
	}
	fmt.Println("dsrdiffusion rendererCmdParam: ", rendererCmdParam)

	if self.ProcType == "local" {
		rcfgPath := "static/sys/local/config.json"
		GetLocalSysCfg(rcfgPath)
	} else {
		rcfgPath := rcfgFilePath
		hasFilePath, _ := PathExists(rcfgPath)
		if hasFilePath {
			GetLocalSysCfg(rcfgPath)
		} else {
			syncRProcRes(self)
		}
	}
	return self
}
