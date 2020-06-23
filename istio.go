package istioproxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// getBool returns a boolean value from environment variable or default value.
func getBool(key string, defaultValue bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	result, err := strconv.ParseBool(val)
	if err != nil {
		return defaultValue
	}

	return result
}

// Proxy interface to perform operations with istio-proxy.
type Proxy interface {
	Wait() error
	Close() error
}

type serverInfoResponse struct {
	State string `json:"state"`
}

// proxy allows you to perform operations with istio-proxy.
type proxy struct {
	serverInfoAddress string
	serverQuitAddress string
	client            *http.Client
	retryDelay        time.Duration
	maxRetries        int
}

// Wait until the istio-proxy is ready.
func (p *proxy) Wait() error {
	retries := 0

	for {
		retries++
		if retries > p.maxRetries {
			return fmt.Errorf("wait_max_retries_exceeded, max_retries=%d", p.maxRetries)
		}

		response, err := p.client.Get(p.serverInfoAddress)
		if err != nil {
			log.Printf("wait_client_get, retries=%d, max_retries=%d, error=%s", retries, p.maxRetries, err.Error())
			time.Sleep(p.retryDelay)
			continue
		}

		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("wait_ioutil_readall, retries=%d, max_retries=%d, error=%s", retries, p.maxRetries, err.Error())
			time.Sleep(p.retryDelay)
			continue
		}

		serverResponse := serverInfoResponse{}
		if err := json.Unmarshal(responseBody, &serverResponse); err != nil {
			log.Printf("wait_json_unmarshal, retries=%d, max_retries=%d, error=%s", retries, p.maxRetries, err.Error())
			time.Sleep(p.retryDelay)
			continue
		}
		if serverResponse.State == "LIVE" {
			break
		} else {
			log.Printf("wait_server_response_state, retries=%d, max_retries=%d, state=%s", retries, p.maxRetries, serverResponse.State)
		}

		time.Sleep(p.retryDelay)
	}

	return nil
}

// Close istio-proxy.
func (p *proxy) Close() error {
	retries := 0

	for {
		retries++
		if retries > p.maxRetries {
			return fmt.Errorf("close_max_retries_exceeded, max_retries=%d", p.maxRetries)
		}

		response, err := p.client.Post(p.serverQuitAddress, "application/json", nil)
		if err != nil {
			log.Printf("close_client_post, retries=%d, max_retries=%d, error=%s", retries, p.maxRetries, err.Error())
			time.Sleep(p.retryDelay)
			continue
		}

		if response.StatusCode == http.StatusOK {
			break
		} else {
			log.Printf("close_response_status, retries=%d, max_retries=%d, status=%d", retries, p.maxRetries, response.StatusCode)
			time.Sleep(p.retryDelay)
			continue
		}

	}

	return nil
}

// mockProxy is a mock for Proxy interface.
type mockProxy struct{}

func (m *mockProxy) Wait() error {
	return nil
}

func (m *mockProxy) Close() error {
	return nil
}

// New returns a new Proxy based on ISTIO_PROXY_ENABLED environment variable.
// If "ISTIO_PROXY_ENABLED" != "true" a mock of the Proxy interface will be returned.
func New(timeout, retryDelay time.Duration, maxRetries int) Proxy {
	istioProxyEnabled := getBool("ISTIO_PROXY_ENABLED", false)
	if !istioProxyEnabled {
		return &mockProxy{}
	}
	return &proxy{
		serverInfoAddress: "http://localhost:15000/server_info",
		serverQuitAddress: "http://localhost:15020/quitquitquit",
		client:            &http.Client{Timeout: timeout},
		retryDelay:        retryDelay,
		maxRetries:        maxRetries,
	}
}
