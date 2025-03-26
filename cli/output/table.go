package output

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"
	"unicode"

	"github.com/go-viper/mapstructure/v2"
)

// NewTable creates a new Table.
func NewTable(out io.Writer) *Table {
	padding := 2

	return &Table{
		w:             tabwriter.NewWriter(out, 0, 0, padding, ' ', 0),
		columns:       map[string]bool{},
		fieldMapping:  map[string]FieldFn{},
		fieldAlias:    map[string]string{},
		allowedFields: map[string]bool{},
	}
}

type FieldFn func(obj any) string

type writerFlusher interface {
	io.Writer
	Flush() error
}

// Table is a generic way to format object as a table.
type Table struct {
	w             writerFlusher
	columns       map[string]bool
	fieldMapping  map[string]FieldFn
	fieldAlias    map[string]string
	allowedFields map[string]bool
}

// Columns returns a list of known output columns.
func (o *Table) Columns() (cols []string) {
	for c := range o.columns {
		cols = append(cols, c)
	}
	sort.Strings(cols)
	return
}

// AddFieldAlias overrides the field name to allow custom column headers.
func (o *Table) AddFieldAlias(field, alias string) *Table {
	o.fieldAlias[strings.ToLower(alias)] = field
	return o
}

// AddFieldFn adds a function which handles the output of the specified field.
func (o *Table) AddFieldFn(field string, fn FieldFn) *Table {
	o.fieldMapping[strings.ToLower(field)] = fn
	o.allowedFields[strings.ToLower(field)] = true
	o.columns[strings.ToLower(field)] = true
	return o
}

// AddAllowedFields reads all first level field names of the struct and allows them to be used.
func (o *Table) AddAllowedFields(obj any) (*Table, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Struct {
		return o, fmt.Errorf("AddAllowedFields input must be a struct")
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		k := t.Field(i).Type.Kind()
		if k != reflect.Bool &&
			k != reflect.Float32 &&
			k != reflect.Float64 &&
			k != reflect.String &&
			k != reflect.Int &&
			k != reflect.Int64 {
			// only allow simple values
			// complex values need to be mapped via a FieldFn
			continue
		}
		o.allowedFields[strings.ToLower(t.Field(i).Name)] = true
		o.allowedFields[fieldName(t.Field(i).Name)] = true
		o.columns[fieldName(t.Field(i).Name)] = true
	}
	return o, nil
}

// RemoveAllowedField removes fields from the allowed list.
func (o *Table) RemoveAllowedField(fields ...string) *Table {
	for _, field := range fields {
		delete(o.allowedFields, field)
		delete(o.columns, field)
	}
	return o
}

// ValidateColumns returns an error if invalid columns are specified.
func (o *Table) ValidateColumns(cols []string) error {
	var invalidCols []string
	for _, col := range cols {
		if _, ok := o.allowedFields[strings.ToLower(col)]; !ok {
			invalidCols = append(invalidCols, col)
		}
	}
	if len(invalidCols) > 0 {
		return fmt.Errorf("invalid table columns: %s", strings.Join(invalidCols, ","))
	}
	return nil
}

// WriteHeader writes the table header.
func (o *Table) WriteHeader(columns []string) {
	var header []string
	for _, col := range columns {
		header = append(header, strings.ReplaceAll(strings.ToUpper(col), "_", " "))
	}
	_, _ = fmt.Fprintln(o.w, strings.Join(header, "\t"))
}

func (o *Table) Flush() error {
	return o.w.Flush()
}

// Write writes a table line.
func (o *Table) Write(columns []string, obj any) error {
	var data map[string]any

	if err := mapstructure.Decode(obj, &data); err != nil {
		return fmt.Errorf("failed to decode object: %w", err)
	}

	dataL := map[string]any{}
	for key, value := range data {
		dataL[strings.ToLower(key)] = value
	}

	var out []string
	for _, col := range columns {
		colName := strings.ToLower(col)
		if alias, ok := o.fieldAlias[colName]; ok {
			colName = strings.ToLower(alias)
		}
		if fn, ok := o.fieldMapping[strings.ReplaceAll(colName, "_", "")]; ok {
			out = append(out, sanitizeString(fn(obj)))
			continue
		}
		if value, ok := dataL[strings.ReplaceAll(colName, "_", "")]; ok {
			if value == nil {
				out = append(out, NA(""))
				continue
			}
			if b, ok := value.(bool); ok {
				out = append(out, YesNo(b))
				continue
			}
			if s, ok := value.(string); ok {
				out = append(out, NA(sanitizeString(s)))
				continue
			}
			out = append(out, sanitizeString(value))
		}
	}
	_, _ = fmt.Fprintln(o.w, strings.Join(out, "\t"))

	return nil
}

func NA(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func YesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func fieldName(name string) string {
	r := []rune(name)
	var out []rune
	for i := range r {
		if i > 0 && (unicode.IsUpper(r[i])) && (i+1 < len(r) && unicode.IsLower(r[i+1]) || unicode.IsLower(r[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(r[i]))
	}
	return string(out)
}

func sanitizeString(value any) string {
	str := fmt.Sprintf("%v", value)
	replacer := strings.NewReplacer("\n", " ", "\r", " ")
	return strings.TrimSpace(replacer.Replace(str))
}
