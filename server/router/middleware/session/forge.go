package session

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/server/forge"
	"github.com/woodpecker-ci/woodpecker/server/forge/loader"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

func Forge(c *gin.Context) forge.Forge {
	v, ok := c.Get("forge")
	if !ok {
		return nil
	}
	f, ok := v.(forge.Forge)
	if !ok {
		return nil
	}
	return f
}

func SetForge() gin.HandlerFunc {
	return func(c *gin.Context) {
		_store := store.FromContext(c)
		repo := Repo(c)

		if repo == nil {
			log.Debug().Msg("Cannot find repository")
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		forge, err := loader.GetForge(_store, repo)
		if err != nil {
			log.Debug().Err(err).Msg("Cannot get forge")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Set("forge", forge)
		c.Next()
	}
}
