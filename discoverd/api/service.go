package api

type MicroApp struct {
	Addr      string   `json:"addr"`
	Providers []string `json:"providers"`
	Consumers []string `json:"consumers"`
}

type AppService struct {
	Addr     string   `json:"addr"`
	Services []string `json:"services"`
}

type AppStatus struct {
	IsLive   bool   `json:"isLive"`
	RouterCS string `json:"routerCS"`
}

type RouterTable struct {
	Routers  []Router `json:"routers"`
	Checksum string   `json:"checksum"`
}

type NodeAddr struct {
	Node string `json:"node"`
	Addr string `json:"addr"`
}

type Router struct {
	Service  string     `json:"service"`
	Addrs    []NodeAddr `json:"addrs"`
	Checksum []byte     `json:"checksum,omitempty"`
}

type InnerAppService struct {
	NodeAddr NodeAddr `json:"nodeaddr"`
	Services []string `json:"services"`
}
