package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func AsyncRecorder(recs <-chan RecRequest) error {
	for {
		if err := handleRec(<-recs); err != nil {
			return err
		}
	}
}

func handleRec(r RecRequest) error {
	defer close(r.Finished)

	if err := powerReceiver(); err != nil {
		return err
	}

	if err := startProgram(r.Showing); err != nil {
		return err
	}

	record(r.Showing, nil, timeNow().Add(r.Showing.End.Sub(r.Showing.Start)))

	return nil
}

func powerReceiver() error {
	bin := fmt.Sprintf("%s/rec2a.pl", filepath.Dir(os.Args[0]))
	if _, err := os.Stat(bin); os.IsNotExist(err) {
		bin, err = filepath.Abs("rec2a.pl")
		if err != nil {
			return err
		}
	}

	args := getSlingArgs()
	args = append(args, "-selectonly=1")

	// TODO IR blaster is acting up, try a few times
	for i := 0; i < 5; i++ {
		cmd := exec.Command(bin, args...)
		if err := cmd.Start(); err != nil {
			return errors.New("Could not start power-on command: " + err.Error())
		}
		if err := cmd.Wait(); err != nil {
			return errors.New(fmt.Sprintf("Could not wait on power-on command[%s %v]: %s", bin, args, err.Error()))
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func startProgram(s Showing) error {
	u := fmt.Sprintf("http://newdish.sling.com/dvr/rest/v1/receiverid/%s?correlationid=%s", config.ReceiverId, config.CorrelationId)
	//	'User-Agent:da-android-phone/3.6.6' actiontype=playbackStart programid=8052 pvrattribute=16789504 creator=0
	resp, err := http.PostForm(u, url.Values{
		"actiontype":   {"playbackStart"},
		"programid":    {s.Id},
		"pvrattribute": {s.PVRAttrib},
		"creator":      {"0"},
	})

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("Did not get OK response for starting program")
	}

	return nil
}
