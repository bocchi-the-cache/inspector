package server

import (
	"github.com/bocchi-the-cache/inspector/pkg/common/logger"
	"github.com/bocchi-the-cache/inspector/pkg/validator"
	"github.com/spf13/viper"
	"io"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Serve() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/", dispatchRequest)

	logger.Infof("*** start http server, listen port: %s", viper.GetString("http.listen_port"))
	logger.Infof("*** metrics endpoint: %s", "/metrics")
	logger.Infof("*** note: Inspector returns 200 OK immediately, and validate the request in background.")
	logger.Infof("*** only **GET** requests will be validated. ")
	err := http.ListenAndServe(":"+viper.GetString("http.listen_port"), mux)
	if err != nil {
		logger.Panicf("http server error, err: %s", err)
	}
}

func dispatchRequest(w http.ResponseWriter, r *http.Request) {
	// TODO Chan Pipeline
	validator.PushRequest(r)
	_, err := io.WriteString(w, "Hello, HTTP!\n")
	if err != nil {
		logger.Errorf("write response error, err: %s", err)
		return
	}
}
