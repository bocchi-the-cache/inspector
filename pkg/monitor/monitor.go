package monitor

import (
	"net"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestReceiveTotalCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "bocchi_inspector_request_receive_total",
		Help: "灰度工具收到的请求总数",
	}, []string{"node", "method", "host"})

	RequestSendTotalCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "bocchi_inspector_request_send_total",
		Help: "灰度工具发出的请求总数",
	}, []string{"node", "method", "host", "dst", "status"})

	ResultTotalCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "bocchi_inspector_result_total",
		Help: "灰度工具对比结果",
	}, []string{"node", "method", "status"})

	ErrorTotalCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "bocchi_inspector_error_total",
		Help: "灰度工具错误统计",
	}, []string{"node", "method", "process", "error"})

	ElapsedMonitor = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "RequestTimeConsumingStatistics",
		Help:    "Api请求耗时统计",
		Buckets: []float64{0.1, 0.5, 1, 5, 10, 20, 50, 100, 500, 1000, 5000},
	}, []string{"node", "process"})

	node = "unknown"
)

func RequestReceiveTotalCounterIncr(method, host string) {
	RequestReceiveTotalCounter.WithLabelValues(node, method, host).Inc()
}

func RequestSendTotalCounterIncr(method, host, dst, status string) {
	RequestSendTotalCounter.WithLabelValues(node, method, host, dst, status).Inc()
}

func ResultTotalCounterIncr(method, status string) {
	ResultTotalCounter.WithLabelValues(node, method, status).Inc()
}

func ErrorTotalCounterIncr(method, process, error string) {
	ErrorTotalCounter.WithLabelValues(node, method, process, error).Inc()
}

func ElapsedMonitorIncr(process string, elapsed float64) {
	ElapsedMonitor.WithLabelValues(node, process).Observe(elapsed)
}

func Init() {
	node = getNodeIp()
	prometheus.MustRegister(RequestReceiveTotalCounter, RequestSendTotalCounter, ResultTotalCounter, ErrorTotalCounter, ElapsedMonitor)
}

// Get node ip by net.InterfaceAddrs()
func getNodeIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "unknown"
}
