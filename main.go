package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"sync"
	"syscall"
	"time"

	_ "github.com/eugenekadish/graceful-server/api"
)

// Job is used to control some processing started by a client request
type Job struct {
	// ctx    context.Context
	// cancel context.CancelFunc

	Message string `json:"message"`
}

// Work is a job that is queued and can be cancelled
type Work struct {
	Job

	cancel context.CancelFunc
}

// Result records some state about the processing started by a client request
type Result struct {
	Status string `json:"status"`
	Result string `json:"result"`

	StartTime  time.Time `json:"startTime"`
	FinishTime time.Time `json:"finishedTime"`
}

// GracefulServer is a robust server with well defined error handling
type GracefulServer struct {
	s *http.Server
	l net.Listener

	context context.Context
}

// Start fires up the server
func (gs *GracefulServer) Start() error {
	return gs.s.Serve(gs.l)
}

// Stop shuts down the server
func (gs *GracefulServer) Stop() error {
	var ctx, cancel = context.WithTimeout(gs.context, time.Minute)
	defer cancel()

	return gs.s.Shutdown(ctx)
}

// Option provides the client a callback that is used dynamically to specify
// attributes for a GracefulServer
type Option func(*GracefulServer)

// NewGracefulServer is a variadic constructor for a GracefulServer
func NewGracefulServer(opts ...Option) *GracefulServer {
	var gs = new(GracefulServer)

	gs.s = new(http.Server)

	var opt Option
	for _, opt = range opts {
		opt(gs)
	}

	return gs
}

// WithServerContext creates an Option that is used for specifying the
// a context for the GracefulServer on shutdown
func WithServerContext(context context.Context) Option {
	return func(gs *GracefulServer) { gs.context = context }
}

// WithServerHandler creates an Option that is used for specifying the
// handler for the GracefulServer
func WithServerHandler(h http.Handler) Option {
	return func(gs *GracefulServer) { gs.s.Handler = h }
}

// WithServerListener creates an Option that is used for specifying the
// Listener for a GracefulServer
func WithServerListener(listener net.Listener) Option {
	return func(gs *GracefulServer) { gs.l = listener }
}

// JobsPool manages resources for processing client requests
var JobsPool *sync.Pool

// WorkTable is a global thread safe map for storing controls for clients to
// manage the data processing
var WorkTable map[string]*Work

// ResultsTable is a global record of all the state of processing started by a
// client requests
var ResultsTable map[string]*Result

// JobIDPattern is the regular expression for getting a job ID from the path
// "^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"
var JobIDPattern = regexp.MustCompile("[0-9]{4}")

// InfoHandler returns summarizing data of the current state of the system
func InfoHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	var status int

	var resBody bytes.Buffer

	switch r.Method {
	case http.MethodGet:
		var pending, success, failed int

		var res *Result
		for _, res = range ResultsTable {
			switch res.Status {
			case "PENDING":
				pending++
			case "SUCCESS":
				success++
			case "FAILED":
				failed++
			}
		}

		var summary struct {
			Info struct {
				Pending int `json:"pending"`
				Success int `json:"success"`
				Failed  int `json:"failed"`
			} `json:"info"`
			Total int `json:"total"`
		}

		summary.Info.Pending = pending
		summary.Info.Success = success

		summary.Info.Failed = failed
		summary.Total = pending + success + failed

		status = http.StatusOK

		if err = json.NewEncoder(&resBody).Encode(&summary); err != nil {
			status = http.StatusInternalServerError

			resBody.Reset()
			_, _ = resBody.Write([]byte(fmt.Sprintf("{ \"error\": \"%s\" }", err.Error())))
		}

		w.WriteHeader(status)
		_, _ = w.Write(resBody.Bytes())

	default:
		// TODO: Handle unsupported HTTP verbs
	}
}

