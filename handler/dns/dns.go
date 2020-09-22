package dns

import "github.com/miekg/dns"

type DnsClient interface {
	Exchange(m *dns.Msg) (msg *dns.Msg, err error)
}
