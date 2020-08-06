package main

import (
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
	ctx    context.Context
	cancel context.CancelFunc
}

// Result records some state about the processing started by a client request
type Result struct {
	Status string `json:"status"`
	Result string `json:"result"`

	StartTime  time.Time `json:"startTime"`
	FinishTime time.Time `json:"finishedTime"`
}

// ResultsAggregator combines and formats the total state for processing on
// behalf of all client requests
func ResultsAggregator(agg []*Result) func(key, val interface{}) bool {
	return func(key interface{}, val interface{}) bool {

		var ok bool
		var err error

		var r *Result
		if r, ok = val.(*Result); !ok {
			err = fmt.Errorf("failed to read job record with id %v", key)

			r.Status = "FAULTY"
			r.Result = err.Error()
		}

		agg = append(agg, r)

		return true
	}
}

// JobsPool manages resources for processing client requests
var JobsPool sync.Pool

// WorkTable is a global thread safe map for storing controls for clients to
// manage the data processing
var WorkTable *sync.Map

// ResultsTable is a global record of all the state of processing started by a client requests
var ResultsTable *sync.Map

// JobIDPattern is the regular expression for getting a job ID from the path
// "^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"
var JobIDPattern = regexp.MustCompile("[0-9]{4}")

// InfoHandler hanldes requests to the /info path
func InfoHandler(w http.ResponseWriter, r *http.Request) {
	curTime := time.Now().Format(time.Kitchen)
	w.Write([]byte(fmt.Sprintf("the current time is %v", curTime)))
}

// JobHandler responsible for changing an individual Job specified by an ID
func JobHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:

		var ok bool
		var err error

		fmt.Printf("URL for job: %s \n", r.URL.Path)
		fmt.Printf("Job ID: %s \n", JobIDPattern.FindString(r.URL.Path))

		var jobID uint64
		if jobID, err = strconv.ParseUint(JobIDPattern.FindString(r.URL.Path), 10, 32); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{ \"error\": %s }", err.Error())))

			return
		}

		var val interface{}
		if val, ok = ResultsTable.Load(jobID); !ok {
			err = fmt.Errorf("job with id %d not found", jobID)

			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("{ \"error\": %s }", err.Error())))

			return
		}

		var r *Result
		if r, ok = val.(*Result); !ok {
			err = fmt.Errorf("failed to read job record with id %d", jobID)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{ \"error\": %s }", err.Error())))

			return
		}

		if json.NewEncoder(w).Encode(r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{ \"error\": %s }", err.Error())))

			return
		}

		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:

		var ok bool
		var err error

		var jobID uint64
		if jobID, err = strconv.ParseUint(JobIDPattern.FindString(r.URL.Path), 10, 32); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{ \"error\": %s }", err.Error())))

			return
		}

		var val interface{}
		if val, ok = WorkTable.Load(jobID); !ok {
			err = fmt.Errorf("job with id %d not found", jobID)

			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("{ \"error\": %s }", err.Error())))

			return
		}

		var j *Job
		if j, ok = val.(*Job); !ok {
			err = fmt.Errorf("failed to read job record with id %d", jobID)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{ \"error\": %s }", err.Error())))

			return
		}

		j.cancel()
		WorkTable.Delete(jobID)

		if json.NewEncoder(w).Encode(fmt.Sprintf("{ \"message\": job with ID %d cancelled at %v }", jobID, time.Now())); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{ \"error\": %s }", err.Error())))

			return
		}

		w.WriteHeader(http.StatusOK)

	default:
		// TODO: Handle unsupported HTTP verbs

	}
}

