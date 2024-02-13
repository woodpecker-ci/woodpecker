package setup

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func receiveTokenFromUI(c context.Context, serverURL string) (string, error) {
	port := randomPort()

	tokenReceived := make(chan string)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	srv.Handler = setupRouter(tokenReceived)

	go func() {
		log.Debug().Msgf("Listening for token response on :%d", port)
		_ = srv.ListenAndServe()
	}()

	defer func() {
		log.Debug().Msg("Shutting down server")
		_ = srv.Shutdown(c)
	}()

	err := openBrowser(fmt.Sprintf("%s/cli/auth", serverURL))
	if err != nil {
		return "", err
	}

	// wait for token to be received or timeout
	select {
	case token := <-tokenReceived:
		return token, nil
	case <-c.Done():
		return "", c.Err()
	case <-time.After(5 * time.Minute):
		return "", errors.New("timed out waiting for token")
	}
}

func setupRouter(tokenReceived chan string) *gin.Engine {
	e := gin.New()
	e.UseRawPath = true
	e.Use(gin.Recovery())

	e.Use(func(c *gin.Context) {
		if c.Request.Method != "OPTIONS" {
			c.Next()
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "authorization, origin, content-type, accept")
			c.Header("Allow", "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS")
			c.Header("Content-Type", "application/json")
			c.AbortWithStatus(200)
		}
	})

	e.POST("/token", func(c *gin.Context) {
		token := c.PostForm("token")
		tokenReceived <- token

		c.JSON(200, gin.H{
			"ok": "true",
		})
	})

	return e
}

func openBrowser(url string) error {
	var err error

	log.Debug().Msgf("Opening browser with URL: %s", url)

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

func randomPort() int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(10000) + 20000
}
