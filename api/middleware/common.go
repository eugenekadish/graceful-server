package middleware

import (
	"fmt"
	"time"

	"net/http"

	"gitlab.ido-services.com/luxtrust/base-component/util"
	"gitlab.ido-services.com/luxtrust/logging"
	"gitlab.ido-services.com/luxtrust/monitoring"
)

// CommonMiddleware provides a middleware that performs tasks common to all endpoints.
type CommonMiddleware struct {
	config Configurer
	logger logging.Logger
}

// CommonOption provides the client a callback that is used to dynamically specify attributes for a
// CommonMiddleware.
type CommonOption func(*CommonMiddleware)

// WithCommonLogger is used for specifying the Logger for a CommonMiddleware.
func WithCommonLogger(logger logging.Logger) CommonOption {
	return func(cm *CommonMiddleware) { cm.logger = logger }
}

// NewCommonMiddleware is a variadic constructor for an AuthMiddleware.
func NewCommonMiddleware(cfg Configurer, opts ...CommonOption) *CommonMiddleware {

	var defaultLoggerInfo = logging.DefaultLoggerInfo{
		Build:           util.Build,
		Component:       cfg.GetString("name"),
		APIVersion:      util.APIVersion,
		SoftwareVersion: util.SoftwareVersion,
	}
	var log = logging.New(defaultLoggerInfo, "json")

	var cm = &CommonMiddleware{
		config: cfg,
		logger: log,
	}

	for _, opt := range opts {
		opt(cm)
	}

	return cm
}

// Wrapper is a pass through function for handlers that implicitly performs additional business
// logic per request.
func (cm *CommonMiddleware) Wrapper(next http.Handler) http.Handler {

	// NOTE: There is no error handeling in this function
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var code int
		var msg string

		var apiVersion = cm.config.GetString("api.version")
		var prometheusMonitor = cm.config.GetBool("monitoring.prometheus")

		rec := &LogRecord{
			ResponseWriter: w,
			req:            r,
			config:         cm.config,
		}

		start := time.Now()
		next.ServeHTTP(rec, r) // next(rec. r)

		end := time.Now()
		time := end.Sub(start)

		if code = rec.status; code == 0 {
			msg = "slient closed connection"
		} else {
			msg = "served API request"
		}

		cm.logger.
			WithField("endpoint", r.RequestURI).
			WithField("remote_addr", r.RemoteAddr).
			WithField("exec_time", time).
			WithField("http_code", code).
			WithField("http_method", r.Method).
			WithField("user_agent", r.UserAgent()).
			WithField("api_version", apiVersion).
			Info(msg)

		if prometheusMonitor {
			monitoring.APIExecTime.
				WithLabelValues(fmt.Sprintf("%d", code)).
				Observe(time.Seconds())
		}
	})
}
