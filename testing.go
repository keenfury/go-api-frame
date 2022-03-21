package main

import "fmt"

type (
	PostPutTest struct {
		Name         string
		ForColumn    string
		ColumnLength int
		Failure      bool
	}

	ColumnTest struct {
		GoType string
	}
)

var (
	PostTests map[string]ColumnTest
	PutTests  map[string]ColumnTest
)

func InitializeColumnTests() {
	PostTests = make(map[string]ColumnTest)
	PutTests = make(map[string]ColumnTest)
}

func AppendColumnTest(name, goType string, justPut bool) {
	if _, ok := PutTests[name]; !ok {
		PutTests[name] = ColumnTest{GoType: goType}
	}
	if !justPut {
		if _, ok := PostTests[name]; !ok {
			PostTests[name] = ColumnTest{GoType: goType}
		}
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
	case "int":
		value := 1
		if !valid {
			value = 0
		}
		return fmt.Sprintf("%s: %d", columnName, value)
	case "null.Int":
		return fmt.Sprintf("%s: null.NewInt(1, %t)", columnName, valid)
	case "null.Float":
		return fmt.Sprintf("%s: null.NewFloat(1.0, %t)", columnName, valid)
	case "null.Time":
		return fmt.Sprintf("%s: null.NewTime(time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC), %t)", columnName, valid)
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
