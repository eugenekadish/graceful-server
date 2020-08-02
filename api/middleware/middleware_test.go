package middleware_test

import (
	"testing"
	// "gitlab.ido-services.com/luxtrust/base-component/api/middleware"
)

// TODO: Write more test. Improve existing tests :)

type configurationStub struct{}

func (configurationStub) GetString(s string) string {
	if s == "api.public_key" {
		return `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4HA/WopDR0IlMktWiM8N
mYhCNb0pdvyLy12v9q+mH5C4wNhfb/ZE3X3ceddQFzn9TJO+GiDGbLs1dFn3WKtY
Ww6Ju36FLYc2iMpbdlsQ9eksRZwG0QfjeZXokVurKyn82INUIe5oBrqB+ADl+5xS
EkYTXlK5gvwAH1diyEuYlHq6F5+EY8SOzyrZf/pwjVO2BdxGdBfUqoKIBgKd+ZxJ
wfmyon6rgmikxU9++LSsMNBeJcYtfx2W9/6hK5rC9MiZ7qeR5jU6yMdUMNQKeJSN
Y9kK7E40YgZsI/vW/5HhhkdMMK2wd4p98n+AYXY1D3IL7dextKgD0owTvqmT08M0
vQIDAQAB
-----END PUBLIC KEY-----`
	}

	return ""
}

func (configurationStub) GetBool(s string) bool {
	return true
}

func TestMiddlewareDefinitions(t *testing.T) {
	// TODO: Create type instances and check them with reflection
}
