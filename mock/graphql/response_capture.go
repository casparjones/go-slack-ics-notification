package graphql

import (
	"bytes"
	"net/http"
)

// ResponseCapture fängt die Response ab, bevor sie an den Client geschickt wird.
type ResponseCapture struct {
	http.ResponseWriter
	buf        bytes.Buffer
	statusCode int
}

func (rc *ResponseCapture) WriteHeader(code int) {
	rc.statusCode = code
	// WriteHeader wird hier nicht sofort ausgeführt, sondern später
}

func (rc *ResponseCapture) Write(b []byte) (int, error) {
	return rc.buf.Write(b)
}
