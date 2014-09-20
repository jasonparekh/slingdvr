package main

import (
	"encoding/json"
	"flag"
	"os"
	"os/user"
	"strings"
)

var configPath = flag.String("configPath", "~/.slingdvr.json", "path to config")

type Config struct {
	ReceiverId    string   `json:"receiverId"`
	CorrelationId string   `json:"correlationId"`
	Titles        []string `json:"titles"`
	RecordingDir  string   `json:"recordingDir"`
}

var config Config
var rawConfig map[string]interface{}

func ReadConfig() (err error) {
	if (*configPath)[:2] == "~/" {
		usr, _ := user.Current()
		dir := usr.HomeDir
		*configPath = strings.Replace(*configPath, "~", dir, 1)
	}

	file, err := os.Open(*configPath)
	if err != nil {
		return
	}

	// Overwrite existing config
	config = Config{}
	if err = json.NewDecoder(file).Decode(&config); err != nil {
		return
	}

	// Overwrite existing config
	rawConfig = make(map[string]interface{})
	file.Seek(0, 0)
	if err = json.NewDecoder(file).Decode(&rawConfig); err != nil {
		return
	}

	return
}
