package raw

type (
	RawConfig struct {
		Listen                                string
		RedirectHttps                         string
		InboundBufferSize, OutboundBufferSize int
		VHosts                                []RawVHost
	}
	RawVHost struct {
		Name          string
		TlsOffloading bool
		ManagedCert   bool
		Cert          string
		Key           string
		KeyType       string
		Alpn          string
		Protocols     string
		Http          RawHttpHandler
		Http2         []RawPathHandler
		Trojan        RawHandler
		Default       RawHandler
	}
	RawHandler struct {
		Handler string
		Args    string
	}
	RawHttpHandler struct {
		Paths   []RawPathHandler
		Handler string
		Args    string
	}
	RawPathHandler struct {
		Path       string
		Handler    string
		Args       string
		TrimPrefix string
	}
)
