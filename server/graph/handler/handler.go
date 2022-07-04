package handler

import (
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/woodpecker-ci/woodpecker/server/graph/directives"
	"github.com/woodpecker-ci/woodpecker/server/graph/generated"
	"github.com/woodpecker-ci/woodpecker/server/graph/resolvers"
)

// Defining the Graphql handler
func GraphqlHandler() gin.HandlerFunc {
	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in the resolver.go file
	g := generated.Config{Resolvers: &resolvers.Resolver{}}
	g.Directives.IsAuthenticated = directives.IsAuthenticated
	h := handler.NewDefaultServer(generated.NewExecutableSchema(g))
	h.AddTransport(transport.POST{})
	h.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		// TODO: provide user context from cookie
		// InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, error) {
		// 	userId, err := validateAndGetUserID(payload["token"])
		// 	if err != nil {
		// 		return nil, err
		// 	}

		// 	// get the user from the database
		// 	user := getUserByID(db, userId)

		// 	// put it in context
		// 	userCtx := context.WithValue(r.Context(), userCtxKey, user)

		// 	// and return it so the resolvers can see it
		// 	return userCtx, nil
		// },
	})
	h.Use(extension.Introspection{})

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// Defining the Playground handler
func PlaygroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/api/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