// JobsHandler adds new Jobs and retrieves the aggregate state of all the
// processing
func JobsHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:

		var err error

		var response []*Result
		ResultsTable.Range(ResultsAggregator(response))
		// ResultsTable.Range(func(key interface{}, val interface{}) bool {

		// 	var ok bool
		// 	var err error

		// 	var r *Result
		// 	if r, ok = val.(*Result); !ok {
		// 		err = fmt.Errorf("failed to read job record with id %v", key)

		// 		r.Status = "FAULTY"
		// 		r.Result = err.Error()
		// 	}

		// 	response = append(response, r)

		// 	return true
		// })

		if json.NewEncoder(w).Encode(response); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{ \"error\": %s }", err.Error())))

			return
		}

		w.WriteHeader(http.StatusOK)

	case http.MethodPost:

		var ok bool
		var err error

		var payload struct {
			Message string `json:"message"`
		}

		if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
			fmt.Printf("decoding failed with error: %+v \n", err)
		}

		var j *Job
		if j, ok = JobsPool.Get().(*Job); !ok {
			err = fmt.Errorf("pool element cast success: %t", ok)
			fmt.Printf("type cast failed with error: %+v \n", err)
		}

		// var jobID = uuid.New()
		var jobID = rand.Int63n(9999)

		var r = new(Result)

		r.Status = "PENDING"
		r.StartTime = time.Now()

		WorkTable.Store(jobID, j)
		ResultsTable.Store(jobID, r)

		var fail = make(chan error, 1)
		var success = make(chan interface{}, 1)

		// var t = time.NewTimer(time.Duration(rand.Int63n(8)) * time.Second)

		// go func(t *time.Timer, m string) {
		// 	<-t.C
		// 	success <- m
		// }(t, payload.Message)

		go func(r *Result, j *Job, succes chan interface{}, fail chan error) {

			var e error
			var s, c interface{}

			select {
			case s = <-success: // The task completed successfully
				fmt.Printf("job succeeded with message: %s \n", s)
				// return s, nil
			case e = <-fail: // The task failed
				fmt.Printf("job failed with error: %+v \n", e)
				// return nil, f
			case c = <-j.ctx.Done(): // The task timed out
				fmt.Printf("job cancelled %+v with error: %+v \n", c, j.ctx.Err())
				// return nil, ctx.Err()
			}

			// TODO: Make sure to put the job back in the pool!!
		}(r, j, success, fail)

		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(fmt.Sprintf("{ \"jobID\": %d }", jobID)))

	default:
		// TODO: Handle unsupported HTTP verbs

	}
}

// ExecTimer prints how long it takes for a handler to execute
type ExecTimer struct {
	handler http.Handler
}

// ServeHTTP handles the request by passing it to the wrapped handler while
// making the needed prints
func (e *ExecTimer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("started %s %s at %s \n", r.Method, r.URL.Path, time.Now())
	e.handler.ServeHTTP(w, r)
	fmt.Printf("finished %s %s at %s \n", r.Method, r.URL.Path, time.Now())
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
			w.Write([]byte(fmt.Sprintf("{ \"error\": %s }", err.Error())))

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
			w.Write([]byte(fmt.Sprintf("{ \"error\": %s }", err.Error())))

			return
		}

		h.ServeHTTP(w, r)
	})
}

func main() {

	var err error

	// https://stackoverflow.com/questions/30474313/how-to-use-regexp-get-url-pattern-in-golang
	var mux = http.NewServeMux()

	mux.HandleFunc("/info", InfoHandler)

	// Adapter pattern: https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81
	mux.HandleFunc("/jobs", JobsHandler)
	mux.HandleFunc("/jobs/", JobHandler)

	WorkTable = new(sync.Map)
	ResultsTable = new(sync.Map)

	JobsPool = sync.Pool{
		New: func() interface{} {

			// https://play.golang.org/p/SfYFNZGzShR

			var j = new(Job)

			// j.ctx, j.cancel = context.WithCancel(context.Background())
			j.ctx, j.cancel = context.WithTimeout(context.Background() /*4*time.Second*/, 4*time.Minute)

			return j
		},
	}

	var gracefulServer = http.Server{
		Handler: mux,
	}

	var listener net.Listener
	if listener, err = net.Listen("tcp", ":8080"); err != nil {
		fmt.Printf("error creating listener: %s \n", err.Error())

		return
	}

	var done = make(chan bool, 1)
	var sigs = make(chan os.Signal, 1)

	// QUESTION: Should we add build tags for system calls: https://youtu.be/PAAkCSZUG1c?t=672 ?
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func(signal chan os.Signal, done chan bool) {

		<-sigs
		var ctx, cancel = context.WithTimeout(context.Background(), time.Minute)

		defer cancel()
		gracefulServer.Shutdown(ctx)

		done <- true
	}(sigs, done)

	fmt.Printf("Server is listening on port 80 \n")
	if err = gracefulServer.Serve(listener); err != nil {
		fmt.Printf("error starting server %+v \n", err)
	}

	<-done
	fmt.Printf("Server shutting down, good bye :) \n")
}
