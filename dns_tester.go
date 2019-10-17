package main

import (
	"fmt"
	"sync"

	"github.com/miekg/dns"
)

type outputHandler struct {
	lock          sync.Mutex
	displayOutput bool
	testInfo      []*outputInfo
	totalRequests int
}

type outputInfo struct {
	transport     string
	displayOffset int
	totalRequests int
	numSuccess    int
	numFailed     int
}

func newOutputHandler(totalRequests int) outputHandler {
	return outputHandler{
		lock:          sync.Mutex{},
		displayOutput: true,
		testInfo:      make([]*outputInfo, 0),
		totalRequests: totalRequests,
	}
}

func (opt *outputHandler) display(success bool, transport string) {
	if !opt.displayOutput {
		return
	}

	var (
		outInfo *outputInfo
		found   bool
	)

	for _, val := range opt.testInfo {
		if val.transport == transport {
			outInfo = val
			found = true
		}
	}

	if !found {
		opt.lock.Lock() // Lock for the modification we are making to the array
		var newOutInfo = &outputInfo{
			transport:     transport,
			displayOffset: len(opt.testInfo),
			totalRequests: opt.totalRequests,
			numSuccess:    0,
			numFailed:     0,
		}
		fmt.Println() // In this circumstance, we know that the very next thing
		// is going to be output, so create the blank line for that output to exist on
		outInfo = newOutInfo
		opt.testInfo = append(opt.testInfo, newOutInfo)
		opt.lock.Unlock()
	}

	// Move cursor to appropriate line
	// Calculate the proper display offset
	var whereToGo = len(opt.testInfo) - outInfo.displayOffset
	whereToGo = +whereToGo // ensure + sign
	opt.lock.Lock()        // Lock the output so that another thread doesn't try to update
	fmt.Printf("\033[%dA", whereToGo)
	fmt.Printf("\033[2K\r")
	dilog.Printf("Calculated display offset (%s): %d\n", outInfo.transport, whereToGo)

	if success {
		outInfo.numSuccess++
		fmt.Printf("%s \033[38;5;118msuccess\033[m: [%d/%d] \033[38;5;196mfailures\033[m: [%d/%d] \033[38;5;63mtotal\033[m: [%d/%d]",
			transport, outInfo.numSuccess, outInfo.totalRequests,
			outInfo.numFailed, outInfo.totalRequests,
			outInfo.numFailed+outInfo.numSuccess, outInfo.totalRequests)
	} else {
		outInfo.numFailed++
		fmt.Printf("%s \033[38;5;118msuccess\033[m: [%d/%d] \033[38;5;196mfailures\033[m: [%d/%d] \033[38;5;63mtotal\033[m: [%d/%d]",
			transport, outInfo.numSuccess, outInfo.totalRequests,
			outInfo.numFailed, outInfo.totalRequests,
			outInfo.numFailed+outInfo.numSuccess, outInfo.totalRequests)
	}

	// Move cursor back to bottom line
	fmt.Printf("\033[%dB\r", whereToGo)
	opt.lock.Unlock()
}

type tester struct {
	transport       string
	transportClient *dns.Client
	servers         *[]string
	queries         *[]string
}

func newTester(connectionType string, servers, queries *[]string) *tester {
	var tester = tester{}

	tester.transport = connectionType
	tester.transportClient = &dns.Client{Net: connectionType}
	tester.servers = servers
	tester.queries = queries

	return &tester
}

func (tstr *tester) test(outputHandler *outputHandler, exitIndicate chan bool) {
	dilog.Printf("Beginning %s test\n", tstr.transport)
	for _, server := range *tstr.servers {
		for _, query := range *tstr.queries {
			dilog.Printf("Testing %v @ %v (%s)\n", query, server, tstr.transport)
			var msg = dns.Msg{}
			msg.SetQuestion(dns.Fqdn(query), dns.TypeA)
			if rsp, _, err := tstr.transportClient.Exchange(&msg, server); err == nil {
				if rsp.Rcode == 0 {
					dslog.Printf("Successfully resolved %s (%s)\n", query, tstr.transport)
					outputHandler.display(true, tstr.transport)
				} else {
					dslog.Printf("Failed to resolve %s (%s)\n", query, tstr.transport)
					outputHandler.display(false, tstr.transport)
				}
			} else {
				delog.Printf("DNS request failed: %v (%s)\n", err, tstr.transport)
				outputHandler.display(false, tstr.transport)
			}
		}
	}

	dslog.Printf("Finished %s test\n", tstr.transport)
	exitIndicate <- true
}
