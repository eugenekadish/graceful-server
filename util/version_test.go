package util_test

import (
	"testing"

	"github.com/magiconair/properties/assert"
	"gitlab.ido-services.com/luxtrust/base-component/util"
)

// TODO: Write more test. Improve existing tests :)

func TestVariableDeclerations(t *testing.T) {
	assert.Equal(t, "dev", util.SoftwareVersion, "expected software version to be dev")
	assert.Equal(t, "dev", util.APIVersion, "expected api version to be dev")
	assert.Equal(t, "N/A", util.ReleaseTimestamp, "expected release time to be N/A")
	assert.Equal(t, "N/A", util.Build, "expected build timestamp to be N/A")
}
