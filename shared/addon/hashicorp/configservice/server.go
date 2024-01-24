package configservice

import (
	"encoding/json"

	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/config"
)

type ExtensionRPCServer struct {
	Impl config.Extension
}

func (s *ExtensionRPCServer) FetchConfig(args []byte, resp *[]byte) error {
	var a arguments
	err := json.Unmarshal(args, &a)
	configs, useOld, err := s.Impl.FetchConfig(a.Repo, a.Pipeline, a.CurrentFileMeta, a.Netrc, a.Timeout)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(response{
		ConfigData: configs,
		UseOld:     useOld,
	})
	return err
}
