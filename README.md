# Go-Wild-DNS

Micro DNS-server that implements wildcard-ip DNS magic functionality: 

```
        10.0.0.1.ip.mydomain.io  resolves to   10.0.0.1
    www.10.0.0.1.ip.mydomain.io  resolves to   10.0.0.1
 mysite.10.0.0.1.ip.mydomain.io  resolves to   10.0.0.1
foo.bar.10.0.0.1.ip.mydomain.io  resolves to   10.0.0.1
```

It's like your own [xip.io](http://xip.io)


## Setup 

Create NS record for magic-domain: 
```
ip.mydomain.io 	NS 	ns1.ip.mydomain.io
```
and A record too: 
```
ns1.ip.mydomain.io  A 1.2.3.4
```

Now go to the 1.2.3.4 server: 

0. Setup go compiler, build go-wild-dns binary, allow firewall access to port 53
1. Edit go-wild-dns.service and put ```ip.mydomain.io.``` (note dot at the end) to DOMAIN_SUFFIX environment variable. Edit path to binary if necessary. 
2. Copy go-wild-dns.service to /etc/systemd/system/
3. systemctl daemon-reload && systemctl enable go-wild-dns && systemctl start go-wild-dns 


Now repeat everything for secondary server. 


## Notes 

- It seems to be not so much DNS-compliant (no NS or SOA records) for now 
- No tests (just works on the resolvers I've tested)
- Most certanly not production grade, runs as root 
