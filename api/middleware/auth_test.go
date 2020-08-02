package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"

	"gitlab.ido-services.com/luxtrust/auth"

	"gitlab.ido-services.com/luxtrust/base-component/api/middleware"
)

// TODO: Write more test. Improve existing tests :)

func TestAuthenticationFlow(t *testing.T) {
	// TODO: Provide a test were handlerFunc wrapped by the middleware fully executes
}

func TestUserToken(t *testing.T) {

	var (
		err error

		testUser middleware.User
	)

	var testToken = &jwt.Token{
		Claims: &auth.CustomClaims{},
		Raw:    "blargus",
	}

	if err = testUser.SetToken(testToken); err != nil {
		t.Fatalf("token could not be set for: %v", testUser)
	}

	// NOTE: Careful of `nil` pointer exception
	if testToken.Raw != testUser.GetToken().Raw {
		t.Fatalf("expected testUser to have token: %v got %v", testToken.Raw, testUser.GetToken().Raw)
	}
}

func TestBlankAuthenticationToken(t *testing.T) {

	var (
		counter int

		cfgs          configurationStub
		serverHandler http.Handler

		req http.Request
		am  *middleware.AuthMiddleware
	)

	var rec = httptest.NewRecorder()

	// Stub
	var handlerFuncStub = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++
	})

	am = middleware.NewAuthMiddleware(cfgs)
	serverHandler = am.Wrapper(handlerFuncStub)

	serverHandler.ServeHTTP(rec, &req) // handlerFuncStub(rec, &req)

	if http.StatusForbidden != rec.Code {
		t.Fatalf("expected response status code to be: %d got %d", http.StatusForbidden, rec.Code)
	}

	// handlerFuncStub never gets called because the status code is a 401
	if 0 != counter {
		t.Fatalf("expected counter to be: %d got %d", 0, counter)
	}
}

func TestInvalidAuthenticationToken(t *testing.T) {
	// TODO: Provide a test for code execution where the Authenticate method on validator returns
	// false as the first response parameter: line 125
}
