package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"time"
)

var configPath = flag.String("configPath", "~/.slingdvr.json", "path to config")
var recordedPath = flag.String("recordedPath", "~/.slingdvr-recorded.json", "path to recorded")

type Config struct {
	ReceiverId          string    `json:"receiverId"`
	CorrelationId       string    `json:"correlationId"`
	Titles              []string  `json:"titles"`
	EarliestShowingTime time.Time `json:"earliestShowingTime"`
	RecStartTime        time.Time `json:"recStartTime"`
	RecEndTime          time.Time `json:"recEndTime"`
	RecordingDir        string    `json:"recordingDir"`
}

var config Config
var rawConfig map[string]interface{}

func ReadConfig() (err error) {
	*configPath = expandConfigPath(*configPath)

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

func ReadRecorded() ([]string, error) {
	*recordedPath = expandConfigPath(*recordedPath)
	b, err := ioutil.ReadFile(*recordedPath)
	if err != nil {
		return nil, errors.New("Cannot read recorded: " + err.Error())
	}

	var o []string
	if err = json.Unmarshal(b, &o); err != nil {
		return nil, errors.New("Cannot unmarshal recorded: " + err.Error())
	}

	return o, nil
}

func WriteRecorded(r []string) (err error) {
	b, err := json.Marshal(r)
	if err != nil {
		return errors.New("Cannot marshal recorded: " + err.Error())
	}

	return ioutil.WriteFile(*recordedPath, b, 0600)
}

func expandConfigPath(p string) string {
	if p[:2] == "~/" {
		usr, _ := user.Current()
		dir := usr.HomeDir
		return strings.Replace(p, "~", dir, 1)
	}

	return p
}
