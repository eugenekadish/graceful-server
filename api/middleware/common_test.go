package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.ido-services.com/luxtrust/base-component/api/middleware"
)

// TODO: Write more test. Improve existing tests :)

func TestCommonMiddleware(t *testing.T) {

	var (
		counter int

		cfgs          configurationStub
		serverHandler http.Handler

		req http.Request
		cm  *middleware.CommonMiddleware
	)

	var rec = httptest.NewRecorder()

	// Stub
	var handlerFuncStub http.HandlerFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++
	})

	cm = middleware.NewCommonMiddleware(cfgs)
	serverHandler = cm.Wrapper(handlerFuncStub)

	if 0 != counter {
		t.Fatalf("expected counter to be: %d got %d", 0, counter)
	}

	serverHandler.ServeHTTP(rec, &req) // handlerFuncStub(rec, &req)

	if 1 != counter {
		t.Fatalf("expected counter to be: %d got %d", 1, counter)
	}

	// TODO: Verify attributes on the ResponseRecorder
}
