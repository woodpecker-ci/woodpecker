package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/franela/goblin"
	"golang.org/x/oauth2"
)

func TestCurrentUser(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`tal@netic.dk`))
	}))

	g := goblin.Goblin(t)
	g.Describe("Bitbucket Current User", func() {
		g.After(func() {
			s.Close()
		})
		g.It("should return current user id", func() {
			ctx := context.Background()
			ts := mockSource("bearer-token")
			client := NewClientWithToken(ctx, ts, s.URL)
			uid, err := client.FindCurrentUser(ctx)
			g.Assert(err).IsNil()
			g.Assert(uid).Equal("tal_netic.dk")
		})
	})
}

type mockSource string

func (ds mockSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: string(ds)}, nil
}
