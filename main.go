package main

import (
	"github.com/bocchi-the-cache/inspector/pkg/client"
	"github.com/bocchi-the-cache/inspector/pkg/common/logger"
	"github.com/bocchi-the-cache/inspector/pkg/common/result_logger"
	"github.com/bocchi-the-cache/inspector/pkg/monitor"
	"github.com/bocchi-the-cache/inspector/pkg/server"
	"github.com/bocchi-the-cache/inspector/pkg/storage"
	"github.com/spf13/viper"
)

func initLog() {
	logger.InitLogger("log", "log.txt", "debug")
}

func initResultLog() {
	result_logger.InitLogger("log", "result.txt", "debug")
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config/")
	err := viper.ReadInConfig()
	if err != nil {
		logger.Panic("config init failed. exit!")
		panic(err)
	}
}

func initStorage() {
	storage.Init()
}

func initClient() {
	client.Init()
}

func initMonitor() {
	monitor.Init()
}

func main() {
	initLog()
	initResultLog()
	initConfig()
	initStorage()
	initClient()
	initMonitor()

	logger.Info("all init done, start server")
	server.Serve()
}
