// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shared

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

// Base64Decoder decodes one Base64 argument and writes its bytes to stdout.
func Base64Decoder(_ context.Context, c *cli.Command) error {
	if c.Args().Len() != 1 {
		return fmt.Errorf("expected exactly one base64 argument")
	}

	decoded, err := base64.StdEncoding.DecodeString(c.Args().First())
	if err != nil {
		return fmt.Errorf("decode base64: %w", err)
	}

	_, err = os.Stdout.Write(decoded)
	return err
}
