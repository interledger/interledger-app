package db

import (
	"errors"
	"fmt"
	"strings"
)

type InsertBuilder struct {
	table     string
	fields    []string
	returning []string
	values    []interface{}
}

func NewInsert(table string) *InsertBuilder {
	return &InsertBuilder{
		table:  table,
		fields: []string{},
		values: []interface{}{},
	}
}

func (b *InsertBuilder) Value(field string, value interface{}) *InsertBuilder {
	b.fields = append(b.fields, field)
	b.values = append(b.values, value)

	return b
}

func (b *InsertBuilder) Returning(field string) *InsertBuilder {
	b.returning = append(b.returning, field)

	return b
}

func (b InsertBuilder) GetStatement() (string, []interface{}, error) {
	if len(b.fields) != len(b.values) {
		return "", nil, errors.New("insert builder: do not have the same amount of insert fields and insert values.")
	}

	placeholders := make([]string, len(b.values))
	for i := range b.values {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	var returning string
	if len(b.returning) > 0 {
		returning = fmt.Sprintf("RETURNING %s", strings.Join(b.returning, ","))
	}

	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) %s;",
		b.table,
		strings.Join(b.fields, ","),
		strings.Join(placeholders, ","),
		returning,
	), b.values, nil
}
