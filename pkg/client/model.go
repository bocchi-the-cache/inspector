package client

import "net/http"

type Content struct {
	Status  int
	Header  http.Header
	Content []byte
}
