package session

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/forge"
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
		repo := Repo(c)

		if repo == nil {
			user := User(c)
			if user == nil {
				log.Error().Msg("Needs a user or repository to load get the forge")
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			forge, err := server.Config.Services.Forge.FromUser(user)
			if err != nil {
				log.Debug().Err(err).Msg("Cannot get forge")
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			c.Set("forge", forge)
			c.Next()
			return
		}

		forge, err := server.Config.Services.Forge.FromRepo(repo)
		if err != nil {
			log.Debug().Err(err).Msg("Cannot get forge")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Set("forge", forge)
		c.Next()
	}
}
