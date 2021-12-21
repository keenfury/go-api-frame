package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func (s *Sql) ParseLine(tableName *string, cols *[]Column) (error, bool) {
	// join all lines
	sql := strings.Join(s.RawSql, "")
	// trim ends
	sql = strings.TrimSpace(sql)
	// look for "(" => paratheseIdx
	openParenIdx := strings.Index(sql, "(")
	if openParenIdx == -1 {
		return fmt.Errorf("Syntax Error: missing open '('"), false
	}
	// look for ")" for syntax completeness
	lastParenIdx := strings.LastIndex(sql, ")")
	if lastParenIdx == -1 || lastParenIdx <= len(sql)-3 {
		return fmt.Errorf("Syntax Error: missing closing ')'"), false
	}
	// look for "create table" or "create table if not exists" => tableIdx
	tableIdx := strings.Index(sql, "exists")
	if tableIdx == -1 {
		tableIdx = strings.Index(sql, "table") + 5
	} else {
		tableIdx += 6
	}
	// difference between tableIdx and paratheseIdx => get table name
	tableNameTemp := sql[tableIdx:openParenIdx]
	tableNameTemp = strings.TrimSpace(tableNameTemp)
	schemaIdx := strings.Index(tableNameTemp, ".")
	if schemaIdx > -1 {
		tableNameTemp = tableNameTemp[schemaIdx+1:]
	}
	*tableName = tableNameTemp
	// find next "," => parse sytax validity; get column name, type and other attributes
	startIdx := 0
	primaryKeyStr := ""
	columnStr := sql[openParenIdx+1:]
	columns := []Column{}
	foundEnd := false
	foundNull := false
	for {
		if foundEnd {
			break
		}
		foundColumn := true
		commaIdx := strings.Index(columnStr[startIdx:], ",")
		if commaIdx == -1 {
			endCheck := strings.LastIndex(columnStr[startIdx:], ")")
			strLen := len(columnStr[startIdx:])
			if endCheck == strLen-1 {
				// this is to test ')'
				foundEnd = true
				commaIdx = endCheck
			} else if endCheck == strLen-2 {
				// this is to test '))'
				foundEnd = true
				commaIdx = endCheck // - 1
			} else {
				foundEnd = true
			}
		}
		commaIdx += startIdx
		// look for primary key phrase for postgres and mysql equivalent
		if strings.Index(columnStr[startIdx:commaIdx], "primary key(") > -1 { // for postgres one-liner of setting primary key
			// postgres
			if !foundEnd {
				commaIdx++
			}
			primaryKeyStr = columnStr[startIdx:commaIdx]
			startIdx = commaIdx + 1
			continue
		}
		if len(columnStr[startIdx:commaIdx]) < 4 {
			// if have a "();" or "()" or ");" then break
			break
		}
		colStr := strings.TrimSpace(columnStr[startIdx:commaIdx])
		column := Column{}
		if errParse := ParseColumn(colStr, &column); errParse != nil {
			fmt.Printf("skipping column: %s - %s\n", colStr, errParse)
			foundColumn = false
		}
		if foundColumn {
			columns = append(columns, column)
			if column.Null && !foundNull {
				foundNull = true
			}
		}
		startIdx = commaIdx + 1
		if foundEnd {
			break
		}
	}
	// handle primary keys if needed
	if len(primaryKeyStr) > 0 {
		primaryParamStart := strings.Index(primaryKeyStr, "(") + 1
		primaryParamEnd := strings.Index(primaryKeyStr, ")")
		primaryList := primaryKeyStr[primaryParamStart:primaryParamEnd]
		splitList := strings.Split(primaryList, ",")
		for _, key := range splitList {
			for i, column := range columns {
				if key == column.ColumnName.RawName {
					columns[i].PrimaryKey = true
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
