package core

import (
	"encoding/json"
)

type CommandType string

const (
	CMD_Set    CommandType = "SET"
	CMD_Delete CommandType = "DELETE"
)

type CMD struct {
	Type  CommandType
	Key   string
	Value string
}

func NewCMD(b []byte) *CMD {
	c := &CMD{}
	_ = json.Unmarshal(b, c)
	return c
}
