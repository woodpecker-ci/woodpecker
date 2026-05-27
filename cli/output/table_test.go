// Copyright 2024 Woodpecker Authors
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

package output

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type writerFlusherStub struct {
	bytes.Buffer
}

func (s writerFlusherStub) Flush() error {
	return nil
}

type testFieldsStruct struct {
	Name   string
	Number int
	Bool   bool
}

func TestTableOutput(t *testing.T) {
	var wfs writerFlusherStub
	to := NewTable(os.Stdout)
	to.w = &wfs

	t.Run("AddAllowedFields", func(t *testing.T) {
		_, _ = to.AddAllowedFields(testFieldsStruct{})
		_, ok := to.allowedFields["name"]
		assert.True(t, ok)
	})
	t.Run("AddFieldAlias", func(t *testing.T) {
		to.AddFieldAlias("WoodpeckerCI", "wp")
		alias, ok := to.fieldAlias["wp"]
		assert.True(t, ok)
		assert.Equal(t, "WoodpeckerCI", alias)
	})
	t.Run("AddFieldOutputFn", func(t *testing.T) {
		to.AddFieldFn("WoodpeckerCI", FieldFn(func(_ any) string {
			return "WOODPECKER CI!!!"
		}))
		_, ok := to.fieldMapping["woodpeckerci"]
		assert.True(t, ok)
	})
	t.Run("ValidateColumns", func(t *testing.T) {
		err := to.ValidateColumns([]string{"non-existent", "NAME"})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "non-existent")
		assert.NotContains(t, err.Error(), "name")

		assert.NoError(t, to.ValidateColumns([]string{"name"}))
	})
	t.Run("WriteHeader", func(t *testing.T) {
		to.WriteHeader([]string{"wp", "name"})
		assert.Equal(t, "WP\tNAME\n", wfs.String())
		wfs.Reset()
	})
	t.Run("WriteLine", func(t *testing.T) {
		err := to.Write([]string{"wp", "name", "number", "bool"}, &testFieldsStruct{"test123", 1000000000, true})
		assert.NoError(t, err)
		err = to.Write([]string{"wp", "name", "number", "bool"}, &testFieldsStruct{"", 1000000000, false})
		assert.NoError(t, err)
		assert.Equal(t, "WOODPECKER CI!!!\ttest123\t1000000000\tyes\nWOODPECKER CI!!!\t-\t1000000000\tno\n", wfs.String())
		wfs.Reset()
	})
	t.Run("Columns", func(t *testing.T) {
		assert.Len(t, to.Columns(), 4)
	})
}
