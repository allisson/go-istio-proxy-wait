# go-istio-proxy-wait
[![Build Status](https://github.com/allisson/go-istio-proxy-wait/workflows/tests/badge.svg)](https://github.com/allisson/go-istio-proxy-wait/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/allisson/go-istio-proxy-wait)](https://goreportcard.com/report/github.com/allisson/go-istio-proxy-wait)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/allisson/go-istio-proxy-wait)

Wait until the istio-proxy is working.

## Example

```golang
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	istioproxy "github.com/allisson/go-istio-proxy-wait"
)

func main() {
	// If "ISTIO_PROXY_ENABLED" != "true" a mock of the Proxy interface will be returned.
	os.Setenv("ISTIO_PROXY_ENABLED", "true")

	// Create a new proxy.
	timeout := time.Duration(1) * time.Second
	retryDelay := time.Duration(1) * time.Second
	maxRetries := 10
	istioProxy := istioproxy.New(timeout, retryDelay, maxRetries)

	// Wait until the istio-proxy is ready.
	if err := istioProxy.Wait(); err != nil {
		log.Fatalf("wait-for-istio-proxy, error=%s", err.Error())
	}

	// Execute the code.
	fmt.Println("Hello")

    // Close istio-proxy.
    // You could also use the defer istioProxy.Close() after istioproxy.New().
	if err := istioProxy.Close(); err != nil {
		log.Fatalf("close-istio-proxy, error=%s", err.Error())
	}
}
```
