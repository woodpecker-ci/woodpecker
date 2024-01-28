package environservice

import (
	"encoding/json"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type RPCServer struct {
	Impl model.EnvironService
}

func (s *RPCServer) EnvironList(args []byte, resp *[]byte) error {
	var r *model.Repo
	err := json.Unmarshal(args, r)
	if err != nil {
		return err
	}
	env, err := s.Impl.EnvironList(r)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(env)
	return err
}
