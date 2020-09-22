package handler

import (
	"github.com/dvsekhvalnov/jose2go/base64url"
	D "github.com/liberal-boy/tls-shunt-proxy/handler/dns"
	"github.com/miekg/dns"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
)

type DohServer struct {
	connListener *ConnListener
}

func NewDohServer(args string) *DohServer {
	ln := NewConnListener()
	go func() {
		var client D.DnsClient
		if strings.HasPrefix(args, "https://") {
			client = D.NewDohClient(args)
		} else {
			client = D.NewDo53Client(args)
		}
		doh := dohHandler{
			client: client,
			cache:  D.NewCache(),
		}

		h2s := &http2.Server{}
		err := http.Serve(ln, h2c.NewHandler(&doh, h2s))
		if err != nil {
			log.Fatalln(err)
		}
	}()
	return &DohServer{ln}
}

func (h *DohServer) Handle(conn net.Conn) {
	err := h.connListener.HandleConn(conn)
	if err != nil {
		log.Println(err)
	}
}

type dohHandler struct {
	client D.DnsClient
	cache  *D.Cache
}

func (h *dohHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var dnsByte []byte
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("failed to ReadBody", err)
		return
	}
	if len(body) != 0 {
		dnsByte = body
	} else {
		err := r.ParseForm()
		if err != nil {
			log.Println("failed to ParseForm", err)
			return
		}
		dnsParam := r.Form.Get("dns")
		dnsByte, err = base64url.Decode(dnsParam)
		if err != nil {
			log.Println("failed to Decode", err)
			return
		}
	}
	var msg dns.Msg
	err = msg.Unpack(dnsByte)
	if err != nil {
		log.Println("failed to Unpack", err)
		return
	}

	ip, _, _ := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	ecs := appendEdnsClientSubnet(&msg, net.ParseIP(ip))

	var resp *dns.Msg
	if len(msg.Question) > 0 {
		resp = h.cache.Get(msg.Question[0].Name, msg.Question[0].Qtype, ecs.String())
	}

	if resp == nil {
		resp, err = h.client.Exchange(&msg)
		if err != nil {
			log.Println("failed to Exchange", err)
			return
		}

		if len(msg.Question) > 0 {
			h.cache.Put(msg.Question[0].Name, msg.Question[0].Qtype, ecs.String(), resp)
		}
	}

	pack, err := resp.Pack()
	if err != nil {
		log.Println("failed to Pack", err)
		return
	}
	_, err = w.Write(pack)
	if err != nil {
		log.Println("failed to Write", err)
		return
	}
}

func appendEdnsClientSubnet(msg *dns.Msg, clientIP net.IP) (ecs net.IP) {
	var family uint16
	var sourceNetmask uint8

	if len(clientIP.To4()) == net.IPv4len {
		family = 1
		sourceNetmask = 24
		ecs = clientIP.Mask(net.CIDRMask(int(sourceNetmask), net.IPv4len<<3))
	} else {
		family = 2
		sourceNetmask = 48
		ecs = clientIP.Mask(net.CIDRMask(int(sourceNetmask), net.IPv6len<<3))
	}

	o := &dns.OPT{
		Hdr: dns.RR_Header{
			Name:   ".",
			Rrtype: dns.TypeOPT,
		},
	}

	ed := &dns.EDNS0_SUBNET{
		Code:          dns.EDNS0SUBNET,
		Address:       clientIP,
		Family:        family,
		SourceNetmask: sourceNetmask,
	}
	o.Option = append(o.Option, ed)
	msg.Extra = append(msg.Extra, o)
	return
}
