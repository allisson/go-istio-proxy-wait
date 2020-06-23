package istioproxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"
)

func newServer(port, statusCode int, endpoint, response string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(statusCode)
		// nolint
		w.Write([]byte(response))
	})
	server := httptest.NewUnstartedServer(mux)
	l, _ := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	server.Listener = l
	return server
}

func assertEqual(t *testing.T, expected, current interface{}) {
	t.Helper()

	if !reflect.DeepEqual(expected, current) {
		t.Fatalf("assertion_type=Equal, expected_value=%#v, expected_type=%T, current_value=%#v, current_type=%T", expected, expected, current, current)
	}
}

func TestProxy(t *testing.T) {
	timeout := time.Duration(10) * time.Millisecond
	retryDelay := time.Duration(10) * time.Millisecond
	os.Setenv("ISTIO_PROXY_ENABLED", "true")

	t.Run("Wait success", func(t *testing.T) {
		serverInfoResponse := `{"version":"9b4239dee83dd8894bfc579d412ccd894cff2597/1.13.1-dev/Clean/RELEASE/BoringSSL","state":"LIVE","hot_restart_version":"11.104","command_line_options":{"base_id":"0","concurrency":2,"config_path":"/etc/istio/proxy/envoy-rev0.json","config_yaml":"","allow_unknown_static_fields":false,"reject_unknown_dynamic_fields":false,"admin_address_path":"","local_address_ip_version":"v4","log_level":"info","component_log_level":"misc:error","log_format":"[Envoy (Epoch 0)] [%Y-%m-%d %T.%e][%t][%l][%n] %v","log_format_escaped":false,"log_path":"","service_cluster":"redis.cardprocessor","service_node":"sidecar~10.0.4.5~redis.cardprocessor~cardprocessor.svc.cluster.local","service_zone":"","mode":"Serve","hidden_envoy_deprecated_max_stats":"0","hidden_envoy_deprecated_max_obj_name_len":"0","disable_hot_restart":false,"enable_mutex_tracing":false,"restart_epoch":0,"cpuset_threads":false,"disabled_extensions":[],"file_flush_interval":"10s","drain_time":"45s","parent_shutdown_time":"60s"},"uptime_current_epoch":"131s","uptime_all_epochs":"131s"}`
		maxRetries := 10
		server := newServer(15000, http.StatusOK, "/server_info", serverInfoResponse)
		server.Start()
		defer server.Close()

		p := New(timeout, retryDelay, maxRetries)
		err := p.Wait()
		assertEqual(t, nil, err)
	})

	t.Run("Wait fail state", func(t *testing.T) {
		serverInfoResponse := `{"version":"9b4239dee83dd8894bfc579d412ccd894cff2597/1.13.1-dev/Clean/RELEASE/BoringSSL","state":"TAO-DEIXANDO-A-GENTE-SONHAR","hot_restart_version":"11.104","command_line_options":{"base_id":"0","concurrency":2,"config_path":"/etc/istio/proxy/envoy-rev0.json","config_yaml":"","allow_unknown_static_fields":false,"reject_unknown_dynamic_fields":false,"admin_address_path":"","local_address_ip_version":"v4","log_level":"info","component_log_level":"misc:error","log_format":"[Envoy (Epoch 0)] [%Y-%m-%d %T.%e][%t][%l][%n] %v","log_format_escaped":false,"log_path":"","service_cluster":"redis.cardprocessor","service_node":"sidecar~10.0.4.5~redis.cardprocessor~cardprocessor.svc.cluster.local","service_zone":"","mode":"Serve","hidden_envoy_deprecated_max_stats":"0","hidden_envoy_deprecated_max_obj_name_len":"0","disable_hot_restart":false,"enable_mutex_tracing":false,"restart_epoch":0,"cpuset_threads":false,"disabled_extensions":[],"file_flush_interval":"10s","drain_time":"45s","parent_shutdown_time":"60s"},"uptime_current_epoch":"131s","uptime_all_epochs":"131s"}`
		maxRetries := 2
		server := newServer(15000, http.StatusOK, "/server_info", serverInfoResponse)
		server.Start()
		defer server.Close()
		expectedError := fmt.Sprintf("wait_max_retries_exceeded, max_retries=%d", maxRetries)

		p := New(timeout, retryDelay, maxRetries)
		err := p.Wait()
		assertEqual(t, expectedError, err.Error())
	})

	t.Run("Wait fail with connection refused", func(t *testing.T) {
		maxRetries := 2
		expectedError := fmt.Sprintf("wait_max_retries_exceeded, max_retries=%d", maxRetries)

		// Execute without server to respond
		p := New(timeout, retryDelay, maxRetries)
		err := p.Wait()
		assertEqual(t, expectedError, err.Error())
	})

	t.Run("Close success", func(t *testing.T) {
		serverQuitResponse := `{"success": true}`
		maxRetries := 10
		server := newServer(15020, http.StatusOK, "/quitquitquit", serverQuitResponse)
		server.Start()
		defer server.Close()

		p := New(timeout, retryDelay, maxRetries)
		err := p.Close()
		assertEqual(t, nil, err)
	})

	t.Run("Close fail with status not equals 200", func(t *testing.T) {
		serverQuitResponse := `{"success": true}`
		maxRetries := 2
		expectedError := fmt.Sprintf("close_max_retries_exceeded, max_retries=%d", maxRetries)
		server := newServer(15020, http.StatusInternalServerError, "/quitquitquit", serverQuitResponse)
		server.Start()
		defer server.Close()

		p := New(timeout, retryDelay, maxRetries)
		err := p.Close()
		assertEqual(t, expectedError, err.Error())
	})

	t.Run("Close fail with connection refused", func(t *testing.T) {
		maxRetries := 2
		expectedError := fmt.Sprintf("close_max_retries_exceeded, max_retries=%d", maxRetries)

		p := New(timeout, retryDelay, maxRetries)
		err := p.Close()
		assertEqual(t, expectedError, err.Error())
	})
}
