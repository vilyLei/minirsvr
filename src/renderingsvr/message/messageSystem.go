package message

import (
	"fmt"

	"renderingsvr.com/rdata"
)

// go mod init renderingsvr.com/message

type RenderingSTChannelData struct {
	PathDir    string
	TaskName   string
	TaskAction string
	ResUrl     string
	RootDir    string
	StType     int
	Flag       int
	TaskID     int64

	RNode rdata.RTRenderingNode `json:"rnode"`
}

func (self *RenderingSTChannelData) Reset() {

	self.TaskID = 0
	self.TaskName = ""
	self.ResUrl = ""
	self.StType = 0
	self.Flag = 0
}

var STRenderingCh chan RenderingSTChannelData

func Init() {
	fmt.Println("message module Init() ...")
	STRenderingCh = make(chan RenderingSTChannelData, 8)
}
