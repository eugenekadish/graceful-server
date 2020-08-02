package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Job struct {
	ID uuid.UUID

	Ctx    context.Context
	Cancel context.CancelFunc
}

var jobsPool sync.Pool
var jobResults = make(map[string]struct {
	Status string
	Result string

	Duration time.Time
})

// ShutdownPeriod is the duration to wait before a blocking function should
//  abandon its work.
const ShutdownPeriod = 5 * time.Minute

// InfoHandler hanldes requests to the /info path
func InfoHandler(w http.ResponseWriter, r *http.Request) {
	curTime := time.Now().Format(time.Kitchen)
	w.Write([]byte(fmt.Sprintf("the current time is %v", curTime)))
}

// JobsHandler hanldes requests to the /jobs path
func JobsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))

	var j = jobsPool.Get().(*Job)

	var fail = make(chan error, 1)
	var success = make(chan interface{}, 1)

	go func(ctx context.Context, succes chan interface{}, fail chan error) {

		var e error
		var s, c interface{}

		select {
		case s = <-success: // The task completed successfully
			fmt.Printf("job succeeded with message: %s \n", s)
			// return s, nil
		case e = <-fail: // The task failed
			fmt.Printf("job failed with error: %+v \n", e)
			// return nil, f
		case c = <-ctx.Done(): // The task timed out
			fmt.Printf("job cancelled %+v with error: %+v \n", c, ctx.Err())
			// return nil, ctx.Err()
		}
	}(j.Ctx, success, fail)

	// TODO: defer close the channels

	// Asynchronously run the Task supplied to the Job. The Task can succeed, error or timeout. Also,
	// the client can cancel the Task before the timeout has expired by manually triggering it to
	// cancel.
	go func(m string) {
		var t = time.NewTimer( /* rand.Intn(10) */ 2 * time.Second)

		<-t.C
		success <- m

		// TODO: Call j.Cancel

		// if res, err := tsk(j, j.Params); err != nil {
		// 	fail <- err
		// } else {
		// 	success <- res
		// }
		// t.
	}("blargus")

	// TODO: Retrieve the job from the map on delete
}

func main() {
	var err error

	var mux = http.NewServeMux()

	mux.HandleFunc("/info", InfoHandler)
	mux.HandleFunc("/jobs", JobsHandler)

	// if err = http.ListenAndServe(":8080", mux); err != nil {
	// 	fmt.Printf("error starting server %+v \n", err)
	// }

	var gracefulServer = http.Server{
		Handler: mux,
	}

	var listener net.Listener
	if listener, err = net.Listen("tcp", ":8080"); err != nil {
		fmt.Printf("error creating listener: %+v \n", err)

		fmt.Println(err.Error())
	}

	var done = make(chan bool, 1)
	var sigs = make(chan os.Signal, 1)

	// QUESTION: Should we add build tags for system calls: https://youtu.be/PAAkCSZUG1c?t=672 ?
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func(signal chan os.Signal, done chan bool) {

		<-sigs
		var ctx, cancel = context.WithTimeout(context.Background(), ShutdownPeriod)

		// QUESTION: Should this call be made in conjunction with a `select` statement:
		// https://golang.org/pkg/context/#WithTimeout

		defer cancel()
		gracefulServer.Shutdown(ctx)

		done <- true
	}(sigs, done)

	// if err = gracefulServer.ListenAndServe(); err != nil {
	// 	fmt.Printf("error starting server %+v \n", err)
	// }

	var ctx context.Context
	var cancel context.CancelFunc

	ctx, _ = context.WithCancel(context.Background())

	jobsPool = sync.Pool{
		New: func() interface{} {
			return Job{
				ID:     uuid.NewV4(),
				Ctx:    ctx,
				Cancel: cancel,
			}
		},
	}

	fmt.Printf("server is listening at 80 \n")
	if err = gracefulServer.Serve(listener); err != nil {
		fmt.Printf("error starting server %+v \n", err)
	}

	// if err = http.ListenAndServe(":8080", mux); err != nil {
	// 	fmt.Printf("error starting server %+v \n", err)
	// }

	// <-done
	fmt.Printf("Server shutting down, good bye :) \n")

	// log.Printf("server is listening at %s", addr)
	// log.Fatal(http.ListenAndServe(":8080", mux))
}

// TODO: A `context` should flow through the application - https://medium.com/@cep21/how-to-correctly-use-context-context-in-go-1-7-8f2c0fafdf39

// Constants for configuring the server
// const (
// 	HeaderBytes  int           = 1 << 16
// 	ReadTimeout  time.Duration = 30 * time.Second
// 	WriteTimeout time.Duration = 30 * time.Second
// )

