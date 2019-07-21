package version_test

import (
	"github.com/analogj/capsulecd/pkg/version"
	"github.com/Masterminds/semver"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVersion(t *testing.T) {
	t.Parallel()

	//test
	v, nerr := semver.NewVersion(version.VERSION)

	//assert
	require.NoError(t, nerr, "should be a valid semver")
	require.Equal(t, version.VERSION, v.String())
}
