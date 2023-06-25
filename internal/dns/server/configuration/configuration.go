package configuration

type udpEndpoint struct {
	Enabled bool
	Address string `json:"address"`
}

type externalSource struct {
	Type     string `json:"type"`
	Endpoint string `json:"endpoint"`
}

type custom struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type cache struct {
	Size    int64  `json:"size"`
	Basettl uint32 `json:"basettl"`
}

//ServerConf represents the configuration of the dns server
type ServerConf struct {
	AllowExternal bool           `json:"allow_external"`
	BlockingLists []string       `json:"blocking_list"`
	Custom        []custom       `json:"custom"`
	Cache         cache          `json:"cache"`
	External      externalSource `json:"external"`
	Endpoint      udpEndpoint    `json:"endpoint"`
}

//Default generate the default configuration
func Default() ServerConf {
	return ServerConf{
		AllowExternal: true,
		BlockingLists: []string{
			"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
		},
		Custom: []custom{
			{"cloudflare-dns.com", "104.16.249.249"},
			{"cloudflare-dns.com", "2606:4700::6810:f8f"},
		},
		Cache: cache{
			Size:    1000000,
			Basettl: 600,
		},
		External: externalSource{
			Type:     "DOH",
			Endpoint: "https://cloudflare-dns.com/dns-query",
		},
		Endpoint: udpEndpoint{
			Enabled: true,
			Address: "127.0.0.1:53",
		},
	}
}

// Client(Blocker) -> Client(Memory) -> Cache(Multiple(Client(udp/https)))
