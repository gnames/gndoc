package gnhttp

import (
	"io"
	"time"
)

// GNhttp sends a request to a remote url and gets back retuls.
type GNhttp interface {
	// Get takes a URL and returns back the response status code, MIME type,
	// body of the response and a possible error during the request.
	Get(URL string) (int, string, io.Reader, error)

	// SetConnMax sets maximum wait time for attempting a connection.
	SetConnMax(time.Duration)

	// ConnMax returns wait time for establishing connection.
	ConnMax() time.Duration

	// SetReqMax sets maximum wait time for a result of a request.
	SetReqMax(time.Duration)

	// ReqMax returns wait time for processing an HTTP request.
	ReqMax() time.Duration
}
