package dns

import (
	"github.com/miekg/dns"
)

type Do53Client struct {
	server string
}

func NewDo53Client(server string) *Do53Client {
	return &Do53Client{server: server}
}

func (d *Do53Client) Exchange(m *dns.Msg) (msg *dns.Msg, err error) {
	msg, err = dns.Exchange(m, d.server)
	return
}
