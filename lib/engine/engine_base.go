package engine

import (
	"github.com/Masterminds/semver"
	"capsulecd/lib/config"
	"fmt"
	"capsulecd/lib/errors"
)

type EngineBase struct {

}

func (e *EngineBase) BumpVersion(currentVersion string) (string, error) {
	v, nerr := semver.NewVersion(currentVersion)
	if(nerr != nil) {return "", nerr}

	switch bumpType := config.GetString("engine_version_bump_type"); bumpType {
	case "major":
		return fmt.Sprintf("%d.%d.%d", v.Major() + 1, 0, 0), nil
	case "minor":
		return fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor() + 1, 0), nil
	case "patch":
		return fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch() + 1), nil
	default:
		return "", errors.Custom("Unknown version bump interval")
	}

}