package fkhj212

import (
	log "github.com/sirupsen/logrus"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// MessageHandler is a callback type
type MessageHandler func(FkhjClient, map[string]interface{})

// ConnectionLostHandler is a callback type handle lost connect
type ConnectionLostHandler func(FkhjClient, error)

// OnConnectHandler is a callback that is called when the client
type OnConnectHandler func(FkhjClient)

// ReconnectHandler is invoked prior to reconnecting after
type ReconnectHandler func(FkhjClient, *ClientOptions)

// ClientOptions contains configurable options for an Client.
type ClientOptions struct {
	Servers              []*url.URL
	CleanSession         bool
	Order                bool
	WillRetained         bool
	ProtocolVersion      uint
	KeepAlive            int64
	PingTimeout          time.Duration
	ConnectTimeout       time.Duration
	MaxReconnectInterval time.Duration
	AutoReconnect        bool
	ConnectRetryInterval time.Duration
	ConnectRetry         bool
	DefaultSendHandler   MessageHandler
	OnConnect            OnConnectHandler
	OnConnectionLost     ConnectionLostHandler
	OnReconnecting       ReconnectHandler
	WriteTimeout         time.Duration
}

// NewClientOptions will create a new ClientClientOptions type with some
// default values.
func NewClientOptions() *ClientOptions {
	o := &ClientOptions{
		CleanSession:         true,
		Order:                true,
		ProtocolVersion:      0,
		KeepAlive:            30,
		PingTimeout:          10 * time.Second,
		ConnectTimeout:       30 * time.Second,
		MaxReconnectInterval: 10 * time.Minute,
		AutoReconnect:        true,
		ConnectRetryInterval: 30 * time.Second,
		ConnectRetry:         false,
		OnConnect:            nil,
		OnConnectionLost:     DefaultConnectionLostHandler,
		WriteTimeout:         0, // 0 represents timeout disabled
	}
	return o
}

// AddBroker adds a broker URI to the list of brokers to be used. The format should be
// scheme://host:port
// Where "scheme" is one of "tcp", "ssl", or "ws", "host" is the ip-address (or hostname)
// and "port" is the port on which the broker is accepting connections.
//
// Default values for hostname is "127.0.0.1", for schema is "tcp://".
//
// An example broker URI would look like: tcp://foobar.com:8899
func (o *ClientOptions) AddTargetServer(server string) *ClientOptions {
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
	o.Servers = append(o.Servers, brokerURI)
	return o
}

// SetCleanSession will set the "clean session" flag in the connect message
// when this client connects to an MQTT broker. By setting this flag, you are
// indicating that no messages saved by the broker for this client should be
// delivered. Any messages that were going to be sent by this client before
// diconnecting previously but didn't will not be sent upon connecting to the
// broker.
func (o *ClientOptions) SetCleanSession(clean bool) *ClientOptions {
	o.CleanSession = clean
	return o
}

// SetKeepAlive will set the amount of time (in seconds) that the client
// should wait before sending a PING request to the broker. This will
// allow the client to know that a connection has not been lost with the
// server.
func (o *ClientOptions) SetKeepAlive(k time.Duration) *ClientOptions {
	o.KeepAlive = int64(k / time.Second)
	return o
}

// SetPingTimeout will set the amount of time (in seconds) that the client
// will wait after sending a PING request to the broker, before deciding
// that the connection has been lost. Default is 10 seconds.
func (o *ClientOptions) SetPingTimeout(k time.Duration) *ClientOptions {
	o.PingTimeout = k
	return o
}

// SetDefaultPublishHandler sets the MessageHandler that will be called when a message
// is received that does not match any known subscriptions.
func (o *ClientOptions) SetDefaultPublishHandler(defaultHandler MessageHandler) *ClientOptions {
	o.DefaultSendHandler = defaultHandler
	return o
}

// SetOnConnectHandler sets the function to be called when the client is connected. Both
// at initial connection time and upon automatic reconnect.
func (o *ClientOptions) SetOnConnectHandler(onConn OnConnectHandler) *ClientOptions {
	o.OnConnect = onConn
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
	o.WriteTimeout = t
	return o
}

// SetConnectTimeout limits how long the client will wait when trying to open a connection
// to an MQTT server before timing out and erroring the attempt. A duration of 0 never times out.
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
// in the event of a failure (when true the token returned by the Connect function will
// not complete until the connection is up or it is cancelled)
// If ConnectRetry is true then subscriptions should be requested in OnConnect handler
// Setting this to TRUE permits mesages to be published before the connection is established
func (o *ClientOptions) SetConnectRetry(a bool) *ClientOptions {
	o.ConnectRetry = a
	return o
}
