package penny

import (
	"strconv"
	"strings"
	"time"
)

func init() {
	AddFunctions(&Schedule{})
}

type Schedule struct {}

func (s *Schedule) Name() string { return "schedule" }

func (s *Schedule) Run(slice PageSlice, rawParams string) string {
	var (
		interval int
		start time.Time
	)
	params := strings.Fields(rawParams)	
	for _, p := range params {
		kv := strings.Split(p, "=")
		if kv[0] == "interval" {
			if i, err := strconv.Atoi(kv[1]); err != nil {
				return "interval must be an integer"
			} else {
				interval = i
			}
		}
		if kv[0] == "start" {
			if t, err := time.Parse("2006-01-02", kv[1]); err != nil {
				return "Expected date format YYYY-MM-DD"
			} else {
				start = t
			}
		}
	}
	for i, p := range slice {
		p.Date = start.Add(time.Duration(24*i*interval)*time.Hour)
		p.Save()
	}
	return ""
}
