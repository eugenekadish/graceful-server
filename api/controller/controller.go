package controller

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.ido-services.com/luxtrust/base-component/util"
	"gitlab.ido-services.com/luxtrust/logging"
)

// Middleware is used to for providing additional logic on endpoints registered with a Controller.
type Middleware interface {
	Wrapper(next http.Handler) http.Handler
}

// MuxRouter sets up the Request Handlers for the server.
type MuxRouter interface {
	Handle(string, http.Handler) *mux.Route
}

// Endpoint is the URLs and corresponding HTTP verbs that are used to identify Handlers.
type Endpoint struct {
	URL     string
	Methods map[string]http.HandlerFunc
}

// Controller is used for registering enpoints with middlewares that can be specified by a client.
type Controller struct {
	router MuxRouter
	logger logging.Logger
}

// Option provides the client a callback that is used to dynamically specify attributes for a
// Controller.
type Option func(*Controller)

// WithRouter is an Option for specifying the Router for a Controller.
func WithRouter(router MuxRouter) Option {
	return func(c *Controller) { c.router = router }
}

// WithLogger is an Option for specifying the Logger for a Controller.
func WithLogger(logger logging.Logger) Option {
	return func(c *Controller) { c.logger = logger }
}

// New is a variadic constructor for a Controller.
func New(rtr MuxRouter, opts ...Option) *Controller {

	var defaultLoggerInfo = logging.DefaultLoggerInfo{
		Build:           util.Build,
		Component:       "base-component",
		APIVersion:      util.APIVersion,
		SoftwareVersion: util.SoftwareVersion,
	}
	var log = logging.New(defaultLoggerInfo, "json")

	var c = &Controller{
		router: rtr,
		logger: log,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Register is called for setting up endpoints with handlers.
func (c *Controller) Register(endpoints []Endpoint, mids ...Middleware) error {

	var (
		method string

		handler    http.Handler
		handlefunc http.HandlerFunc

		end Endpoint
		mid Middleware
	)

	for _, end = range endpoints {
		for method, handlefunc = range end.Methods {
			switch method {
			case "GET", "POST", "PUT", "PATCH", "DELETE":
				handler = handlefunc
				for _, mid = range mids {
					handler = mid.Wrapper(handler)
				}

				c.router.
					Handle(end.URL, handler).
					Methods(method)
			default:
				return errors.New("HTTP method is not allowed: " + method)
			}
		}
	}

	return nil
}

func init() {
	// Initialization . . .
}
