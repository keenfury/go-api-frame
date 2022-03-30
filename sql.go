package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func (s *Sql) ParseLine(tableName *string, cols *[]Column) (error, bool) {
	sql := formatSql(s.RawSql)
	openParenIdx, closeParenIdx, err := determineCorrectFormat(sql)
	if err != nil {
		return err, false
	}
	*tableName = determineTableName(sql[:openParenIdx])
	colRows := breakCols(sql[openParenIdx+1 : closeParenIdx])

	primaryKeys := []string{}
	foundNull := false
	columns := []Column{}

	for _, c := range colRows {
		colStr := strings.TrimSpace(c)
		if strings.Index(colStr, "primary") == 0 {
			primaryKeys = determinePrimaryKeyNames(c)
			continue
		}
		if strings.Index(colStr, "key ") == 0 {
			fmt.Println("key: ignore")
			continue
		}
		column := Column{}
		if errParse := ParseColumn(colStr, &column); errParse != nil {
			fmt.Printf("skipping column: %s - %s\n", colStr, errParse)
			continue
		}
		columns = append(columns, column)
		if column.Null && !foundNull {
			foundNull = true
		}
	}
	if len(primaryKeys) > 0 {
		for _, key := range primaryKeys {
			for i, column := range columns {
				if key == column.ColumnName.RawName {
					columns[i].PrimaryKey = true
					break
				}
			}
		}
	}
	*cols = append(*cols, columns...)
	return nil, foundNull
}

func ParseColumn(colStr string, column *Column) error {
	varcharMatch := regexp.MustCompile(`^varchar[\(]\d+[\)]`)
	varyingMatch := regexp.MustCompile(`^varying[\(]\d+[\)]`)
	charMatch := regexp.MustCompile(`^char[\(]\d+[\)]`)
	binaryMatch := regexp.MustCompile(`^binary[\(]\d+[\)]`)
	varbinaryMatch := regexp.MustCompile(`^binary[\(]\d+[\)]`)
	split := strings.Split(colStr, " ")
	if len(split) < 2 {
		return fmt.Errorf("Not enough column arguments")
	}
	column.ColumnName.RawName = split[0]
	column.ColumnName.NameConverter()
	// let's check for 'not' or 'default' first
	column.Null = true
	// autoIncrement = false
	for i := 2; i < len(split); i++ {
		if strings.ToLower(split[i]) == "not" {
			column.Null = false
		}
		if strings.ToLower(split[i]) == "default" {
			if len(split[i])-1 > i+1 {
				column.DefaultValue = split[i+1]
			} else {
				return fmt.Errorf("Syntax error with default word but no value")
			}
		}
		if strings.ToLower(split[i]) == "primary" {
			column.PrimaryKey = true
		}
		if strings.ToLower(split[i]) == "autoincrement" || strings.ToLower(split[i]) == "auto_increment" {
			column.DBType = "autoincrement"
			column.GoType = "int"
			column.GoTypeNonSql = "int"
			return nil
		}
	}
	toLower := strings.ToLower(split[1])
	switch {
	case strings.Contains(toLower, "int"):
		column.DBType = toLower
		column.GoType = "null.Int"
		column.GoTypeNonSql = "int"
	case toLower == "numeric":
		column.DBType = "numeric"
		column.GoType = "null.Float"
		column.GoTypeNonSql = "float64"
	case toLower == "decimal" || toLower == "dec":
		column.DBType = "decimal"
		column.GoType = "null.Float"
		column.GoTypeNonSql = "float64"
	case toLower == "float":
		column.DBType = "float"
		column.GoType = "null.Float"
		column.GoTypeNonSql = "float64"
	case toLower == "double":
		column.DBType = "double"
		column.GoType = "null.Float"
		column.GoTypeNonSql = "float64"
	case toLower == "real":
		column.DBType = "real"
		column.GoType = "null.Float"
		column.GoTypeNonSql = "float64"
	case toLower == "money":
		column.DBType = "money"
		column.GoType = "null.Float"
		column.GoTypeNonSql = "float64"
	case strings.Contains(toLower, "text"):
		column.DBType = toLower
		column.GoType = "null.String"
		column.GoTypeNonSql = "string"
	case toLower == "json":
		column.DBType = "json"
		column.GoType = "*json.RawMessage"
		column.GoTypeNonSql = "[]byte"
	case strings.Contains(toLower, "blob"):
		column.DBType = toLower
		column.GoType = "null.Byte"
		column.GoTypeNonSql = "[]byte"
	case strings.Contains(toLower, "time"):
		column.DBType = toLower
		column.GoType = "null.Time"
		column.GoTypeNonSql = "time.Time"
	case toLower == "date":
		column.DBType = "date"
		column.GoType = "null.Time"
		column.GoTypeNonSql = "time.Time"
	case toLower == "uuid":
		column.DBType = "uuid"
		column.GoType = "string"
		column.GoTypeNonSql = "string"
	case toLower == "autoincrement":
		column.DBType = "autoincrement"
		column.GoType = "int"
		column.GoTypeNonSql = "int"
	case toLower == "serial":
		column.DBType = "autoincrement"
		column.GoType = "int"
		column.GoTypeNonSql = "int"
	case toLower == "boolean" || toLower == "bool":
		column.DBType = "boolean"
		column.GoType = "null.Bool"
		column.GoTypeNonSql = "bool"
	case varcharMatch.MatchString(toLower):
		column.DBType = "varchar"
		column.GoType = "null.String"
		column.GoTypeNonSql = "string"
		length, err := SplitChar(split[1])
		if err != nil {
			return fmt.Errorf("Syntax error in getting length from varchar field: %s", split[1])
		}
		column.Length = length
	case varyingMatch.MatchString(toLower):
		column.DBType = "varying"
		column.GoType = "null.String"
		column.GoTypeNonSql = "string"
		length, err := SplitChar(split[1])
		if err != nil {
			return fmt.Errorf("Syntax error in getting length from varying field: %s", split[1])
		}
		column.Length = length
	case charMatch.MatchString(toLower):
		column.DBType = "char"
		column.GoType = "null.String"
		column.GoTypeNonSql = "string"
		length, err := SplitChar(split[1])
		if err != nil {
			return fmt.Errorf("Syntax error in getting length from char field: %s", split[1])
		}
		column.Length = length
	case binaryMatch.MatchString(toLower):
		column.DBType = "binary"
		column.GoType = "null.Byte"
		column.GoTypeNonSql = "byte"
		length, err := SplitChar(split[1])
		if err != nil {
			return fmt.Errorf("Syntax error in getting length from binary field: %s", split[1])
		}
		column.Length = length
	case varbinaryMatch.MatchString(toLower):
		column.DBType = "varbinary"
		column.GoType = "null.Byte"
		column.GoTypeNonSql = "[]byte"
		length, err := SplitChar(split[1])
		if err != nil {
			return fmt.Errorf("Syntax error in getting length from varbinary field: %s", split[1])
		}
		column.Length = length
	default:
		return fmt.Errorf("Unknown column type specified")
	}
	return nil
}

