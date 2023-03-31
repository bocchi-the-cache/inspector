package client

import (
	"fmt"
	"github.com/bocchi-the-cache/inspector/pkg/common/logger"
	"github.com/bocchi-the-cache/inspector/pkg/monitor"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

var BaselineFetcher *Fetcher
var TestFetcher *Fetcher

func Init() {
	BaselineFetcher = NewHttpFetcher(viper.GetString("host.baseline"))
	TestFetcher = NewHttpFetcher(viper.GetString("host.test"))
}

type Fetcher struct {
	HttpClient  *http.Client
	RewriteHost string
}

func NewHttpFetcher(rHost string) *Fetcher {
	c := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          500,
			MaxIdleConnsPerHost:   500,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: time.Duration(15) * time.Second,
	}
	f := &Fetcher{
		HttpClient:  c,
		RewriteHost: rHost,
	}
	return f
}

func (f *Fetcher) Do(r *http.Request) (int, http.Header, []byte, error) {
	// TODO: Host rewrite
	// For client requests, the URL's Host specifies the server to
	// connect to, while the Request's Host field optionally
	req, err := http.NewRequest(r.Method, fmt.Sprintf("http://%s%s", r.Host, r.URL.String()), r.Body)
	if err != nil {
		monitor.RequestSendTotalCounterIncr(r.Method, r.Host, f.RewriteHost, "ErrorRequest")
		logger.Error("new request error, err: %s", err)
	}
	req.Header = r.Header
	req.URL.Host = f.RewriteHost
	resp, err := f.HttpClient.Do(req)
	if err != nil {
		monitor.RequestSendTotalCounterIncr(r.Method, r.Host, f.RewriteHost, "ErrorSend")
		return 0, nil, nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		monitor.RequestSendTotalCounterIncr(r.Method, r.Host, f.RewriteHost, "ErrorReadBody")
		return resp.StatusCode, nil, nil, err
	}

	monitor.RequestSendTotalCounterIncr(r.Method, r.Host, f.RewriteHost, resp.Status)
	return resp.StatusCode, resp.Header, body, nil
}
