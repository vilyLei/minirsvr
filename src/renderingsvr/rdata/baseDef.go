package rdata

import (
	"encoding/json"
	"fmt"
)

// go mod init renderingsvr.com/rdata
type RTMaterialNode struct {
	ModelName      string     `json:"modelName"`
	UVScales       [2]float64 `json:"uvScales"`
	Type           string     `json:"type"`
	Color          uint       `json:"color"`
	Specular       float64    `json:"specular"`
	Metallic       float64    `json:"metallic"`
	Roughness      float64    `json:"roughness"`
	NormalStrength float64    `json:"normalStrength"`
	Act            string     `json:"act"` // its values: add, del, update
}
type RTEnvNode struct {
	Path       string  `json:"path"`
	Type       string  `json:"type"`
	Brightness float64 `json:"brightness"`
	AO         float64 `json:"ao"`
	Rotation   float64 `json:"rotation"`
}
type RTOutputNode struct {
	Path          string `json:"path"`
	OutputType    string `json:"outputType"`
	Resolution    [2]int `json:"resolution"`
	BGTransparent int    `json:"bgTransparent"`
	BGColor       uint   `json:"bgColor"`
}
type RTRCameraNode struct {
	Type      string      `json:"type"`
	ViewAngle float64     `json:"viewAngle"`
	Near      float64     `json:"near"`
	Far       float64     `json:"far"`
	Matrix    [16]float64 `json:"matrix"`
}
type RTRenderingNode struct {
	Name    string `json:"name"`
	Unit    string `json:"unit"`
	Version int64  `json:"version"`

	Camera    RTRCameraNode    `json:"camera"`
	Output    RTOutputNode     `json:"output"`
	Env       RTEnvNode        `json:"env"`
	Materials []RTMaterialNode `json:"materials"`
}

func (self *RTRenderingNode) setFromJson(rnodeJsonStr string) {
	err := json.Unmarshal([]byte(rnodeJsonStr), self)
	if err != nil {
		fmt.Printf("RTRenderingNode::setFromJson() Unmarshal failed, err: %v\n", err)
	}
	fmt.Println("RTRenderingNode::setFromJson(), self: ", self)
}
