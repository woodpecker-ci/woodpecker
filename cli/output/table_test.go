package output

import (
	"bytes"
	"os"
	"strings"
	"testing"
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
}

func TestTableOutput(t *testing.T) {
	var wfs writerFlusherStub
	to := NewTable(os.Stdout)
	to.w = &wfs

	t.Run("AddAllowedFields", func(t *testing.T) {
		_, _ = to.AddAllowedFields(testFieldsStruct{})
		if _, ok := to.allowedFields["name"]; !ok {
			t.Error("name should be a allowed field")
		}
	})
	t.Run("AddFieldAlias", func(t *testing.T) {
		to.AddFieldAlias("woodpecker_ci", "woodpecker ci")
		if alias, ok := to.fieldAlias["woodpecker_ci"]; !ok || alias != "woodpecker ci" {
			t.Errorf("woodpecker_ci alias should be 'woodpecker ci', is: %v", alias)
		}
	})
	t.Run("AddFieldOutputFn", func(t *testing.T) {
		to.AddFieldFn("woodpecker ci", FieldFn(func(_ any) string {
			return "WOODPECKER CI!!!"
		}))
		if _, ok := to.fieldMapping["woodpecker ci"]; !ok {
			t.Errorf("'woodpecker ci' field output fn should be set")
		}
	})
	t.Run("ValidateColumns", func(t *testing.T) {
		err := to.ValidateColumns([]string{"non-existent", "NAME"})
		if err == nil ||
			strings.Contains(err.Error(), "name") ||
			!strings.Contains(err.Error(), "non-existent") {
			t.Errorf("error should contain 'non-existent' but not 'name': %v", err)
		}
	})
	t.Run("WriteHeader", func(t *testing.T) {
		to.WriteHeader([]string{"woodpecker_ci", "name"})
		if wfs.String() != "WOODPECKER CI\tNAME\n" {
			t.Errorf("written header should be 'WOODPECKER CI\\tNAME\\n', is: %q", wfs.String())
		}
		wfs.Reset()
	})
	t.Run("WriteLine", func(t *testing.T) {
		_ = to.Write([]string{"woodpecker_ci", "name", "number"}, &testFieldsStruct{"test123", 1000000000})
		if wfs.String() != "WOODPECKER CI!!!\ttest123\t1000000000\n" {
			t.Errorf("written line should be 'WOODPECKER CI!!!\\ttest123\\t1000000000\\n', is: %q", wfs.String())
		}
		wfs.Reset()
	})
	t.Run("Columns", func(t *testing.T) {
		if len(to.Columns()) != 3 {
			t.Errorf("unexpected number of columns: %v", to.Columns())
		}
	})
}
