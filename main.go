package main

import (
	"fmt"

	"github.com/miekg/dns"
)

func main() {
	var message = dns.Msg{}
	message.SetQuestion(dns.Fqdn("www.google.com"), dns.TypeA)

	var in, err = dns.Exchange(&message, "127.0.0.1:53")
	if err == nil {
		fmt.Printf("Good response:\n")
		fmt.Printf("%v\n", in)
	} else {
		fmt.Printf("Bad response: %v\n", err)
		fmt.Printf("%v\n", in)
	}
}