// JobHandler responsible for changing an individual Job specified by an ID
func JobHandler(w http.ResponseWriter, r *http.Request) {

	var ok bool
	var err error

	var status int
	var resBody bytes.Buffer

	switch r.Method {
	case http.MethodGet:

		var jobID = JobIDPattern.FindString(r.URL.Path)

		var r *Result
		if r, ok = ResultsTable[jobID]; !ok {
			err = fmt.Errorf("job with id %s not found", jobID)

			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("{ \"error\": \"%s\" }", err.Error())))

			return
		}

		status = http.StatusOK

		if err = json.NewEncoder(&resBody).Encode(r); err != nil {
			status = http.StatusInternalServerError

			resBody.Reset()
			_, _ = resBody.Write([]byte(fmt.Sprintf("{ \"error\": \"%s\" }", err.Error())))
		}

		w.WriteHeader(status)
		_, _ = w.Write(resBody.Bytes())

	case http.MethodDelete:

		var jobID = JobIDPattern.FindString(r.URL.Path)

		var work *Work
		if work, ok = WorkTable[jobID]; !ok {
			err = fmt.Errorf("job with id %s not found", jobID)

			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("{ \"error\": \"%s\" }", err.Error())))

			return
		}

		work.cancel()
		delete(WorkTable, jobID)

		work.Message = fmt.Sprintf("job with ID %s cancelled at %v", jobID, time.Now())

		status = http.StatusOK

		if err = json.NewEncoder(&resBody).Encode(work); err != nil {
			status = http.StatusInternalServerError

			resBody.Reset()
			_, _ = resBody.Write([]byte(fmt.Sprintf("{ \"error\": \"%s\" }", err.Error())))
		}

		w.WriteHeader(status)
		_, _ = w.Write(resBody.Bytes())

	default:
		// TODO: Handle unsupported HTTP verbs

	}
}

// JobsHandler adds new Jobs and retjob wasrieves the aggregate state of all the
// processing
func JobsHandler(w http.ResponseWriter, r *http.Request) {

	var ok bool
	var err error

	var status int
	var resBody bytes.Buffer

	switch r.Method {
	case http.MethodGet:

		var r *Result
		var response []*Result

		for _, r = range ResultsTable {
			response = append(response, r)
		}

		status = http.StatusOK

		if json.NewEncoder(&resBody).Encode(response); err != nil {
			status = http.StatusInternalServerError

			resBody.Reset()
			_, _ = resBody.Write([]byte(fmt.Sprintf("{ \"error\": \"%s\" }", err.Error())))
		}

		w.WriteHeader(status)
		_, _ = w.Write(resBody.Bytes())

	case http.MethodPost:

		var payload struct {
			Message string `json:"message"`
		}

		if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {

			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("{ \"error\": \"%s\" }", err.Error())))

			return
		}

		var j *Job
		if j, ok = JobsPool.Get().(*Job); !ok {
			err = fmt.Errorf("pool element cast success: %t", ok)

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf("{ \"error\": \"%s\" }", err.Error())))

			return
		}

		j.Message = payload.Message

		// var jobID = uuid.New()
		var jobID = rand.Intn(9999)

		var work = new(Work)
		var result = new(Result)

		result.Status = "PENDING"
		result.StartTime = time.Now()

		var fail = make(chan error, 1)
		var success = make(chan interface{}, 1)

		var t = time.NewTimer(time.Duration(rand.Int63n(8)) * time.Second)

		var ctx context.Context
		ctx, work.cancel = context.WithTimeout(context.Background(), 4*time.Second)

		WorkTable[strconv.Itoa(jobID)] = work
		ResultsTable[strconv.Itoa(jobID)] = result

		// https://play.golang.org/p/SfYFNZGzShR

		go func(jP *sync.Pool, r *Result, j *Job, ctx context.Context, t *time.Timer, succes chan interface{}, fail chan error) {

			// var e error

			var s time.Time
			var c interface{}

			select {
			case s = <-t.C:

				fmt.Printf("job recieved at time: %v \n", s)

				r.Status = "SUCCESS"
				r.Result = j.Message
				r.FinishTime = time.Now()

				fmt.Printf("job succeeded with message: %s \n", r.Result)

			// case e = <-fail:

			// 	r.Status = "FAILED"
			// 	r.Result = fmt.Sprintf("job failed with error: %s", e.Error()) // Check NIL!!

			// 	fmt.Printf("job failed with error: %s \n", e.Error())

			case c = <-ctx.Done(): // TODO: Remove context from job and create a new on every time

				r.Status = "FAILED"
				r.Result = fmt.Sprintf("job was cancelled or timed out with error: %s", ctx.Err())

				fmt.Printf("job was cancelled or timed out for %v with error: %s \n", c, ctx.Err())
			}

			JobsPool.Put(j)
		}(JobsPool, result, j, ctx, t, success, fail)

		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(fmt.Sprintf("{ \"jobID\": %d }", jobID)))

	default:
		// TODO: Handle unsupported HTTP verbs

	}
}

