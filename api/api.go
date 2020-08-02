package api

import (
	"context"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"gitlab.ido-services.com/luxtrust/luxtrust-component/util"
)

// ShutdownPeriod is the duration to wait before a blocking function should abandon its work.
const ShutdownPeriod = 5 * time.Minute

// Option provides the client a callback that is used dynamically to specify attributes for a GracefulServer.
type Option func(*GracefulServer)

// GracefulServer is a robust server that can be easily be spun up and shut down with well implemented error handeling.
type GracefulServer struct {
	logger   logging.Logger
	endpoint *http.Server
	listener net.Listener
}

// WithGracefulServerLogger creates an Option that is used for specifying the Logger for a GracefulServer.
func WithGracefulServerLogger(logger logging.Logger) Option {
	return func(gs *GracefulServer) { gs.logger = logger }
}

// WithGracefulServerEndpoint creates an Option that is used for specifying the Endpoint for a GracefulServer.
func WithGracefulServerEndpoint(endpoint *http.Server) Option {
	return func(gs *GracefulServer) { gs.endpoint = endpoint }
}

// WithGracefulServerListener creates an Option that is used for specifying the Listener for a GracefulServer.
func WithGracefulServerListener(listener net.Listener) Option {
	return func(gs *GracefulServer) { gs.listener = listener }
}

// NewGracefulServer is a variadic constructor for a GracefulServer.
func NewGracefulServer(name, version, build string, opts ...Option) *GracefulServer {

	var defaultLoggerInfo = logging.DefaultLoggerInfo{
		Build:           build,
		Component:       name,
		APIVersion:      util.APIVersion,
		SoftwareVersion: version,
	}
	var log = logging.New(defaultLoggerInfo, "json")

	var gs = &GracefulServer{
		logger:   log,
		endpoint: &http.Server{},
	}

	for _, opt := range opts {
		opt(gs)
	}

	return gs
}

// Startup starts a robust server with well defined error handeling.
func (gs *GracefulServer) Startup() {

	var err error

	var listener = gs.listener

	if err = gs.endpoint.Serve(listener); err != nil {
		// TODO: Clean this up in some kind of wrapper function
		// Could use: https://golang.org/pkg/log/#pkg-constants
		_, file, line, _ := runtime.Caller(1)
		gs.logger.
			WithError(err).
			Fatalf("%s:%d %v", file, line, err)
	}
}

// Shutdown stops the server gracefully.
func (gs *GracefulServer) Shutdown(signal os.Signal, done chan bool) {

	var err error
	var ctx, cancel = context.WithTimeout(context.Background(), ShutdownPeriod)

	// QUESTION: Should this call be made in conjunction with a `select` statement:
	// https://golang.org/pkg/context/#WithTimeout
	defer cancel()

	gs.logger.
		WithField("type", "api").
		WithField("forced_shutdown_after", ShutdownPeriod).
		WithField("signal", signal).
		Warn("initiating server shutdown")

	if err = gs.endpoint.Shutdown(ctx); err != nil {
		// TODO: Clean this up in some kind of wrapper function
		// Could use: https://golang.org/pkg/log/#pkg-constants
		_, file, line, _ := runtime.Caller(1)
		gs.logger.
			WithField("type", "api").
			WithError(err).
			Fatalf("%s:%d %v", file, line, err)
	}

	done <- true
}

func init() {
	// Initialization . . .
}
