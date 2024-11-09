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

	"github.com/charmbracelet/huh/spinner"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func receiveTokenFromUI(c context.Context, serverURL string) (string, error) {
	port := randomPort()

	tokenReceived := make(chan string)

	srv := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", port)}
	srv.Handler = setupRouter(tokenReceived)

	go func() {
		log.Debug().Msgf("Listening for token response on :%d", port)
		_ = srv.ListenAndServe()
	}()

	defer func() {
		log.Debug().Msg("Shutting down server")
		_ = srv.Shutdown(c)
	}()

	err := openBrowser(fmt.Sprintf("%s/cli/auth?port=%d", serverURL, port))
	if err != nil {
		return "", err
	}

	spinnerCtx, spinnerDone := context.WithCancelCause(c)
	go func() {
		err = spinner.New().
			Title("Waiting for token ...").
			Context(spinnerCtx).
			Run()
		if err != nil {
			return
		}
	}()

	// wait for token to be received or timeout
	select {
	case token := <-tokenReceived:
		spinnerDone(nil)
		return token, nil
	case <-c.Done():
		spinnerDone(nil)
		return "", c.Err()
	case <-time.After(5 * time.Minute):
		spinnerDone(nil)
		return "", errors.New("timed out waiting for token")
	}
}

func setupRouter(tokenReceived chan string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.UseRawPath = true
	e.Use(gin.Recovery())

	e.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	e.POST("/token", func(c *gin.Context) {
		data := struct {
			Token string `json:"token"`
		}{}

		err := c.BindJSON(&data)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to bind JSON")
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		tokenReceived <- data.Token

		c.JSON(http.StatusOK, gin.H{
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
	const minPort = 10000
	const maxPort = 65535

	source := rand.NewSource(time.Now().UnixNano())
	rand := rand.New(source)
	return rand.Intn(maxPort-minPort+1) + minPort
}
