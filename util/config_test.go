package util_test

import (
	"os"
	"testing"

	"gitlab.ido-services.com/luxtrust/base-component/util"
)

// TODO: Write more test. Improve existing tests :)

func TestLoadConfigLocal(t *testing.T) {

	var configFile = os.Getenv("GOPATH") + "/src/gitlab.ido-services.com/luxtrust/base-component/util/testdata/config.yaml"
	var config = util.LoadConfig(configFile)

	if "base-component" != config.GetString("name") {
		t.Fatalf("[fatal] expected name field of the config to be: %s got %s", "base-component", config.GetString("name"))
	}

	if "8080" != config.GetString("api.port") {
		t.Fatalf("[fatal] expected the api port specified in the config to be: %s got %s",
			"8080", config.GetString("api.port"))
	}

	if "" == config.GetString("api.version") {
		t.Fatal("[fatal] api version was not set in the config")
	}

	if "" == config.GetString("api.public_key") {
		t.Fatal("[fatal] public key was not set in the config")
	}

	if config.GetBool("logging.elasticsearch") {
		t.Fatal("[fatal] expected logging elasticsearch to not be set in the config")
	}

	if !config.GetBool("monitoring.prometheus") {
		t.Fatal("[fatal] expected monitoring to be set in the config")
	}
}

func TestLoadConfigDocker(t *testing.T) {

	configPath := os.Getenv("GOPATH") +
		"/src/gitlab.ido-services.com/luxtrust/base-component/util/testdata/docker/config.yaml"

	_ = os.Setenv("CONFIG", configPath)

	var configFile = os.Getenv("CONFIG")
	var config = util.LoadConfig(configFile)

	if "base-component" != config.GetString("name") {
		t.Fatalf("[fatal] expected name field of the docker config to be: %s got %s",
			"base-component", config.GetString("name"))
	}

	if "80" != config.GetString("api.port") {
		t.Fatalf("[fatal] expected the api port specified in the docker config to be: %s got %s",
			"8080", config.GetString("api.port"))
	}

	if "" == config.GetString("api.version") {
		t.Fatal("[fatal] api version was not set in the docker config")
	}

	if "" == config.GetString("api.public_key") {
		t.Fatal("[fatal] public key was not set in the docker config")
	}

	if !config.GetBool("logging.elasticsearch") {
		t.Fatal("[fatal] expected logging elasticsearch to not be set in the docker config")
	}

	if !config.GetBool("monitoring.prometheus") {
		t.Fatal("[fatal] expected monitoring to not be set in the docker config")
	}
}

func TestLoadConfigFailure(t *testing.T) {
	// TODO: Fix this test: https://stackoverflow.com/questions/27629380/how-to-exit-a-go-program-honoring-deferred-calls
	// BETTER: https://stackoverflow.com/questions/26225513/how-to-test-os-exit-scenarios-in-go

	// defer func() {
	// 	if r := recover(); r == nil {
	// 		t.Fatal("[fatal] code did not panic")
	// 	}
	// }()
	//
	// os.Setenv("CONFIG", "/path/to/wrong/config")
	//
	// var configFile = os.Getenv("CONFIG")
	// var config = util.LoadConfig(configFile)
	//
	// sigs := make(chan os.Signal, 1)
	//
	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	//
	//
	// if config != nil {
	// 	t.Fatal("[fatal] loaded unspecified config")
	// }
}
