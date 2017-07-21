package version_test

import (
	"capsulecd/pkg/version"
	"github.com/Masterminds/semver"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVersion(t *testing.T) {
	t.Parallel()

	//test
	v, nerr := semver.NewVersion(version.VERSION)

	//assert
	assert.NoError(t, nerr, "should be a valid semver")
	assert.Equal(t, version.VERSION, v.String())
}
