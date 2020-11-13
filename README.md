# Go-Wild-DNS

Micro DNS-server that implements wildcard-ip DNS magic functionality: 

```
        10.0.0.1.ip.mydomain.io  resolves to   10.0.0.1
    www.10.0.0.1.ip.mydomain.io  resolves to   10.0.0.1
 mysite.10.0.0.1.ip.mydomain.io  resolves to   10.0.0.1
foo.bar.10.0.0.1.ip.mydomain.io  resolves to   10.0.0.1
```

It's like your own [xip.io](http://xip.io)

## TLS Mode

Optionally **go-wild-dns** supports *TLS Mode* for wildcard certificates (which can only wildcard up to first dot): 
 
```
        10-0-0-1.ip.mydomain.io  resolves to   10.0.0.1
    www-10-0-0-1.ip.mydomain.io  resolves to   10.0.0.1
 mysite-10-0-0-1.ip.mydomain.io  resolves to   10.0.0.1
foo-bar-10-0-0-1.ip.mydomain.io  resolves to   10.0.0.1
```

## TXT records supports (for letencrypt)

See information about TXT_RECORDS_PATH in go-wild-dns.service 

## Setup 

Create NS record for magic-domain: 
```
ip.mydomain.io 	NS 	ns1.ip.mydomain.io
```
and A record too: 
```
ns1.ip.mydomain.io  A 1.2.3.4
```

Now go to the 1.2.3.4 server (assuming you're using systemd-based Linux): 

0. Setup go compiler, build go-wild-dns binary (just run `go build`), copy binary to /opt/bin/go-wild-dns 
1. Edit go-wild-dns.service: specify your DOMAIN_SUFFIX environment variable. Specify IP of other NS servers if you are using > 1.
2. Copy go-wild-dns.service to /etc/systemd/system/
3. systemctl daemon-reload && systemctl enable go-wild-dns && systemctl start go-wild-dns 
4. Allow firewall access to port 53 (UDP)

Now repeat everything for secondary server (optional)


## DNS Pecularities 

- No NS / SOA records implemented
- Returns its own IP as a fallback 
