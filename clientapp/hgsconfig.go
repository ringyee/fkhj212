package clientapp

// UpServers ....
type UpServers struct {
	UpServers []UpServer
}

// UpServer ....
type UpServer struct {
	Address    string
	Port       int64
	MN         string
	Protocol   string
	Uptime     string
	HBEnable   bool
	HBInterval int64
}

// ConfFactors config and forctors
type ConfFactors struct {
	Factors []ConfFactor `json:"factors"`
}

// ConfFS config FS
type ConfFS struct {
	NM string  `json:"nm"`
	ID string  `json:"id"`
	CI int64   `json:"ci"`
	UN string  `json:"un"`
	RL float32 `json:"rl"`
	RH float32 `json:"rh"`
}

// ConfFactor config and forctor
type ConfFactor struct {
	CO string   `json:"co"`
	BR int      `json:"br"`
	DB int      `json:"db"`
	PB string   `json:"pb"`
	SB int      `json:"sb"`
	AR uint8    `json:"ar"`
	PC string   `json:"pc"`
	TP string   `json:"tp"`
	FS []ConfFS `json:"fs"`
}

//init config
var (
	ConfPath     = `/etc/lchj212/`
	UpS          UpServers
	CF           ConfFactors
	ReConfig     = make(chan struct{})
	ReConfigFlag bool
)
