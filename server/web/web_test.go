package web

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/woodpecker-ci/woodpecker/server"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func Test_custom_file_returns_OK_and_empty_content(t *testing.T) {
	gin.SetMode(gin.TestMode)

	customFiles := []string{
		"/assets/custom.js",
		"/assets/custom.css",
	}

	for _, f := range customFiles {
		t.Run(f, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodGet, f, nil)
			request.RequestURI = f // additional required for mocking
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			router, _ := New()
			router.ServeHTTP(rr, request)

			assert.Equal(t, 200, rr.Code)
			assert.Equal(t, []byte(nil), rr.Body.Bytes())
		})
	}
}

func Test_custom_file_return_actual_content(t *testing.T) {
	gin.SetMode(gin.TestMode)

	temp, err := os.CreateTemp(os.TempDir(), "data.txt")
	assert.NoError(t, err)
	temp.Write([]byte("EXPECTED-DATA"))
	temp.Close()

	server.Config.Server.CustomJsFile = temp.Name()
	server.Config.Server.CustomCssFile = temp.Name()

	customRequestedFilesToTest := []string{
		"/assets/custom.js",
		"/assets/custom.css",
	}

	for _, f := range customRequestedFilesToTest {
		t.Run(f, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodGet, f, nil)
			request.RequestURI = f // additional required for mocking
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			router, _ := New()
			router.ServeHTTP(rr, request)

			assert.Equal(t, 200, rr.Code)
			assert.Equal(t, []byte("EXPECTED-DATA"), rr.Body.Bytes())
		})
	}
}
