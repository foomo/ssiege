package benchmark

import (
	"math"
	"time"
)

type LinePoint struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

type Line struct {
	Label  string       `json:"label"`
	Values []*LinePoint `json:"values"`
	Color  string       `json:"color"`
}

type Graph struct {
	Requests []*Line `json:"requests"`
	RPS      []*Line `json:"RPS"`
}

type Analyzer struct {
	Siege *Siege
}

func NewAnalyzer(siege *Siege) *Analyzer {
	return &Analyzer{
		Siege: siege,
	}
}

func (a *Analyzer) Graph() *Graph {
	startUnix, endUnix := a.GetCallsWindow()
	return &Graph{
		Requests: a.lines(startUnix),
		RPS:      a.RPS(startUnix, endUnix),
	}
}

func (a *Analyzer) lines(startUnix int64) []*Line {
	first, _ := a.GetCallsWindow()
	lines := []*Line{}
	for i, r := range a.Siege.Results {
		runner := a.Siege.Runners[i]
		l := &Line{
			Label:  runner.Benchmark.Name,
			Color:  runner.Benchmark.Color,
			Values: []*LinePoint{},
		}
		for _, sessionStats := range r.SessionStats {
			for _, callStats := range sessionStats.Calls {
				milliseconds := callStats.Duration.Nanoseconds() / 1000000
				l.Values = append(l.Values, &LinePoint{
					X: callStats.Start.Unix() - first,
					Y: int64(math.Log10(float64(milliseconds)) * 1000000),
				})
			}
		}
		lines = append(lines, l)
	}
	return lines
}

const avgWindowSize = 5

func (a *Analyzer) RPS(startUnix int64, endUnix int64) []*Line {
	numberOfSeconds := endUnix - startUnix
	lines := []*Line{}
	if numberOfSeconds > 0 {
		for i, r := range a.Siege.Results {
			// next benchmark
			runner := a.Siege.Runners[i]
			linePoints := make([]*LinePoint, numberOfSeconds)
			for i := int64(0); i < numberOfSeconds; i++ {
				linePoints[i] = &LinePoint{
					X: i,
					Y: -1,
				}
			}
			//log.Println("line points ----------------------------------------------------------------", numberOfSeconds, len(linePoints))
			for _, sessionStats := range r.SessionStats {
				// next session
				for _, callStats := range sessionStats.Calls {
					// next call
					x := callStats.Start.Unix() - startUnix
					x = x - x%avgWindowSize
					if x < int64(len(linePoints)) {
						linePoints[x].Y++
						//lp := linePoints[x]
						//lp.Y++
						//log.Println("trying to access", x, lp.Y)
					} else {
						// data out of initial scope benchmark is running ...
						//log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! x not found", x, callStats, "in", len(linePoints))
					}
				}
			}
			lastY := int64(-1)
			for _, linePoint := range linePoints {
				if linePoint.Y > -1 {
					lastY = linePoint.Y / avgWindowSize
				}
				if lastY > -1 && linePoint.Y == -1 {
					linePoint.Y = lastY
				} else {
					linePoint.Y /= avgWindowSize
				}
			}
			lines = append(lines, &Line{
				Label:  runner.Benchmark.Name,
				Color:  runner.Benchmark.Color,
				Values: linePoints,
			})
		}
	}

	return lines
}

func (a *Analyzer) GetCallsWindow() (first int64, last int64) {
	first = time.Now().Add(time.Hour * 365 * 24).Unix()
	last = 0
	for _, r := range a.Siege.Results {
		for _, sessionStats := range r.SessionStats {
			for _, callStats := range sessionStats.Calls {
				unixTime := callStats.Start.Unix()
				if unixTime < first {
					first = unixTime
				}
				if unixTime > last {
					last = unixTime
				}
			}
		}
	}
	return first, last
}
