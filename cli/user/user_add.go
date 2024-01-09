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

package user

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

var userAddCmd = &cli.Command{
	Name:      "add",
	Usage:     "adds a user",
	ArgsUsage: "<username>",
	Action:    userAdd,
}

func userAdd(c *cli.Context) error {
	login := c.Args().First()

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	user, err := client.UserPost(&woodpecker.User{Login: login})
	if err != nil {
		return err
	}
	fmt.Printf("Successfully added user %s\n", user.Login)
	return nil
}
