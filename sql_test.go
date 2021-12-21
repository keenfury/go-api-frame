package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSqlSingleTable(t *testing.T) {
	sql := Sql{}
	sql.RawSql = append(sql.RawSql, "create table user ()")
	tableName := ""
	columns := []Column{}
	err, _ := sql.ParseLine(&tableName, &columns)
	assert.Nil(t, err, "Should have no error")
	assert.Equal(t, "user", tableName)
}

func TestParseSqlSchemaTable(t *testing.T) {
	sql := Sql{}
	sql.RawSql = append(sql.RawSql, "create table schema.user ()")
	tableName := ""
	columns := []Column{}
	err, _ := sql.ParseLine(&tableName, &columns)
	assert.Nil(t, err, "Should have no error")
	assert.Equal(t, "user", tableName)
}

func TestParseSqlExistsSchemaTable(t *testing.T) {
	sql := Sql{}
	sql.RawSql = append(sql.RawSql, "create table if not exists schema.user ()")
	tableName := ""
	columns := []Column{}
	err, _ := sql.ParseLine(&tableName, &columns)
	assert.Nil(t, err, "Should have no error")
	assert.Equal(t, "user", tableName)
}

func TestParseLineWithColumn(t *testing.T) {
	s := Sql{RawSql: []string{"create table test (", "id serial primary key", ")"}}
	columns := []Column{}
	tableName := ""
	err, _ := s.ParseLine(&tableName, &columns)
	assert.Nil(t, err, "No error from ParseLine")
	assert.Equal(t, "test", tableName, "table name should be test")
	assert.Equal(t, 1, len(columns), "column count should be 1")
	assert.Equal(t, true, columns[0].PrimaryKey, "first column should be primary")
}

func TestParseLineWithColumnDoubleParen(t *testing.T) {
	s := Sql{RawSql: []string{"create table test (", "id serial primary key,", "name varchar(10))"}}
	columns := []Column{}
	tableName := ""
	err, _ := s.ParseLine(&tableName, &columns)
	assert.Nil(t, err, "No error from ParseLine")
	assert.Equal(t, "test", tableName, "table name should be test")
	assert.Equal(t, 2, len(columns), "column count should be 2")
	assert.Equal(t, true, columns[0].PrimaryKey, "first column should be primary")
}

func TestParseLineWithColumnDoubleParenSemiColon(t *testing.T) {
	s := Sql{RawSql: []string{"create table test (", "id serial primary key,", "name varchar(10));"}}
	columns := []Column{}
	tableName := ""
	err, _ := s.ParseLine(&tableName, &columns)
	assert.Nil(t, err, "No error from ParseLine")
	assert.Equal(t, "test", tableName, "table name should be test")
	assert.Equal(t, 2, len(columns), "column count should be 2")
	assert.Equal(t, true, columns[0].PrimaryKey, "first column should be primary")
}

func TestParseLineWithColumns(t *testing.T) {
	s := Sql{RawSql: []string{"create table test (",
		"id serial,",
		"first_name varchar(50),",
		"age int not null,",
		"active boolean not null default true,",
		"primary key(id)",
		")"}}
	columns := []Column{}
	tableName := ""
	err, _ := s.ParseLine(&tableName, &columns)
	assert.Nil(t, err, "No error from ParseLine")
	assert.Equal(t, "test", tableName, "table name should be test")
	assert.Equal(t, 4, len(columns), "column count should be 4")
	assert.Equal(t, true, columns[0].PrimaryKey, "first column should be primary")
}

func TestParseLineWithColumnsParen(t *testing.T) {
	s := Sql{RawSql: []string{"create table customer (",
		"id serial,",
		"first_name varchar(100) not null",
		")"}}
	columns := []Column{}
	tableName := ""
	err, _ := s.ParseLine(&tableName, &columns)
	assert.Nil(t, err, "No error from ParseLine")
	assert.Equal(t, "customer", tableName, "table name should be customer")
	assert.Equal(t, 2, len(columns), "column count should be 2")
	assert.Equal(t, int64(100), columns[1].Length, "second column should have length of 100")
}

