package multipart

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestReader(t *testing.T) {
	b := bytes.NewBufferString(sample)
	m := New(b)

	part, err := m.NextPart()
	if err != nil {
		t.Error(err)
		return
	}

	header := part.Header()
	if got, want := header.Get("Content-Type"), "text/plain"; got != want {
		t.Errorf("Want Content-Type %q, got %q", want, got)
	}
	body, err := ioutil.ReadAll(part)
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := string(body), sampleTextPlain; got != want {
		t.Errorf("Want body %q, got %q", want, got)
	}

	part, err = m.NextPart()
	if err != nil {
		t.Error(err)
		return
	}
	header = part.Header()
	if got, want := header.Get("Content-Type"), "application/json+coverage"; got != want {
		t.Errorf("Want Content-Type %q, got %q", want, got)
	}
	if got, want := header.Get("X-Covered"), "96.00"; got != want {
		t.Errorf("Want X-Covered %q, got %q", want, got)
	}
	if got, want := header.Get("X-Covered-Lines"), "96"; got != want {
		t.Errorf("Want X-Covered-Lines %q, got %q", want, got)
	}
	if got, want := header.Get("X-Total-Lines"), "100"; got != want {
		t.Errorf("Want X-Total-Lines %q, got %q", want, got)
	}
}

var sample = `PIPELINE
Content-Type: multipart/mixed; boundary=boundary

--boundary
Content-Type: text/plain

match: pipeline/frontend/yaml/compiler/coverage.out
match: pipeline/frontend/yaml/coverage.out
match: pipeline/frontend/yaml/linter/coverage.out

--boundary
Content-Type: application/json+coverage
X-Covered: 96.00
X-Covered-Lines: 96
X-Total-Lines: 100

{"metrics":{"covered_lines":96,"total_lines":100}}

--boundary--
`

var sampleTextPlain = `match: pipeline/frontend/yaml/compiler/coverage.out
match: pipeline/frontend/yaml/coverage.out
match: pipeline/frontend/yaml/linter/coverage.out
`
