package session

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
)

func Forge(c *gin.Context) forge.Forge {
	v, ok := c.Get("forge")
	if !ok {
		log.Error().Msg("Cannot get forge from context")
		return nil
	}
	f, ok := v.(forge.Forge)
	if !ok {
		log.Error().Msg("Cannot detect forge")
		return nil // TODO: this should not happen, either panic or return an error
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

			forge, err := server.Config.Services.Manager.ForgeFromUser(user)
			if err != nil {
				log.Debug().Err(err).Msg("Cannot get forge")
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			c.Set("forge", forge)
			c.Next()
			return
		}

		forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
		if err != nil {
			log.Debug().Err(err).Msg("Cannot get forge")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Set("forge", forge)
		c.Next()
	}
}
