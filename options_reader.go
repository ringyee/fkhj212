package fkhj212

import (
	"net/url"
	"time"
)

// ClientOptionsReader provides an interface for reading ClientOptions after the client has been initialized.
type ClientOptionsReader struct {
	options *ClientOptions
}

//Servers returns a slice of the servers defined in the clientoptions
func (r *ClientOptionsReader) Servers() []*url.URL {
	s := make([]*url.URL, len(r.options.Servers))

	for i, u := range r.options.Servers {
		nu := *u
		s[i] = &nu
	}

	return s
}

//CleanSession returns whether Cleansession is set
func (r *ClientOptionsReader) CleanSession() bool {
	s := r.options.CleanSession
	return s
}

func (r *ClientOptionsReader) Order() bool {
	s := r.options.Order
	return s
}

func (r *ClientOptionsReader) ProtocolVersion() uint {
	s := r.options.ProtocolVersion
	return s
}

func (r *ClientOptionsReader) KeepAlive() time.Duration {
	s := time.Duration(r.options.KeepAlive * int64(time.Second))
	return s
}

func (r *ClientOptionsReader) PingTimeout() time.Duration {
	s := r.options.PingTimeout
	return s
}

func (r *ClientOptionsReader) ConnectTimeout() time.Duration {
	s := r.options.ConnectTimeout
	return s
}

func (r *ClientOptionsReader) MaxReconnectInterval() time.Duration {
	s := r.options.MaxReconnectInterval
	return s
}

func (r *ClientOptionsReader) AutoReconnect() bool {
	s := r.options.AutoReconnect
	return s
}

//ConnectRetryInterval returns the delay between retries on the initial connection (if ConnectRetry true)
func (r *ClientOptionsReader) ConnectRetryInterval() time.Duration {
	s := r.options.ConnectRetryInterval
	return s
}

//ConnectRetry returns whether the initial connection request will be retried until connection established
func (r *ClientOptionsReader) ConnectRetry() bool {
	s := r.options.ConnectRetry
	return s
}

func (r *ClientOptionsReader) WriteTimeout() time.Duration {
	s := r.options.WriteTimeout
	return s
}
