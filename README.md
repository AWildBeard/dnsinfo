# DNSINFO
This tool is just focused around giving you the most common information needed to diagnose DNS issues.
This tool uses miekg/dns as it's core to issue DNS requests (not for system dns or DoH requests tho) 

### Goals
* Perform DNS requests against multiple targets
    * Perform basic DNS requests
    * Perform DNS requests using the systems resolver
    * Perform DNSSEC requests
    * Perform DoT requests
    * Perform DoH requests
* Take the results from the DNS requests and compare them
    * Announce differences in responses. Re-issue requests to determine 
    if a migration has happened during testing
    * Announce requests that never recieved a response (blocked or otherwise)
* Use stylized VT100-escaped output

### General Usage
The program should behave as a one-off program that automatically tests dns in a way that is helpful.
I might add user options later on if I feel like it or they become necessary.

### But WHYYYY
I'm tired of not being able to connect to Panera's wifi without knowing which dns 
security feature I'm using is causing the problem.
