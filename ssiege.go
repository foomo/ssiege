package main

import (
	"fmt"
	"github.com/foomo/ssiege/benchmark"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
)

func main() {

	usage := func() {
		fmt.Println("Usage: just call me with a list of benchmark files - see the examples")
	}

	if len(os.Args) < 2 {
		fmt.Println("nothing to do")
		usage()
		os.Exit(1)
	}

	// we do not have flags yet ...
	for _, arg := range os.Args[1:] {
		if arg == "-help" {
			usage()
			return
		}
	}

	// we are very greedy !
	runtime.GOMAXPROCS(runtime.NumCPU())

	siege := benchmark.NewSiege(os.Args[1:])

	// start siege
	go func() {
		siege.Siege()
	}()

	// start web interface
	fmt.Println("starting web interface on 127.0.0.1:9999")
	go func() {
		http.ListenAndServe("127.0.0.1:9999", benchmark.NewService(siege))
	}()

	// look for exit signal
	doneChan := make(chan int)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log.Println("received interrupt signal ...")
			siege.Exit()
			doneChan <- 0
		}
	}()
	log.Println("exiting", <-doneChan)
}
