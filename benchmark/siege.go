package benchmark

import (
	"fmt"
	"time"
)

type Siege struct {
	Runners           []*BenchmarkRunner
	Results           map[int]*BenchmarkResult
	Benchmark         *Benchmark
	CummulativeResult *BenchmarkResult
	exitChannel       chan int
}

type SiegeResult struct {
	Benchmark *Benchmark
	Result    *BenchmarkResult
	Index     int
}

func NewSiege(benchmarkFiles []string) *Siege {
	s := &Siege{
		exitChannel: make(chan int),
		Runners:     []*BenchmarkRunner{},
		Results:     make(map[int]*BenchmarkResult),
		Benchmark: &Benchmark{
			Name:        "Cummulation of",
			Session:     &Session{},
			Concurrency: 0,
		},
	}

	for _, benchmarkFile := range benchmarkFiles {
		r := NewBenchmarkRunner(benchmarkFile)
		s.Runners = append(s.Runners, r)
		s.Benchmark.Session.Server += ", " + r.Benchmark.Session.Server
		s.Benchmark.Concurrency += r.Benchmark.Concurrency
		s.Benchmark.Name += ", " + r.Benchmark.Name
	}
	return s
}

func (s *Siege) Exit() {
	s.exitChannel <- 0
	s.printStatus(true)
	<-s.exitChannel
}

func (s *Siege) printStatus(full bool) {
	if full {
		for i, r := range s.Results {
			Print(s.Runners[i].Benchmark, r)
		}
	}
	Print(s.Benchmark, s.CummulativeResult)
}
func (s *Siege) Siege() {
	resultChan := make(chan *SiegeResult)
	for i, runner := range s.Runners {
		go runner.Siege(resultChan, i)
	}
	start := time.Now()
	lastPrint := time.Now().Unix()
	exit := false
	for exit == false {
		select {
		case <-s.exitChannel:
			exit = true
			s.printStatus(false)
			s.exitChannel <- 0
		case r := <-resultChan:
			s.Results[r.Index] = r.Result
			go func() {
				s.CummulativeResult = NewBenchmarkResult([]*SessionStats{}, 0)
				for _, r := range s.Results {
					s.CummulativeResult.Add(r)
				}
				s.CummulativeResult.Duration = time.Now().Sub(start)
				if time.Now().Unix()-lastPrint > int64(10) {
					lastPrint = time.Now().Unix()
					hr()
					hr()
					fmt.Println(time.Now(), "new result from", s.Runners[r.Index].Benchmark.Name)
					s.printStatus(false)
				}
			}()
		}
	}
}
