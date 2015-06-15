package benchmark

import (
	"fmt"
)

func Print(b *Benchmark, r *BenchmarkResult) {
	hr()
	fmt.Println("ran benchmark", b.Name, "on", b.Session.Server, "concurrency", b.Concurrency)
	hr()
	if len(r.CallSummaries) > 0 {
		fmt.Println("call summaries")
		for _, callSummary := range r.CallSummaries {
			hr()
			fmt.Println("comment             :", callSummary.Call.Comment)
			fmt.Println("uri                 :", callSummary.Call.URI)
			fmt.Println("min                 :", callSummary.MinDuration)
			fmt.Println("max                 :", callSummary.MaxDuration)
			fmt.Println("avg                 :", callSummary.GetAverageDuration())
		}
		hr()
		fmt.Println("summary")
		hr()
	}
	fmt.Println("duration            :", r.Duration)
	fmt.Println("number of requests  :", r.NumRequests)
	fmt.Println("requests per second :", r.GetRequestsPerSecond())
	hr()
	fmt.Println("number of sessions  :", r.NumSessions)
	fmt.Println("pages per session   :", float64(r.NumRequests)/float64(r.NumSessions))
}

func hr() {
	fmt.Println("----------------------------------------------------------------------------------------")
}
