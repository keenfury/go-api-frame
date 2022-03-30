package main

import (
	"reflect"
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

func TestParseLineWithUpperCaseLine(t *testing.T) {
	s := Sql{RawSql: []string{"CREATE TABLE test (", "id serial,", "primary key(id)", ")"}}
	columns := []Column{}
	tableName := ""
	err, _ := s.ParseLine(&tableName, &columns)
	assert.Nil(t, err, "No error from ParseLine")
	assert.Equal(t, "test", tableName, "table name should be test")
	assert.Equal(t, 1, len(columns), "column count should be 1")
	assert.Equal(t, true, columns[0].PrimaryKey, "first column should be primary")
}

func TestParseLineWithTicks(t *testing.T) {
	s := Sql{RawSql: []string{
		"create table `test` (",
		"	`id` serial,",
		"	`key_one` int not null,",
		"	`key_two` int not null,",
		"	PRIMARY KEY (`id`),",
		"	KEY `index_test_key_one` (`key_one`),",
		"	KEY `index_test_key_two` (`key_one`,`key_two`)",
		");",
	}}
	columns := []Column{}
	tableName := ""
	err, _ := s.ParseLine(&tableName, &columns)
	assert.Nil(t, err, "No error from ParseLine")
	assert.Equal(t, "test", tableName, "table name should be test")
	assert.Equal(t, 3, len(columns), "column count should be 3")
	assert.Equal(t, true, columns[0].PrimaryKey, "first column should be primary")
}

func TestParseLineWithExtraMetaData(t *testing.T) {
	s := Sql{RawSql: []string{
		"create table test (",
		"	id serial,",
		"	key_one int not null,",
		"	key_two int not null,",
		"	PRIMARY KEY (id),",
		"	KEY index_test_key_one (key_one),",
		"	KEY index_test_key_two (key_one, key_two)",
		") ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;",
	}}
	columns := []Column{}
	tableName := ""
	err, _ := s.ParseLine(&tableName, &columns)
	assert.Nil(t, err, "No error from ParseLine")
	assert.Equal(t, "test", tableName, "table name should be test")
	assert.Equal(t, 3, len(columns), "column count should be 3")
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

func Test_formatSql(t *testing.T) {
	type args struct {
		sqlLines []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"formatSql - successful",
			args{
				[]string{
					" create table test (",
					"	id serial,",
					"	key_one int not null,",
					"	key_two int not null,",
					"	PRIMARY KEY (id),",
					"	KEY index_test_key_one (key_one),",
					"	KEY index_test_key_two (key_one, key_two)",
					") ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;  ",
				},
			},
			"create table test (	id serial,	key_one int not null,	key_two int not null,	primary key (id),	key index_test_key_one (key_one),	key index_test_key_two (key_one, key_two)) engine=innodb auto_increment=10 default charset=utf8 collate=utf8_unicode_ci;",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatSql(tt.args.sqlLines); got != tt.want {
				assert.Equal(t, tt.want, got, "formatSql")
			}
		})
	}
}

func Test_determineCorrectFormat(t *testing.T) {
	type args struct {
		sql string
	}
	tests := []struct {
		name         string
		args         args
		wantOpenIdx  int
		wantCloseIdx int
		wantErr      bool
	}{
		{
			"determineCorrectFormat - successful",
			args{
				"create table test (	id serial,	key_one int not null,	key_two int not null,	primary key (id),	key index_test_key_one (key_one),	key index_test_key_two (key_one, key_two)) engine=innodb auto_increment=10 default charset=utf8 collate=utf8_unicode_ci;",
			},
			18,
			168,
			false,
		},
		{
			"determineCorrectFormat - error",
			args{
				"create table test (	id serial,	key_one int not null,	key_two int not null,",
			},
			18,
			-1,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOpenIdx, gotCloseIdx, err := determineCorrectFormat(tt.args.sql)
			if tt.wantErr {
				assert.NotNil(t, err, "determineCorrectFormat - nil")
			}
			if !tt.wantErr {
				assert.Nil(t, err, "determineCorrectFormat - not nil")
			}
			assert.Equal(t, tt.wantOpenIdx, gotOpenIdx, "determineCorrectFormat - open idx")
			assert.Equal(t, tt.wantCloseIdx, gotCloseIdx, "determineCorrectFormat - close idx")
		})
	}
}

func Test_determineTableName(t *testing.T) {
	type args struct {
		sql string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"determineTableName - successful (no exists)",
			args{
				"create table test ",
			},
			"test",
		},
		{
			"determineTableName - successful (with exists)",
			args{
				"create table if not exists test ",
			},
			"test",
		},
		{
			"determineTableName - successful (with schema)",
			args{
				"create table schema.test ",
			},
			"test",
		},
		{
			"determineTableName - successful (empty)",
			args{
				"create table ",
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := determineTableName(tt.args.sql); got != tt.want {
				assert.Equal(t, tt.want, got, "determineTableName")
			}
		})
	}
}

func Test_breakCols(t *testing.T) {
	type args struct {
		sql string
	}
	tests := []struct {
		name     string
		args     args
		wantCols []string
	}{
		{
			"breakCols - successful (simple)",
			args{
				"	id serial,	key_one int not null,	key_two int not null",
			},
			[]string{
				"id serial",
				"key_one int not null",
				"key_two int not null",
			},
		},
		{
			"breakCols - successful (primary key)",
			args{
				"	id serial,	key_one int not null,	key_two int not null, primary key(id)",
			},
			[]string{
				"id serial",
				"key_one int not null",
				"key_two int not null",
				"primary key(id)",
			},
		},
		{
			"breakCols - successful (multiple keys)",
			args{
				"	id serial,	key_one int not null,	key_two int not null,	primary key(id),	key index_key_one (id, key_one)",
			},
			[]string{
				"id serial",
				"key_one int not null",
				"key_two int not null",
				"primary key(id)",
				"key index_key_one (id, key_one)",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotCols := breakCols(tt.args.sql); !reflect.DeepEqual(gotCols, tt.wantCols) {
				assert.Equal(t, tt.wantCols, gotCols, "breakCols")
			}
		})
	}
}

func Test_determinePrimaryKeyNames(t *testing.T) {
	type args struct {
		sql string
	}
	tests := []struct {
		name         string
		args         args
		wantKeyNames []string
	}{
		{
			"determinePrimaryKeyNames - single",
			args{
				"primary key (id)",
			},
			[]string{"id"},
		},
		{
			"determinePrimaryKeyNames - multiple",
			args{
				"primary key (id, client_id)",
			},
			[]string{"id", "client_id"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotKeyNames := determinePrimaryKeyNames(tt.args.sql); !reflect.DeepEqual(gotKeyNames, tt.wantKeyNames) {
				assert.Equal(t, tt.wantKeyNames, gotKeyNames, "determinePrimaryKeyNames")
			}
		})
	}
}
