package main

import "time"

type Showing struct {
	Title     string
	Subtitle  string
	Id        string
	Start     time.Time
	End       time.Time
	PVRAttrib string
}
