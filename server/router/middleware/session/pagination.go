package session

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

const (
	defaultPage    = 1
	defaultPerPage = 25
)

func Pagination(c *gin.Context) *model.PaginationData {
	page, err := strconv.ParseInt(c.Query("page"), 10, 64)
	if err != nil {
		page = defaultPage
	}
	perPage, err := strconv.ParseInt(c.Query("perPage"), 10, 64)
	if err != nil {
		perPage = defaultPerPage
	}
	return &model.PaginationData{
		Page:    page,
		PerPage: perPage,
	}
}
