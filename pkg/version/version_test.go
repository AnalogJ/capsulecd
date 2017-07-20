package version_test

import (
	"testing"
	"github.com/Masterminds/semver"
	"capsulecd/pkg/version"
	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	t.Parallel()

	//test
	v, nerr := semver.NewVersion(version.VERSION)

	//assert
	assert.NoError(t, nerr, "should be a valid semver")
	assert.Equal(t, version.VERSION, v.String())
}