func TestParseLineWithPrimaryKeyLine(t *testing.T) {
	s := Sql{RawSql: []string{"create table test (", "id serial,", "primary key(id)", ")"}}
	columns := []Column{}
	tableName := ""
	err, _ := s.ParseLine(&tableName, &columns)
	assert.Nil(t, err, "No error from ParseLine")
	assert.Equal(t, "test", tableName, "table name should be test")
	assert.Equal(t, 1, len(columns), "column count should be 1")
	assert.Equal(t, true, columns[0].PrimaryKey, "first column should be primary")
}

func TestParseColumnSimple(t *testing.T) {
	colStr := "first_name varchar(50)"
	column := Column{}
	err := ParseColumn(colStr, &column)
	assert.Nil(t, err, "No error from simple name/column")
	assert.Equal(t, "first_name", column.ColumnName.RawName, "Column name of first_name is not correct")
	assert.Equal(t, "null.String", column.GoType, "Go type for first_name should be 'string'")
	assert.Equal(t, int64(50), column.Length, "Length of DB varchar should be 50")
}

func TestParseColumnWithNotNull(t *testing.T) {
	colStr := "first_name char(50) not null"
	column := Column{}
	err := ParseColumn(colStr, &column)
	assert.Nil(t, err, "No error from with not null")
	assert.Equal(t, "first_name", column.ColumnName.RawName, "Column name of first_name is not correct")
	assert.Equal(t, "null.String", column.GoType, "Go type for first_name should be 'string'")
	assert.Equal(t, int64(50), column.Length, "Length of DB varchar should be 50")
	assert.Equal(t, false, column.Null)
}

func TestParseColumnWithNull(t *testing.T) {
	colStr := "age int null"
	column := Column{}
	err := ParseColumn(colStr, &column)
	assert.Nil(t, err, "No error from with null")
	assert.Equal(t, "age", column.ColumnName.RawName, "Column name of age is not correct")
	assert.Equal(t, "null.Int", column.GoType, "Go type for age should be 'int'")
	assert.Equal(t, true, column.Null)
}

func TestParseColumnWithDefault(t *testing.T) {
	colStr := "active boolean not null default true"
	column := Column{}
	err := ParseColumn(colStr, &column)
	assert.Nil(t, err, "No error from with default")
	assert.Equal(t, "active", column.ColumnName.RawName, "Column name of active is not correct")
	assert.Equal(t, "null.Bool", column.GoType, "Go type for active should be 'bool'")
	assert.Equal(t, false, column.Null)
	assert.Equal(t, "true", column.DefaultValue)
}

func TestParseColumnWithPrimary(t *testing.T) {
	colStr := "id serial primary key"
	column := Column{}
	err := ParseColumn(colStr, &column)
	assert.Nil(t, err, "No error from with primary")
	assert.Equal(t, "id", column.ColumnName.RawName, "Column name of id is not correct")
	assert.Equal(t, "int", column.GoType, "Go type for id should be 'int'")
	assert.Equal(t, true, column.PrimaryKey)
}

func TestParseColumnSqliteAutoIncrement(t *testing.T) {
	colStr := "id INTEGER PRIMARY KEY AUTOINCREMENT"
	column := Column{}
	err := ParseColumn(colStr, &column)
	assert.Nil(t, err, "No error from sqlite auto increment")
	assert.Equal(t, "id", column.ColumnName.RawName, "Column name of id is not correct")
	assert.Equal(t, "int", column.GoType, "Go type for id should be 'int'")
	assert.Equal(t, "autoincrement", column.DBType, "DB type should be 'autoincrement'")
	assert.Equal(t, true, column.PrimaryKey)
}
