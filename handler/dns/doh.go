package dns

import (
	"bytes"
	"github.com/miekg/dns"
	"io/ioutil"
	"net/http"
)

const dohMimeType = "application/dns-message"

type DohClient struct {
	url    string
	client *http.Client
}

func (d *DohClient) Exchange(m *dns.Msg) (msg *dns.Msg, err error) {
	buf, err := m.Pack()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, d.url, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", dohMimeType)
	req.Header.Set("accept", dohMimeType)

	resp, err := d.client.Do(req)

	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	buf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	msg = &dns.Msg{}
	err = msg.Unpack(buf)
	return msg, err
}

func NewDohClient(server string) *DohClient {
	return &DohClient{
		url: server,
		client: &http.Client{Transport: &http.Transport{
			ForceAttemptHTTP2: true,
		}},
	}
}
