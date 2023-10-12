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

type UpdateBuilder struct {
	table       string
	fields      []string
	returning   []string
	values      []interface{}
	id          string
	whereFields []string
	whereValues []interface{}
}

func NewUpdate(table string) *UpdateBuilder {
	return &UpdateBuilder{
		table:       table,
		fields:      []string{},
		values:      []interface{}{},
		whereValues: []interface{}{},
		whereFields: []string{},
	}
}

func (b *UpdateBuilder) ID(id string) *UpdateBuilder {
	b.id = id
	return b
}

func (b *UpdateBuilder) Value(field string, value interface{}) *UpdateBuilder {
	b.fields = append(b.fields, field)
	b.values = append(b.values, value)

	return b
}

func (b *UpdateBuilder) Returning(field string) *UpdateBuilder {
	b.returning = append(b.returning, field)

	return b
}

func (b *UpdateBuilder) Where(field string, value interface{}) *UpdateBuilder {
	b.whereFields = append(b.whereFields, field)
	b.whereValues = append(b.whereValues, value)
	return b
}

func (b UpdateBuilder) GetStatement() (string, []interface{}, error) {
	if len(b.fields) != len(b.values) {
		return "", nil, errors.New("update builder: do not have the same amount of update fields and update values.")
	}
	if len(b.whereFields) != len(b.whereValues) {
		return "", nil, errors.New("update builder: do not have the same amount of where fields and where values.")
	}

	placeholders := make([]string, len(b.values))
	var lastIndex int
	for i := range b.values {
		placeholders[i] = fmt.Sprintf("%s=$%d", b.fields[i], i+1)
		lastIndex = i + 1
	}

	var where string
	if len(b.whereFields) > 0 {
		for i := range b.whereFields {
			lastIndex++
			where += fmt.Sprintf("%s=$%d AND", b.whereFields[i], lastIndex)
		}
	}

	var returning string
	if len(b.returning) > 0 {
		returning = fmt.Sprintf("RETURNING %s", strings.Join(b.returning, ","))
	}

	vals := append(b.values, b.whereValues...)
	vals = append(vals, b.id)

	return fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s id=$%d %s;",
		b.table,
		strings.Join(placeholders, ","),
		where,
		lastIndex+1,
		returning,
	), vals, nil
}
