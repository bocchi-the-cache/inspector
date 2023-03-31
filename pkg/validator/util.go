package validator

import (
	"encoding/json"
	"github.com/bocchi-the-cache/inspector/pkg/client"
	"github.com/bocchi-the-cache/inspector/pkg/common/logger"
	"net/http"
	"strings"
)

var noCacheHeaders = map[string]struct{}{
	"no-cache":        {},
	"no-store":        {},
	"no-transform":    {},
	"must-revalidate": {},
	"private":         {},
	"max-age=0":       {},
}

func isRangeRequest(r *http.Request) bool {
	return r.Header.Get("Range") != ""
}

func isNoCacheRequest(r *http.Request) bool {
	isNoCache := r.Header.Get("Pragma") == "no-cache"

	CacheControls := strings.Split(r.Header.Get("Cache-Control"), ",")
	for _, CacheControl := range CacheControls {
		if _, ok := noCacheHeaders[CacheControl]; ok {
			isNoCache = true
			break
		}
	}
	// Check if the request is a no cache request
	return isNoCache
}

func MarshalContent(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func UnmarshalContent(data []byte) (*client.Content, error) {
	c := &client.Content{}
	err := json.Unmarshal(data, c)
	return c, err
}

func handlePanic() {

	// detect if panic occurs or not
	a := recover()

	if a != nil {
		logger.Error("[RECOVER] %x", a)
	}

}
