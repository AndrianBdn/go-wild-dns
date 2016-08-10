package main

import (
	"github.com/miekg/dns"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const ipDiscoveryURL string = "http://whatismyip.akamai.com/"

var defaultIP string
var domainSuffix string

func getMyIP() string {
	resp, err := http.Get(ipDiscoveryURL)
	if err != nil {
		log.Fatalf("HTTP GET error %s", err)
	}

	if resp.StatusCode == 200 {
		defer resp.Body.Close()
		respBody, _ := ioutil.ReadAll(resp.Body)
		ip := strings.TrimSpace(string(respBody))
		if net.ParseIP(ip) == nil {
			log.Fatalf("fail, %s returned bad IP\n")
		}
		return ip
	}

	log.Fatalf("bad response: %s", resp.Status)
	return ""
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	log.Printf("handle")

	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	if r.Opcode == dns.OpcodeQuery {
		for _, q := range m.Question {
			r, _ := regexp.Compile("(\\d+\\.\\d+\\.\\d+\\.\\d+)\\.")
			submatch := r.FindStringSubmatch(q.Name)

			ip := defaultIP
			if len(submatch) > 1 && strings.HasSuffix(q.Name, domainSuffix) {
				ip = submatch[1]
			}

			aRec := &dns.A{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    86400,
				},
				A: net.ParseIP(ip).To4(),
			}
			m.Answer = append(m.Answer, aRec)
		}
	}
	w.WriteMsg(m)
}

func main() {
	domainSuffix = os.Getenv("DOMAIN_SUFFIX")
	if domainSuffix == "" {
		log.Fatal("Error: DOMAIN_SUFFIX environment is not set")
	}
	log.Printf("Will serve zone %s\n", domainSuffix);
	log.Println("Discoverying our IP...")
	defaultIP = getMyIP()
	log.Println(defaultIP)

	log.Printf("Starting DNS server on port 53...\n")
	dns.HandleFunc(".", handleDnsRequest)
	server := &dns.Server{Addr: ":53", Net: "udp"}
	log.Fatal(server.ListenAndServe())
}
