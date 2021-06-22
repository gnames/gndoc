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
	// Sets maximum wait time for attempting a connection.
	ConnMax(time.Duration)
	// Sets maximum wait time for a result of a request.
	ReqMax(time.Duration)
}
