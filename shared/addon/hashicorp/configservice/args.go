package configservice

import (
	"time"

	forgetypes "go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type arguments struct {
	Repo            *model.Repo            `json:"repo"`
	Pipeline        *model.Pipeline        `json:"pipeline"`
	CurrentFileMeta []*forgetypes.FileMeta `json:"current_file_meta"`
	Netrc           *model.Netrc           `json:"netrc"`
	Timeout         time.Duration          `json:"timeout"`
}

type response struct {
	ConfigData []*forgetypes.FileMeta `json:"config_data"`
	UseOld     bool                   `json:"use_old"`
}
