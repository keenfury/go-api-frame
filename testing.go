package main

import "fmt"

type (
	PostTest struct {
		Name         string
		ForColumn    string
		ColumnLength int
		Failure      bool
	}

	ColumnTest struct {
		GoType string
	}
)

var ColumnTests map[string]ColumnTest

func InitializeColumnTests() {
	ColumnTests = make(map[string]ColumnTest)
}

func AppendColumnTest(name, goType string) {
	if _, ok := ColumnTests[name]; !ok {
		ColumnTests[name] = ColumnTest{GoType: goType}
	}
}

func TranslateType(columnName, columnType string, length int, valid bool) string {
	switch columnType {
	case "null.String":
		if length > 0 {
			return fmt.Sprintf("%s: null.NewString(\"%s\", true)", columnName, buildRandomString(length))
		}
		return fmt.Sprintf("%s: null.NewString(\"a\", %t)", columnName, valid)
	case "string":
		if length > 0 {
			return fmt.Sprintf("%s: \"%s\"", columnName, buildRandomString(length))
		}
		return fmt.Sprintf("%s: \"a\"", columnName)
	case "null.Int":
		return fmt.Sprintf("%s: null.NewInt(1, %t)", columnName, valid)
	case "null.Float":
		return fmt.Sprintf("%s: null.NewFloat(1.0, %t)", columnName, valid)
	case "null.Time":
		return fmt.Sprintf("%s: null.NewTime(\"2022-01-01T00:00:00Z\", %t)", columnName, valid)
	case "null.Bool":
		return fmt.Sprintf("%s: null.NewBool(true, %t)", columnName, valid)
	default:
		fmt.Println("Missing type in TranslateType:", columnType)
	}
	return ""
}

func buildRandomString(length int) string {
	randomString := ""
	for {
		randomString += "0123456789"
		if len(randomString) > length {
			break
		}
	}
	return randomString
}
