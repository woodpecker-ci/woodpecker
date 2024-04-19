// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package session

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

const maxPageSize = 50

func Pagination(c *gin.Context) *model.ListOptions {
	page, err := strconv.ParseInt(c.Query("page"), 10, 64)
	if err != nil || page < 1 {
		page = 1
	}
	perPage, err := strconv.ParseInt(c.Query("perPage"), 10, 64)
	if err != nil || perPage < 1 || perPage > maxPageSize {
		perPage = maxPageSize
	}
	return &model.ListOptions{
		Page:    int(page),
		PerPage: int(perPage),
	}
}
