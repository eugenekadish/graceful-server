package mock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"runtime"
	"strings"

	"gitlab.ido-services.com/luxtrust/base-component/handler/endpoint"
	"gitlab.ido-services.com/luxtrust/base-component/util"
	"gitlab.ido-services.com/luxtrust/logging"
)

var defaultLoggerInfo = logging.DefaultLoggerInfo{
	Build:           util.Build,
	Component:       "base-component",
	APIVersion:      util.APIVersion,
	SoftwareVersion: util.SoftwareVersion,
}
var log = logging.New(defaultLoggerInfo, "json")

// StubRequestHandlerFunc is a callback for handeling mocked requests
type StubRequestHandlerFunc func(*http.Request) (interface{}, error)

// ClientStub is a stub for an HTTP Client.
type ClientStub struct {
	ResMap   map[string]StubRequestHandlerFunc
	URLPaths *regexp.Regexp
}

// Do executes the mock request.
func (c *ClientStub) Do(clientReq *http.Request) (clientRes *http.Response, err error) {

	var (
		key string

		match   bool
		payload interface{}

		w bytes.Buffer

		rec          = httptest.NewRecorder()
		relativePath = clientReq.URL.EscapedPath()
	)

	clientRes = new(http.Response)

	if match = c.URLPaths.MatchString(relativePath); !match {
		return
	}

	key = fmt.Sprintf("%s:%s:%s", clientReq.Host, strings.TrimPrefix(relativePath, "/"), clientReq.Method)

	if payload, err = c.ResMap[key](clientReq); err != nil {

		// QUESTION: Logic to be set here?

		return
	}

	if err = json.NewEncoder(&w).Encode(payload); err != nil {
		_, file, line, _ := runtime.Caller(0)
		endpoint.EncodingErrorWrapper(rec, log, file, line, "encoding test response payload failed", err)

		clientRes.Body = ioutil.NopCloser(bytes.NewBuffer(rec.Body.Bytes()))

		return
	}

	clientRes.Body = ioutil.NopCloser(bytes.NewBuffer(w.Bytes()))

	return clientRes, err
}

// AppendHandler appends a callback to a domain, path, HTTP verb triple for checking that the code
// in test is making the expected calls.
func (c *ClientStub) AppendHandler(domain, path, verb string, reqHandler StubRequestHandlerFunc) {
	c.ResMap[fmt.Sprintf("%s:%s:%s", domain, path, verb)] = reqHandler
}
