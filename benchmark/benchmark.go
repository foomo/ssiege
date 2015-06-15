package benchmark

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Benchmark struct {
	Concurrency int64
	Name        string
	Color       string
	Session     *Session
}

type TransportWrapper struct {
	Active    bool
	Transport *http.Transport
}

type BenchmarkRunner struct {
	Benchmark *Benchmark
	Runners   []*SessionRunner
	// megalomania
	Active int64
	Stats  []*SessionStats
}

type CallSummary struct {
	MinDuration         time.Duration
	MaxDuration         time.Duration
	CummulativeDuration time.Duration
	Call                *Call
	Count               int64
}

func (c *CallSummary) GetAverageDuration() time.Duration {
	return c.CummulativeDuration / time.Duration(c.Count)
}

func (c *CallSummary) Add(b *CallSummary) {
	if b.MinDuration < c.MinDuration {
		c.MinDuration = b.MinDuration
	}
	if b.MaxDuration > c.MaxDuration {
		c.MaxDuration = b.MaxDuration
	}
	c.Count += b.Count
	c.CummulativeDuration += b.CummulativeDuration
}

type BenchmarkResult struct {
	Duration      time.Duration
	NumRequests   int64
	NumSessions   int64
	SessionStats  []*SessionStats
	CallSummaries []*CallSummary
}

func NewBenchmarkResult(stats []*SessionStats, duration time.Duration) *BenchmarkResult {
	b := &BenchmarkResult{
		SessionStats:  stats,
		Duration:      duration,
		NumSessions:   0,
		NumRequests:   0,
		CallSummaries: []*CallSummary{},
	}
	return b
}

func (a *BenchmarkResult) Add(b *BenchmarkResult) {
	a.Duration += b.Duration
	a.NumRequests += b.NumRequests
	a.NumSessions += b.NumSessions
	a.SessionStats = append(a.SessionStats, b.SessionStats...)
	for i, callSummaryA := range a.CallSummaries {
		callSummaryB := b.CallSummaries[i]
		callSummaryA.Add(callSummaryB)
	}
}

func (r *BenchmarkResult) GetRequestsPerSecond() float64 {
	return float64(r.NumRequests) / float64(r.Duration.Seconds())
}

type SessionRunnerResult struct {
	TransportWrapper *TransportWrapper
	Stats            *SessionStats
}

func NewBenchmarkRunner(benchmarkFile string) *BenchmarkRunner {
	br := &BenchmarkRunner{
		Active: 0,
		Stats:  []*SessionStats{},
	}

	jsonBytes, e := ioutil.ReadFile(benchmarkFile)
	if e != nil {
		panic(errors.New("can not read benchmark file " + e.Error()))
	}
	br.Benchmark = &Benchmark{}
	jsonErr := json.Unmarshal(jsonBytes, &br.Benchmark)
	if jsonErr != nil {
		panic(jsonErr)
	}
	if len(br.Benchmark.Name) == 0 {
		f, err := os.Open(benchmarkFile)
		if err == nil {
			defer f.Close()
			info, err := f.Stat()
			if err == nil {
				br.Benchmark.Name = info.Name()
			}

		}
	}
	return br
}

func (br *BenchmarkRunner) makeCallSummaries() []*CallSummary {
	callSummaries := []*CallSummary{}
	for _, call := range br.Benchmark.Session.Calls {
		callSummaries = append(callSummaries, &CallSummary{
			MinDuration: time.Hour * 1000, // ugly ...
			MaxDuration: 0,
			Call:        call,
			Count:       0,
		})
	}
	return callSummaries
}

func (br *BenchmarkRunner) Run() *BenchmarkResult {
	stats, duration := br.runSession()
	r := NewBenchmarkResult(stats, duration)
	r.CallSummaries = br.makeCallSummaries()
	for _, s := range stats {
		r.NumSessions++
		for ii, callStats := range s.Calls {
			r.NumRequests++
			callSummary := r.CallSummaries[ii]
			callSummary.CummulativeDuration += callStats.Duration
			callSummary.Count++
			if callSummary.MaxDuration < callStats.Duration {
				callSummary.MaxDuration = callStats.Duration
			}
			if callSummary.MinDuration > callStats.Duration {
				callSummary.MinDuration = callStats.Duration
			}
		}
	}
	return r

}

func (br *BenchmarkRunner) Siege(resultChannel chan *SiegeResult, index int) {
	r := NewBenchmarkResult(br.Stats, 0)
	r.CallSummaries = br.makeCallSummaries()
	for {
		br.Active = 0
		br.Stats = []*SessionStats{}
		r.Add(br.Run())
		resultChannel <- &SiegeResult{
			Benchmark: br.Benchmark,
			Result:    r,
			Index:     index,
		}
	}
}

func (br *BenchmarkRunner) runSession() (stats []*SessionStats, duration time.Duration) {
	start := time.Now()
	resultChannel := make(chan *SessionRunnerResult)
	todo := br.Benchmark.Concurrency
	for todo > 0 {
		for br.Active < br.Benchmark.Concurrency {
			t := &TransportWrapper{
				Active:    false,
				Transport: &http.Transport{},
			}
			go br.spawnSession(resultChannel, t)
			br.Active++
		}
		select {
		case result := <-resultChannel:
			br.Stats = append(br.Stats, result.Stats)
			result.TransportWrapper.Active = false
			br.Active--
			todo--
		}
	}
	return br.Stats, time.Now().Sub(start)
}

func (br *BenchmarkRunner) spawnSession(resultChannel chan *SessionRunnerResult, tr *TransportWrapper) {
	sr := NewSessionRunner(br.Benchmark.Session, tr.Transport)
	stats := sr.Run()
	resultChannel <- &SessionRunnerResult{
		Stats:            stats,
		TransportWrapper: tr,
	}
}
