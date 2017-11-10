package main

// URLRoute present the single rule for url
type URLRoute struct {
	URLPrefix string `json:"url_prefix"`
	Path      string `json:"path"`
	URLType   string `json:"type"`
	DocRoot   string `json:"docroot"`
	UseRegex  bool   `json:"use_regex"`
}

// ListenEntry present a port listen to
type ListenEntry struct {
	Cert        string `json:"cert"`
	Host        string `json:"host"`
	EnableProxy bool   `json:"enable_proxy"`
	Key         string `json:"key"`
	ProxyAuth   bool   `json:"proxy_auth"`
	ProxyUser   string `json:"proxy_user"`
	ProxyPasswd string `json:"proxy_passwd"`
}

// CFG is configure struct
type CFG struct {
	Routes       []URLRoute    `json:"url_routes"`
	Host         []ListenEntry `json:"listen"`
	LocalDomains []string      `json:"local_domains"`
	LogFile      string        `json:"logfile"`
	User         string        `json:"user"`
	Group        string        `json:"group"`
}
