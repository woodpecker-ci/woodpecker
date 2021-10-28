package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	prefix    = "-- name: "
	comment   = "--"
	newline   = "\n"
	delimiter = ";"
)

// Statement represents a statment in the sql file.
type Statement struct {
	Name   string
	Value  string
	Driver string
}

// Parser parses the sql file.
type Parser struct {
	prefix string
}

// New returns a new parser.
func New() *Parser {
	return NewPrefix(prefix)
}

// NewPrefix returns a new parser with the given prefix.
func NewPrefix(prefix string) *Parser {
	return &Parser{prefix: prefix}
}

// ParseFile parses the sql file.
func (p *Parser) ParseFile(filepath string) ([]*Statement, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return p.Parse(f)
}

// Parse parses the sql file and returns a list of statements.
func (p *Parser) Parse(r io.Reader) ([]*Statement, error) {
	var (
		stmts []*Statement
		stmt  *Statement
	)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, prefix) {
			stmt = new(Statement)
			stmt.Name, stmt.Driver = parsePrefix(line, p.prefix)
			stmts = append(stmts, stmt)
		}
		if strings.HasPrefix(line, comment) {
			continue
		}
		if stmt != nil {
			stmt.Value += line + newline
		}
	}
	for _, stmt := range stmts {
		stmt.Value = strings.TrimSpace(stmt.Value)
	}
	return stmts, nil
}

func parsePrefix(line, prefix string) (name string, driver string) {
	line = strings.TrimPrefix(line, prefix)
	line = strings.TrimSpace(line)
	fmt.Sscanln(line, &name, &driver)
	return
}
