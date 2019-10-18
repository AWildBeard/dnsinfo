package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var (
	debug bool
	doTCP bool
	doUDP bool
	doDOT bool

	timeout time.Duration

	dilog *log.Logger
	dslog *log.Logger
	delog *log.Logger
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable debugging message output")
	flag.BoolVar(&doTCP, "tcp", false, "Enable DNS over TCP testing")
	flag.BoolVar(&doUDP, "udp", false, "Enable DNS over UDP testing")
	flag.BoolVar(&doDOT, "dot", false, "Enable DNS over TLS testing")
	flag.DurationVar(&timeout, "timeout", 250*time.Millisecond, "Configure the timeout for all connections")
}

func main() {
	flag.Parse()

	if debug {
		dilog = log.New(os.Stderr, "\033[38;5;63mINFO:\033[m ", 0)
		dslog = log.New(os.Stderr, "\033[38;5;118mSUCC:\033[m ", 0)
		delog = log.New(os.Stderr, "\033[38;5;196mFAIL:\033[m ", 0)
	} else {
		dilog = log.New(ioutil.Discard, "", 0)
		dslog = log.New(ioutil.Discard, "", 0)
		delog = log.New(ioutil.Discard, "", 0)
	}

	defer dslog.Println("Exiting")

	var (
		servers = []string{"1.1.1.1:53", "1.0.0.1:53", "8.8.8.8:53",
			"8.8.4.4:53", "9.9.9.9:53"}
		queries = []string{"www.google.com", "www.youtube.com",
			"www.facebook.com", "www.duckduckgo.com",
			"golang.org", "www.github.com"}
		outputHandler = newOutputHandler()
		exitIndicate  = make(chan bool, 1)
		testsRunning  = 0
	)

	if debug {
		outputHandler.displayOutput = false
	}

	if doTCP {
		var tcpTester = newTester("tcp", &servers, &queries)
		go tcpTester.test(&outputHandler, exitIndicate)
		testsRunning++
	}

	if doUDP {
		var udpTester = newTester("udp", &servers, &queries)
		go udpTester.test(&outputHandler, exitIndicate)
		testsRunning++
	}

	if doDOT {
		var newServers = []string{"1.1.1.1:853", "1.0.0.1:853", "8.8.8.8:853", "8.8.4.4:853"}
		var dotTester = newTester("tcp-tls", &newServers, &queries)
		go dotTester.test(&outputHandler, exitIndicate)
		testsRunning++
	}

	for testsRunning > 0 {
		select {
		case <-exitIndicate:
			testsRunning--
		}
	}
}
