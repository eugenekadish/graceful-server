package middleware

import (
	"net/http"
)

// Configurer provides configuration data necessary for middleware.
type Configurer interface {
	GetBool(string) bool
	GetString(string) string
}

// LogRecord records parameters for responses.
type LogRecord struct {
	http.ResponseWriter
	status int
	req    *http.Request
	config Configurer
}

// Write overrides the http.ResponseWriter's Write function by setting a header and a status code.
func (r *LogRecord) Write(p []byte) (int, error) {

	var apiVersion = r.config.GetString("api.version")

	if r.req.Context().Err() != nil {
		return 0, nil
	}

	if r.status == 0 {
		r.status = 200
	}

	r.Header().Add("api-version", apiVersion)

	return r.ResponseWriter.Write(p)
}

// WriteHeader overrides the http.ResponseWriter's WriteHeader function by setting a header and a
// status code.
func (r *LogRecord) WriteHeader(status int) {

	var apiVersion = r.config.GetString("api.version")

	if r.req.Context().Err() != nil {
		return
	}

	r.status = status
	r.Header().Add("api-version", apiVersion)
	r.ResponseWriter.WriteHeader(status)
}

func init() {
	// Initialization . . .
}
