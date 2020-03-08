package clientapp

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yjiong/fkhj212/packets"
)

// ExCmdHandler is a callback type
type ExCmdHandler func(Fkhjer, packets.Pduer)

// ConnectionLostHandler is a callback type handle lost connect
type ConnectionLostHandler func(Fkhjer, error)

// OnConnectHandler is a callback that is called when the client
type OnConnectHandler func(Fkhjer)

// ReconnectHandler is invoked prior to reconnecting after
type ReconnectHandler func(Fkhjer, *ClientOptions)

// ClientOptions contains configurable options for an Client.
type ClientOptions struct {
	Server               *url.URL              `mapstructure:"server"`
	ServerStr            string                `mapstructure:"seraddress"`
	CleanSession         bool                  `mapstructure:"clean_session"`
	ProtocolVersion      string                `mapstructure:"protocol_version"`
	KeepAlive            int64                 `mapstructure:"keepalive"`
	PingTimeout          time.Duration         `mapstructure:"pingtimeout"`
	ConnectTimeout       time.Duration         `mapstructure:"connect_timeout"`
	MaxReconnectInterval time.Duration         `mapstructure:"max_reconnect_interval"`
	AutoReconnect        bool                  `mapstructure:"auto_reconnect"`
	ConnectRetryInterval time.Duration         `mapstructure:"connect_retry_interval"`
	ConnectRetry         bool                  `mapstructure:"connect_retry"`
	ExCmdHandle          ExCmdHandler          `mapstructure:"excmd_handle"`
	OnConnect            OnConnectHandler      `mapstructure:"on_connect"`
	OnConnectionLost     ConnectionLostHandler `mapstructure:"on_connection_lost"`
	OnReconnecting       ReconnectHandler      `mapstructure:"on_reconnecting"`
	OverTime             time.Duration         `mapstructure:"overtime"`
	MN                   string                `mapstructure:"mn"`
	PW                   string                `mapstructure:"pw"`
	ST                   string                `mapstructure:"st"`
	ReCount              int64                 `mapstructure:"recount"`
	RtdInterval          time.Duration         `mapstructure:"rtdinterval"`
}

func initCOP() *ClientOptions {
	return &ClientOptions{
		CleanSession:         true,
		ProtocolVersion:      "HJ212-2017",
		KeepAlive:            0,
		PingTimeout:          10,
		ConnectTimeout:       10,
		MaxReconnectInterval: 10,
		AutoReconnect:        true,
		ExCmdHandle:          nil,
		ConnectRetryInterval: 30,
		ConnectRetry:         false,
		OnConnect:            nil,
		OnConnectionLost:     DefaultConnectionLostHandler,
		OverTime:             5,
		MN:                   "000000000000000000000000",
		PW:                   "123456",
		ST:                   "99",
		ReCount:              3,
		RtdInterval:          30,
	}
}

// GetConf ....
func GetConf() *viper.Viper {
	v := viper.New()
	v.SetConfigType("yml")
	v.SetConfigName("config")
	v.AddConfigPath(ConfPath)
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.config/lchj212")
	return v
}

// NewClientOptions will create a new ClientClientOptions type with some
// default values.
func NewClientOptions() *ClientOptions {
	v := GetConf()
	o := initCOP()
	if err := v.ReadInConfig(); err == nil {
		v.Unmarshal(&o)
		if o.ServerStr != "" {
			o.SetTargetServer(o.ServerStr)
		}
	} else {
		log.Error(err)
	}
	o.PingTimeout *= time.Second
	o.ConnectTimeout *= time.Second
	o.MaxReconnectInterval *= time.Minute
	o.ConnectRetryInterval *= time.Second
	o.OverTime *= time.Second
	o.RtdInterval *= time.Second
	//log.Debugf("viper Umarshal optiongs is %+v\n", o)
	return o
}

