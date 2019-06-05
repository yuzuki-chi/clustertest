package executors

import (
	"bytes"
	"github.com/yuuki0xff/clustertest/models"
	"time"
)

type MergedResult struct {
	results []models.ScriptResult
}

func (mr *MergedResult) Append(result models.ScriptResult) {
	mr.results = append(mr.results, result)
}
func (mr *MergedResult) Merge(results []models.ScriptResult) {
	mr.results = append(mr.results, results...)
}
func (mr *MergedResult) String() string {
	return "<MergedResult>"
}
func (mr *MergedResult) StartTime() time.Time {
	var ts []time.Time
	for _, r := range mr.results {
		ts = append(ts, r.StartTime())
	}
	return minTime(ts)
}
func (mr *MergedResult) EndTime() time.Time {
	var ts []time.Time
	for _, r := range mr.results {
		ts = append(ts, r.EndTime())
	}
	return maxTime(ts)
}
func (mr *MergedResult) Host() string {
	if len(mr.results) == 0 {
		return ""
	}
	return mr.results[0].Host()
}
func (mr *MergedResult) Output() []byte {
	var buf bytes.Buffer
	for i, r := range mr.results {
		if i > 0 {
			buf.WriteString("================================\n")
		}
		buf.Write(r.Output())
	}
	return buf.Bytes()
}
func (mr *MergedResult) ExitCode() int {
	if len(mr.results) == 0 {
		return 0
	}
	last := mr.results[len(mr.results)-1]
	return last.ExitCode()
}

func minTime(ts []time.Time) time.Time {
	minT := time.Time{}
	if len(ts) > 0 {
		minT = ts[0]
		for _, t := range ts {
			if minT.After(t) {
				minT = t
			}
		}
	}
	return minT
}
func maxTime(ts []time.Time) time.Time {
	maxT := time.Time{}
	if len(ts) > 0 {
		maxT = ts[0]
		for _, t := range ts {
			if maxT.Before(t) {
				maxT = t
			}
		}
	}
	return maxT
}
