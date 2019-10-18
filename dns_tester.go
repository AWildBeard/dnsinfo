package main

import (
	"fmt"
	"math"
	"sync"

	"github.com/miekg/dns"
)

type outputHandler struct {
	lock          sync.Mutex
	displayOutput bool
	testInfo      []*outputInfo
}

type outputInfo struct {
	transport     string
	displayOffset int
	numSuccess    int
	numFailed     int
}

func newOutputHandler() outputHandler {
	return outputHandler{
		lock:          sync.Mutex{},
		displayOutput: true,
		testInfo:      make([]*outputInfo, 0),
	}
}

func (opt *outputHandler) display(success bool, transport string, totalQueries int) {
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
			numSuccess:    0,
			numFailed:     0,
		}
		fmt.Println() // In this circumstance, we know that the very next thing
		// is going to be output, so create the blank line for that output to exist on
		outInfo = newOutInfo
		opt.testInfo = append(opt.testInfo, newOutInfo)
		opt.lock.Unlock()
	}

	if success {
		outInfo.numSuccess++
	} else {
		outInfo.numFailed++
	}

	// Move cursor to appropriate line
	// Calculate the proper display offset
	var whereToGo = len(opt.testInfo) - outInfo.displayOffset
	whereToGo = +whereToGo // ensure + sign
	opt.lock.Lock()        // Lock the output so that another thread doesn't try to update
	fmt.Printf("\033[%dA", whereToGo)
	fmt.Printf("\033[2K\r")
	dilog.Printf("Calculated display offset (%s): %d\n", outInfo.transport, whereToGo)

	var ratio = float64(25) / float64(totalQueries)
	var percentage = 4.0
	var progress = float64(outInfo.numSuccess+outInfo.numFailed) * ratio
	progress = math.Ceil(progress)

	fmt.Printf("%-8s [", transport)

	for i := float64(0); i < progress; i++ {
		fmt.Print("■")
	}

	for i := progress; i < float64(25); i++ {
		fmt.Print(" ")
	}

	fmt.Printf("] %3.0f%% \033[38;5;118mS\033[m:%3d \033[38;5;196mF\033[m: %d", progress*percentage,
		outInfo.numSuccess, outInfo.numFailed)

	//	fmt.Printf("%-10s \033[38;5;118msuccess\033[m: [%d/%d] \033[38;5;196mfailures\033[m: [%d/%d] \033[38;5;63mtotal\033[m: [%d/%d]",
	//		transport, outInfo.numSuccess, totalQueries,
	//		outInfo.numFailed, totalQueries,
	//		outInfo.numFailed+outInfo.numSuccess, totalQueries)
	//
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
	tester.transportClient.Timeout = timeout // global flag from cmd line
	tester.servers = servers
	tester.queries = queries

	return &tester
}

func (tstr *tester) test(outputHandler *outputHandler, exitIndicate chan bool) {
	dilog.Printf("Beginning %s test\n", tstr.transport)

	var totalRequests = len(*tstr.queries) * len(*tstr.servers)

	for _, server := range *tstr.servers {
		for _, query := range *tstr.queries {

			dilog.Printf("Testing %v @ %v (%s)\n", query, server, tstr.transport)

			var msg = dns.Msg{}
			msg.SetQuestion(dns.Fqdn(query), dns.TypeA)

			if rsp, _, err := tstr.transportClient.Exchange(&msg, server); err == nil {
				if rsp.Rcode == 0 {
					dslog.Printf("Successfully resolved %s (%s)\n", query, tstr.transport)
					outputHandler.display(true, tstr.transport, totalRequests)

				} else {
					dslog.Printf("Failed to resolve %s (%s)\n", query, tstr.transport)
					outputHandler.display(false, tstr.transport, totalRequests)

				}
			} else {
				delog.Printf("DNS request failed: %v (%s)\n", err, tstr.transport)
				outputHandler.display(false, tstr.transport, totalRequests)

			}
		}
	}

	dslog.Printf("Finished %s test\n", tstr.transport)
	exitIndicate <- true
}