// var (
// 	log               logging.Logger
// 	config            util.ConfigurationManager
// 	displayInfoOnly   bool
// 	logElasticSearch  bool
// 	prometheusMonitor bool
// 	configFile        string
// 	componentName     string
// 	softwareVersion   string
// 	apiPort           string
// 	apiVersion        string
// 	logEndpoint       string
// )

// func init() {

// 	configFile = os.Getenv("CONFIG")
// 	config = util.LoadConfig(configFile)

// 	componentName = config.GetString("name")
// 	softwareVersion = fmt.Sprintf("%s@%s", util.SoftwareVersion, util.Build)
// 	apiVersion = fmt.Sprintf("%s@%s", util.APIVersion, util.Build)

// 	apiPort = config.GetString("api.port")
// 	logEndpoint = config.GetString("logging.endpoint")
// 	logElasticSearch = config.GetBool("logging.elasticsearch")
// 	prometheusMonitor = config.GetBool("monitoring.prometheus")

// 	var defaultLoggerInfo = logging.DefaultLoggerInfo{
// 		Build:           util.Build,
// 		Component:       componentName,
// 		APIVersion:      util.APIVersion,
// 		SoftwareVersion: util.SoftwareVersion,
// 	}
// 	log = logging.New(defaultLoggerInfo, "json")
// 	displayInfoOnly = len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "info")
// }

// func main() {
// 	var (
// 		endpoint    *http.Server
// 		listener    net.Listener
// 		gracefulSrv *api.GracefulServer
// 		err         error
// 	)

// 	if logElasticSearch {
// 		if err = log.LogToElasticsearch(logEndpoint); err != nil {
// 			// TODO: Clean this up in some kind of wrapper function
// 			// Could use: https://golang.org/pkg/log/#pkg-constants
// 			_, file, line, _ := runtime.Caller(1)
// 			log.
// 				WithError(err).
// 				Errorf("%s:%d %v", file, line, err)
// 		}
// 	}

// 	if prometheusMonitor {
// 		monitoring.MustRegister(monitoring.APIExecTime)

// 		// The following metrics are coming from the `taskmanager` library
// 		monitoring.MustRegister(monitoring.JobExecTime)
// 		monitoring.MustRegister(monitoring.TimeUntilStart)
// 	}

// 	if displayInfoOnly {

// 		fmt.Println(componentName, "[version|info]")
// 		fmt.Println("\t|=> API Version     :\t", apiVersion)
// 		fmt.Println("\t|=> Software Version:\t", softwareVersion)

// 		return
// 	}

// 	log.
// 		WithField("port", apiPort).
// 		WithField("type", "api").
// 		Info("starting API")

// 	if listener, err = net.Listen("tcp", ":"+apiPort); err != nil {
// 		// TODO: Clean this up in some kind of wrapper function
// 		// Could use: https://golang.org/pkg/log/#pkg-constants
// 		_, file, line, _ := runtime.Caller(1)
// 		log.
// 			WithError(err).
// 			WithField("port", apiPort).
// 			WithField("type", "api").
// 			Errorf("%s:%d %v", file, line, err)
// 	}

// 	endpoint = &http.Server{
// 		Addr:           fmt.Sprintf(":%v", apiPort),
// 		ReadTimeout:    ReadTimeout,
// 		WriteTimeout:   WriteTimeout,
// 		MaxHeaderBytes: HeaderBytes,
// 	}

// 	var sigs = make(chan os.Signal, 1)
// 	var done = make(chan bool, 1)

// 	// This is where a client will make modifications
// 	if endpoint.Handler, err = handler.Routes(config, done, &http.Client{}, log); err != nil {
// 		// TODO: Clean this up in some kind of wrapper function
// 		// Could use: https://golang.org/pkg/log/#pkg-constants
// 		_, file, line, _ := runtime.Caller(1)
// 		log.
// 			WithError(err).
// 			Errorf("[error] %s:%d %v", file, line, err)
// 	}

// 	gracefulSrv = api.NewGracefulServer(componentName, util.SoftwareVersion, util.Build,
// 		api.WithGracefulServerLogger(log),
// 		api.WithGracefulServerEndpoint(endpoint),
// 		api.WithGracefulServerListener(listener))

// 	fmt.Printf("listening on port %s ", apiPort)

// 	// QUESTION: Should we add build tags for system calls: https://youtu.be/PAAkCSZUG1c?t=672 ?
// 	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

// 	go func() {
// 		gracefulSrv.Shutdown(<-sigs, done)
// 	}()

// 	gracefulSrv.Startup()
// 	<-done

// 	log.
// 		WithField("type", "api").
// 		Info("good bye :)")
// }
