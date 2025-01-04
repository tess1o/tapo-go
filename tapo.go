package tapo

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type Transport interface {
	ExecuteRequest(request *RequestSpec) (response json.RawMessage, err error)
}

var DefaultHandshakeDelay = 1 * time.Second

var DefaultRetryConfig = &RetryConfig{
	RetryDelay: 3 * time.Second,
	RetryCount: 3,
}

var DefaultHttpClient = &http.Client{
	Timeout: 10 * time.Second,
}

type RetryConfig struct {
	RetryDelay         time.Duration
	RetryCount         int
	Retry403ErrorsOnly bool
}

type Options struct {
	// HandshakeDelay represents number of seconds to wait after a handshake operation is done
	// Higher amounts are more reliable as device is incredible slow performing authorization internally
	HandshakeDelayDuration time.Duration
	RetryConfig            *RetryConfig
	HttpClient             *http.Client
	EnableDebug            bool
}

type Device struct {
	transport              Transport
	retryConfig            *RetryConfig
	httpClient             *http.Client
	handshakeDelayDuration time.Duration
	enableDebug            bool
}

func NewDevice(transport Transport, options Options) *Device {
	var httpClient *http.Client
	if options.HttpClient != nil {
		httpClient = options.HttpClient
	} else {
		httpClient = DefaultHttpClient
	}

	var retryConfig *RetryConfig
	if options.RetryConfig != nil {
		retryConfig = options.RetryConfig
	}

	var handshakeDelayDuration time.Duration
	if options.HandshakeDelayDuration == 0 {
		handshakeDelayDuration = DefaultHandshakeDelay
	}

	enableDebug := options.EnableDebug

	d := &Device{
		httpClient:             httpClient,
		retryConfig:            retryConfig,
		handshakeDelayDuration: handshakeDelayDuration,
		enableDebug:            enableDebug,
		transport:              transport,
	}

	return d
}

func (d *Device) GenerateTerminalUUID() string {
	newUUID := uuid.New()
	hash := md5.Sum(newUUID[:])
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (d *Device) ExecuteMethod(method string, params json.RawMessage, result any) error {
	request := RequestSpec{
		Method:          method,
		RequestTimeMils: time.Now().UnixNano() / 1000000,
		Params:          params,
		TerminalUUID:    d.GenerateTerminalUUID(),
	}

	stringResponse, err := d.transport.ExecuteRequest(&request)
	if err != nil {
		return err
	}

	return json.Unmarshal(stringResponse, &result)

}
