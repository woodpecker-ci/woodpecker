// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpc

import (
	"context"
	"time"

	"go.woodpecker-ci.org/woodpecker/pipeline/rpc/proto"

	"google.golang.org/grpc"
)

type AuthClient struct {
	client     proto.WoodpeckerAuthClient
	conn       *grpc.ClientConn
	agentToken string
	agentID    int64
}

func NewAuthGrpcClient(conn *grpc.ClientConn, agentToken string, agentID int64) *AuthClient {
	client := new(AuthClient)
	client.client = proto.NewWoodpeckerAuthClient(conn)
	client.conn = conn
	client.agentToken = agentToken
	client.agentID = agentID
	return client
}

func (c *AuthClient) Auth() (string, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &proto.AuthRequest{
		AgentToken: c.agentToken,
		AgentId:    c.agentID,
	}

	res, err := c.client.Auth(ctx, req)
	if err != nil {
		return "", -1, err
	}

	c.agentID = res.GetAgentId()

	return res.GetAccessToken(), c.agentID, nil
}
