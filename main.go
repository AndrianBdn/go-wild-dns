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

var defaultIP net.IP
var domainSuffix string


func ipFromHost(host string, def net.IP) net.IP {
	var sip string

	r, _ := regexp.Compile("(\\d+\\.\\d+\\.\\d+\\.\\d+)\\.")
	submatch := r.FindStringSubmatch(host)
	if len(submatch) > 1 {
		sip = submatch[1]
	} else {

		r, _ = regexp.Compile("(\\d+\\-\\d+\\-\\d+\\-\\d+)\\.")
		submatch = r.FindStringSubmatch(host)
		if len(submatch) > 1 {
			daship := submatch[1]
			sip = strings.Replace(daship, "-", ".", 4)
		}
	}

	ip := net.ParseIP(sip)
	if ip == nil {
		return def
	}

	return ip.To4()
}

func getMyIP() net.IP {
	resp, err := http.Get(ipDiscoveryURL)
	if err != nil {
		log.Fatalf("HTTP GET error %s", err)
	}

	if resp.StatusCode == 200 {
		defer resp.Body.Close()
		respBody, _ := ioutil.ReadAll(resp.Body)
		sip := strings.TrimSpace(string(respBody))
		ip := net.ParseIP(sip)
		if  ip == nil {
			log.Fatalf("fail, %s returned bad IP\n")
		}
		return ip.To4()
	}

	log.Fatalf("bad response: %s", resp.Status)
	return nil
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	if r.Opcode == dns.OpcodeQuery {
		for _, q := range m.Question {

			ip := ipFromHost(q.Name, defaultIP)
			if !strings.HasSuffix(q.Name, domainSuffix) {
				ip = defaultIP
			}

			aRec := &dns.A{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    86400,
				},
				A: ip,
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
