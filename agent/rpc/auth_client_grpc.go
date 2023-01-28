package rpc

import (
	"context"
	"time"

	"github.com/woodpecker-ci/woodpecker/pipeline/rpc/proto"

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