// ExecTimer prints how long it takes for a handler to execute
type ExecTimer struct {
	handler http.Handler
}

// ServeHTTP handles the request by passing it to the wrapped handler and prints
// to the console
func (e *ExecTimer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("started %s %s at %s \n", r.Method, r.URL.Path, time.Now())

	defer func(method, path string) {
		fmt.Printf("finished %s %s at %s \n", method, path, time.Now())
	}(r.Method, r.URL.Path)

	e.handler.ServeHTTP(w, r)
}

// NewExecTimer constructs a new ExecTimer middleware handler
func NewExecTimer(handler http.Handler) *ExecTimer {
	return &ExecTimer{handler}
}

// CheeseHeaderWrapper is used as a middleware to protect a specified handler to
// have the value "CHEESE" on the "Token" header key
func CheeseHeaderWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var err error
		var token string

		if token = r.Header.Get("Token"); token != "CHEESE" {
			err = fmt.Errorf("expecting CHEESE as the Token header but got %s", token)

			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(fmt.Sprintf("{ \"error\": \"%s\" }", err.Error())))

			return
		}

		h.ServeHTTP(w, r)
	})
}

// URLPathCheckWrapper checks the URL for the handler is formatted correctly an
// has a regular expression matching ID
func URLPathCheckWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var err error
		var match bool

		if match = JobIDPattern.MatchString(r.URL.Path); !match {
			err = fmt.Errorf("URL path %s does not have the right pattern", r.URL.Path)

			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(fmt.Sprintf("{ \"error\": \"%s\" }", err.Error())))

			return
		}

		h.ServeHTTP(w, r)
	})
}

// ResponseHeaderWrapper checks the URL for the handler is formatted correctly an
// has a regular expression matching ID
func ResponseHeaderWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		h.ServeHTTP(w, r)
	})
}

func main() {

	var err error

	WorkTable = make(map[string]*Work)
	ResultsTable = make(map[string]*Result)

	JobsPool = &sync.Pool{
		New: func() interface{} {
			return new(Job)
		},
	}

	var mux = http.NewServeMux()

	mux.Handle("/info", ResponseHeaderWrapper(http.HandlerFunc(InfoHandler)))

	mux.Handle("/jobs", CheeseHeaderWrapper(ResponseHeaderWrapper(http.HandlerFunc(JobsHandler))))
	mux.Handle("/jobs/", URLPathCheckWrapper(ResponseHeaderWrapper(http.HandlerFunc(JobHandler))))

	var listener net.Listener
	if listener, err = net.Listen("tcp", ":80"); err != nil {
		fmt.Printf("error creating listener: %s \n", err.Error())

		return
	}
	var gracefulServer = NewGracefulServer(
		WithServerContext(context.Background()),

		WithServerHandler(mux),
		WithServerListener(listener),
	)

	var done = make(chan bool, 1)
	var sigs = make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func(signal chan os.Signal, done chan bool, err error) {

		<-sigs
		if err = gracefulServer.Stop(); err != nil {
			fmt.Printf("server shutdown failed with error: %s \n", err.Error())
		}

		done <- true
	}(sigs, done, err)

	fmt.Printf("Server is listening on port 80 \n")
	if err = gracefulServer.Start(); err != nil {
		fmt.Printf("error starting server %s \n", err.Error())
	}

	<-done
	fmt.Printf("Server shutting down, good bye :) \n")
}