func SplitChar(strChar string) (length int64, err error) {
	paranOpenIdx := strings.Index(strChar, "(")
	paranCloseIdx := strings.Index(strChar, ")")
	if paranOpenIdx == -1 || paranCloseIdx == -1 {
		err = fmt.Errorf("Parse error for char column")
		return
	}
	lengthStr := strChar[paranOpenIdx+1 : paranCloseIdx]
	length, err = strconv.ParseInt(lengthStr, 10, 64)
	return
}

func formatSql(sqlLines []string) string {
	// join all lines
	sql := strings.Join(sqlLines, "")
	// trim ends
	sql = strings.TrimSpace(sql)
	// lowercase
	sql = strings.ToLower(sql)
	// remove (`)s
	sql = strings.ReplaceAll(sql, "`", "")
	return sql
}

func determineCorrectFormat(sql string) (openIdx, closeIdx int, err error) {
	// look for "(" => paratheseIdx
	openIdx = strings.Index(sql, "(")
	if openIdx == -1 {
		err = fmt.Errorf("Syntax Error: missing open '('")
		return
	}
	// get the position of the last ","
	lastComma := strings.LastIndex(sql, ",")
	// look for ")" for syntax completeness
	closeIdx = strings.LastIndex(sql, ")")
	if closeIdx == -1 || closeIdx < lastComma {
		err = fmt.Errorf("Syntax Error: missing closing ')'")
	}
	return
}

func determineTableName(sql string) string {
	// look for "create table" or "create table if not exists" => tableIdx
	tableIdx := strings.Index(sql, "exists")
	if tableIdx == -1 {
		tableIdx = strings.Index(sql, "table") + 5
	} else {
		tableIdx += 6
	}
	// difference between tableIdx and paratheseIdx => get table name
	tableNameTemp := sql[tableIdx:]
	tableNameTemp = strings.TrimSpace(tableNameTemp)
	schemaIdx := strings.Index(tableNameTemp, ".")
	if schemaIdx > -1 {
		tableNameTemp = tableNameTemp[schemaIdx+1:]
	}
	return tableNameTemp
}

func breakCols(sql string) (cols []string) {
	colsTemp := strings.Split(sql, ",")
	includeNext := false
	lastIndex := -1
	for _, c := range colsTemp {
		if includeNext {
			cols[lastIndex] = fmt.Sprintf("%s,%s", cols[lastIndex], c)
			includeNext = false
			continue
		}
		if strings.Contains(c, "(") && !strings.Contains(c, ")") {
			includeNext = true
		}
		lastIndex++
		cols = append(cols, strings.TrimSpace(c))
	}
	return
}

func determinePrimaryKeyNames(sql string) (keyNames []string) {
	i := strings.Index(sql, "(")
	if i >= 0 {
		j := strings.Index(sql, ")")
		if j >= 0 {
			keyStr := sql[i+1 : j]
			splitList := strings.Split(keyStr, ",")
			for _, key := range splitList {
				keyNames = append(keyNames, strings.TrimSpace(key))
			}
		}
	}
	return
}
