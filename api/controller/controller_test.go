package controller_test

import (
	"fmt"
	"net/http"

	"testing"

	"github.com/gorilla/mux"

	"gitlab.ido-services.com/luxtrust/base-component/api/controller"
)

// TODO: Write more test. Improve existing tests :)

// TODO: We need to make sure having stateful package level variables will not be a problem if the
// test functions are executed asynchronously.
var recorder []string
var handlerCallCount int

var validHttpMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

type MuxRouterStub struct {
}

func testHandlerFunc(w http.ResponseWriter, req *http.Request) {
}

var mockEndpoints = []controller.Endpoint{
	{
		URL: "/test/path",
		Methods: map[string]http.HandlerFunc{
			"GET":    testHandlerFunc,
			"POST":   testHandlerFunc,
			"PUT":    testHandlerFunc,
			"PATCH":  testHandlerFunc,
			"DELETE": testHandlerFunc,
		},
	},
}

// Handle stubs out the function called to register endpoints with the controller.
func (MuxRouterStub) Handle(path string, dummy http.Handler) *mux.Route {

	var (
		ok bool

		router *mux.Router = mux.NewRouter()
	)

	recorder = append(recorder, path)

	if _, ok = dummy.(http.HandlerFunc); !ok {
		println(fmt.Errorf("could not cast: %v to type http.HandlerFunc", dummy))
	} else {
		handlerCallCount++
	}

	return router.NewRoute()
}

func TestEndpointRegistration(t *testing.T) {

	var (
		index int
		err   error

		rtrStub MuxRouterStub
		checker controller.Endpoint

		testCtlr = controller.New(rtrStub)
	)

	// Register the endpoints
	if err = testCtlr.Register(mockEndpoints); err != nil {
		t.Error(err)
	}

	for index, checker = range mockEndpoints {
		if checker.URL != recorder[index] {
			t.Errorf("expected url: %s got %s", checker.URL, recorder[index])
		}
	}

	if handlerCallCount != len(validHttpMethods) {
		t.Fatalf("expected count: %d got %d", handlerCallCount, len(validHttpMethods))
	}
}

func TestEndpointRegistrationWithInvalidHTTPMethod(t *testing.T) {

	var (
		err error

		rtrStub  MuxRouterStub
		testCtlr *controller.Controller
	)

	mockEndpoints = []controller.Endpoint{
		{
			URL: "/test/path",
			Methods: map[string]http.HandlerFunc{
				"CONNECT": testHandlerFunc,
			},
		},
	}

	testCtlr = controller.New(rtrStub)
	if err = testCtlr.Register(mockEndpoints); err == nil {
		t.Error("expected invalid Http method to cause an error")
	}
}