// SetTargetServer adds a broker URI to the list of brokers to be used. The format should be
// scheme://host:port
// Where "scheme" is one of "tcp"
// Default values for hostname is "127.0.0.1", for schema is "tcp://".
// An example broker URI would look like: tcp://foobar.com:8899
func (o *ClientOptions) SetTargetServer(server string) *ClientOptions {
	re := regexp.MustCompile(`%(25)?`)
	if len(server) > 0 && server[0] == ':' {
		server = "127.0.0.1" + server
	}
	if !strings.Contains(server, "://") {
		server = "tcp://" + server
	}
	server = re.ReplaceAllLiteralString(server, "%25")
	brokerURI, err := url.Parse(server)
	if err != nil {
		log.Errorf("Failed to parse %q broker address: %s", server, err)
		return o
	}
	o.Server = brokerURI
	return o
}

// SetMN will set the amount of time (in seconds)
func (o *ClientOptions) SetMN(mn string) *ClientOptions {
	o.MN = mn
	return o
}

// SetCleanSession will set the "clean session" flag in the connect message
func (o *ClientOptions) SetCleanSession(clean bool) *ClientOptions {
	o.CleanSession = clean
	return o
}

// SetKeepAlive will set the amount of time (in seconds)
func (o *ClientOptions) SetKeepAlive(k time.Duration) *ClientOptions {
	o.KeepAlive = int64(k / time.Second)
	return o
}

// SetPingTimeout will set the amount of time (in seconds)
func (o *ClientOptions) SetPingTimeout(k time.Duration) *ClientOptions {
	o.PingTimeout = k
	return o
}

// SetOnConnectHandler sets the function to be called when the client is connected
func (o *ClientOptions) SetOnConnectHandler(onConn OnConnectHandler) *ClientOptions {
	o.OnConnect = onConn
	return o
}

// SetExCmdHandle sets the function to be exec server cmd
func (o *ClientOptions) SetExCmdHandle(exc ExCmdHandler) *ClientOptions {
	o.ExCmdHandle = exc
	return o
}

// SetConnectionLostHandler will set the OnConnectionLost callback to be executed
// in the case where the client unexpectedly loses connection with the MQTT broker.
func (o *ClientOptions) SetConnectionLostHandler(onLost ConnectionLostHandler) *ClientOptions {
	o.OnConnectionLost = onLost
	return o
}

// SetReconnectingHandler sets the OnReconnecting callback to be executed prior
// to the client attempting a reconnect to the MQTT broker.
func (o *ClientOptions) SetReconnectingHandler(cb ReconnectHandler) *ClientOptions {
	o.OnReconnecting = cb
	return o
}

// SetWriteTimeout puts a limit on how long a mqtt publish should block until it unblocks with a
// timeout error. A duration of 0 never times out. Default 30 seconds
func (o *ClientOptions) SetWriteTimeout(t time.Duration) *ClientOptions {
	o.OverTime = t
	return o
}

// SetConnectTimeout limits how long the client will wait when trying to open a connection
// Default 30 seconds. Currently only operational on TCP/TLS connections.
func (o *ClientOptions) SetConnectTimeout(t time.Duration) *ClientOptions {
	o.ConnectTimeout = t
	return o
}

// SetMaxReconnectInterval sets the maximum time that will be waited between reconnection attempts
// when connection is lost
func (o *ClientOptions) SetMaxReconnectInterval(t time.Duration) *ClientOptions {
	o.MaxReconnectInterval = t
	return o
}

// SetAutoReconnect sets whether the automatic reconnection logic should be used
// when the connection is lost, even if disabled the ConnectionLostHandler is still
// called
func (o *ClientOptions) SetAutoReconnect(a bool) *ClientOptions {
	o.AutoReconnect = a
	return o
}

// SetConnectRetryInterval sets the time that will be waited between connection attempts
// when initially connecting if ConnectRetry is TRUE
func (o *ClientOptions) SetConnectRetryInterval(t time.Duration) *ClientOptions {
	o.ConnectRetryInterval = t
	return o
}

// SetConnectRetry sets whether the connect function will automatically retry the connection
func (o *ClientOptions) SetConnectRetry(a bool) *ClientOptions {
	o.ConnectRetry = a
	return o
}
