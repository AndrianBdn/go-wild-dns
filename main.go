package main

import (
	"github.com/miekg/dns"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const ipDiscoveryURL1 string = "http://whatismyip.akamai.com/"
const ipDiscoveryURL2 string = "https://api.ipify.org/"
const ipDiscoveryURL3 string = "https://ifconfig.co/ip"

var staticA map[string]net.IP
var defaultIP net.IP
var domainSuffix string

func ipFromHost(host string, def net.IP) net.IP {
	var sip string

	r, _ := regexp.Compile("(\\d+\\.\\d+\\.\\d+\\.\\d+)\\.")
	submatch := r.FindStringSubmatch(host)
	if len(submatch) > 1 {
		sip = submatch[1]
	} else {

		r, _ = regexp.Compile("(\\d+-\\d+-\\d+-\\d+)\\.")
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

func getMyIPWithService(serviceURL string) net.IP {
	resp, err := http.Get(serviceURL)
	if err != nil {
		log.Printf("HTTP GET error %s", err)
		return nil
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode == 200 {
		respBody, _ := ioutil.ReadAll(resp.Body)
		sip := strings.TrimSpace(string(respBody))
		ip := net.ParseIP(sip)
		if ip == nil {
			log.Printf("fail, %s returned bad IP\n", sip)
			return nil
		}
		return ip.To4()
	}

	log.Fatalf("bad response: %s", resp.Status)
	return nil
}

func getMyIP() net.IP {
	ip := getMyIPWithService(ipDiscoveryURL1)
	if ip != nil {
		return ip
	}

	ip = getMyIPWithService(ipDiscoveryURL2)
	if ip != nil {
		return ip
	}

	ip = getMyIPWithService(ipDiscoveryURL3)

	return ip
}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	if r.Opcode == dns.OpcodeQuery {
		for _, q := range m.Question {
			var r dns.RR
			if q.Qtype == dns.TypeTXT {
				r = handleTxtRequest(q)
			} else {
				// default - will reply with A request
				r = handleARequest(q)
			}
			if isNil(r) {
				return
			}

			m.Answer = append(m.Answer, r)
		}
	}

	_ = w.WriteMsg(m)
}
func handleARequest(q dns.Question) *dns.A {
	qNameLower := strings.ToLower(q.Name)
	var ip net.IP

	if val, set := staticA[qNameLower]; set {
		ip = val
	} else {
		ip = ipFromHost(q.Name, defaultIP)
		if strings.HasSuffix(qNameLower, domainSuffix) {
			ip = defaultIP
		} else {
			return nil 
		}
	}

	aRec := dns.A{
		Hdr: dns.RR_Header{
			Name:   q.Name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    86400,
		},
		A: ip,
	}
	log.Printf("resolving %v to %v", q.Name, ip)
	return &aRec
}

func handleTxtRequest(q dns.Question) *dns.TXT {
	qNameLower := strings.ToLower(q.Name)

	txtPath := os.Getenv("TXT_RECORDS_PATH")
	if txtPath == "" {
		return nil
	}

	if strings.ContainsAny(qNameLower, "/\\") {
		return nil
	}

	if !strings.HasSuffix(qNameLower, domainSuffix) {
		return nil
	}

	recordPath := filepath.Join(txtPath, qNameLower)

	value, err := ioutil.ReadFile(recordPath)
	if err != nil {
		log.Printf("resolving %v: 404", q.Name)
		return nil
	}
	strValue := string(value)
	strValue = strings.TrimSpace(strValue)
	if len(strValue) > 255 {
		log.Printf("ERROR: resolving %v to (value too big, not sending): %v ", q.Name, strValue)
		return nil
	}

	log.Printf("resolving %v to %v", q.Name, strValue)

	return &dns.TXT{
		Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
		Txt: []string{strValue},
	}
}

func discoverIPWithRetries() {

	for t := 0; t <= 5; t++ {
		log.Println("Discovering our IP...")
		defaultIP = getMyIP()

		if defaultIP != nil {
			break
		} else {
			time.Sleep(time.Second * 5)
		}
	}

	if defaultIP == nil {
		log.Fatalf("Was unable to discover our IP")
	}

	log.Println(defaultIP)
}

func discoverDomainSuffix() {
	domainSuffix = os.Getenv("DOMAIN_SUFFIX")
	if domainSuffix == "" {
		log.Fatal("Error: DOMAIN_SUFFIX environment is not set")
	}

	if !strings.HasSuffix(domainSuffix, ".") {
		domainSuffix = domainSuffix + "."
	}
	domainSuffix = strings.ToLower(domainSuffix)
}

func discoverOtherNS() {

	if domainSuffix == "" {
		log.Fatal("Error: DOMAIN_SUFFIX must be set before")
	}

	staticA = make(map[string]net.IP)

	for i := 1; i <= 4; i++ {
		key := "NS" + strconv.Itoa(i)
		nsval := os.Getenv(key)

		if nsval != "" {
			ip := net.ParseIP(strings.TrimSpace(nsval))
			if ip == nil || ip.To4() == nil {
				continue
			}

			staticA[strings.ToLower(key)+"."+domainSuffix] = ip.To4()
		}
	}
}

func main() {
	discoverDomainSuffix()
	discoverOtherNS()
	log.Printf("Will serve zone %s\n", domainSuffix)
	discoverIPWithRetries()

	log.Printf("Starting DNS server on port 53...\n")
	dns.HandleFunc(".", handleDnsRequest)
	server := &dns.Server{Addr: ":53", Net: "udp"}
	log.Fatal(server.ListenAndServe())
}
