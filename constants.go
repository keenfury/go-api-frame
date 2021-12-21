package main

const (
	POSTGRESQL = "Psql"
	MYSQL      = "MySql"
	SQLITE3    = "Sqlite"

	MODEL_INCLUDE_NULL = "\n\t\"gopkg.in/guregu/null.v3\""
	MODEL_COLUMN       = "\t\t%s\t%s\t`db:\"%s\" json:\"%s\"`"

	HANDLER_PRIMARY_INT = `	%sStr := c.Param("%s")
	%s, err := strconv.ParseInt(%sStr, 10, 64)
	if err != nil {
		bindErr := ae.ParseError("Invalid param value, not a number")
		return c.JSON(bindErr.StatusCode, s.NewOutput(bindErr.BodyError(), false, 0, 0, 0))
	}`  // Lower, Lower, Lower, Lower
	HANDLER_PRIMARY_STR = `	%s := c.Param("%s")` // Lower, Lower
	HANDLER_GET_DELETE  = `	%s := &%s{%s}`       // CamelLower, Camel, HandlerArgSet
	MANAGER_GET_INT     = `	if %s.%s < 1 {
		return ae.MissingParamError("%s")
	}`  // Abbr, Camel, Camel

	MANAGER_GET_STRING = `	if %s.%s == "" {
		return ae.MissingParamError("%s")
	}
	`  // Abbr, Camel, Camel
	MANAGER_POST_AUTOINCREMENT = `if %s.%s < 1 {
		return ae.MissingParamError("%s")
	}
	`  // Abbr, Lower, Camel
	MANAGER_POST_UUID = `if %s.%s == "" {
		return ae.MissingParamError("%s")
	}
	`  // Abbr, Lower, Camel
	MANAGER_POST_NULL = `if !%s.%s.Valid {
		return ae.MissingParamError("%s")
	}
	`  // Abbr, Lower, Camel
	MANAGER_POST_VARCHAR_LEN = `if %s.%s.Valid && len(%s.%s.ValueOrZero()) > %d {
		return ae.StringLengthError("%s", %d)
	}
	`  // Abbr, ColumnCamel, Abbr, ColumnCamel, ColumnLength, ColumnCamel, ColumnLength
	MANAGER_GET_TIME = `	if %s.%s.IsZero() {
		return ae.MissingParamError("%s")
	}
	`  // Abbr, Lower, Camel
	MANAGER_GET_JSON = `	if !%s.%s.ValidJson() {
		return ae.ParseError("%s is invalid JSON")
	}
	`  // Abbr, Lower, Camel
	MANAGER_PUT_SEARCH_STRING = `	%s, ok%s := jParsed.Search("%s").Data().(string)
	if !ok%s {
		return ae.MissingParamError("%s")
	}
	`  // Lower, Camel, Lower, Camel, Camel
	MANAGER_PUT_SEARCH_INT = `	%sFlt, ok%sFlt := jParsed.Search("%s").Data().(float64)
	if !ok%sFlt {
		return ae.MissingParamError("%s")
	}
	%s := int(%sFlt)
	`  // Lower, Camel, Lower, Camel, Camel, Lower, Lower
	MANAGER_PUT_STRUCT_STMT = `	%s := &%s{%s}
	`  // Abbr, Camel, KeySearchList
	MANAGER_PUT_GET_STMT = `	errGet := m.Get(%s)
	if errGet != nil {
		return errGet
	}
	`  // Abbr
	MANAGER_PUT_INT_ASSIGN = `// %s
	%s, ok%s := jParsed.Search("%s").Data().(float64)
	if ok%s {
		%s.%s = int64(%s)
	}
	`  // ColCamel, ColLower, ColCamel, ColLower, ColCamel, Abbr, ColCamel, ColLower
	MANAGER_PUT_INT_NULL_ASSIGN = `// %s
	%s, ok%s := jParsed.Search("%s").Data().(float64)
	if ok%s {
		%s.%s.Scan(int64(%s))
	}
	`  // ColCamel, ColLowerCamel, ColCamel, ColLower, ColCamel, Abbr, ColCamel, ColLowerCamel
	MANAGER_PUT_FLOAT_NULL_ASSIGN = `// %s
	%s, ok%s := jParsed.Search("%s").Data().(float64)
	if ok%s {
		%s.%s.Scan(%s)
	}
	`  // ColCamel, ColLowerCamel, ColCamel, ColLower, ColCamel, Abbr, ColCamel, ColLowerCamel
	MANAGER_PUT_STRING_NULL_ASSIGN = `// %s
	%s, ok%s := jParsed.Search("%s").Data().(string)
	if ok%s {
		%s.%s.Scan(%s)
	}
	`  // ColCamel, ColCamelLower, ColCamel, ColLower, ColCamel, Abbr, ColCamel, ColCamelLower
	MANAGER_PUT_BOOL_NULL_ASSIGN = `// %s
	%s, ok%s := jParsed.Search("%s").Data().(bool)
	if ok%s {
		%s.%s.Scan(%s)
	}
	`  // ColCamel, ColLowerCamel, ColCamel, ColLower, ColCamel, Abbr, ColCamel, ColLowerCamel
	MANAGER_PUT_JSON_NULL_ASSIGN = `// %s
	if jParsed.Exists("%s") {
		%s := json.RawMessage(jParsed.Search("%s").Bytes())
		if !ValidJson(*%s) {
			return ae.ParseError("Invalid JSON syntax for %s")
		}
		%s.%s = &%s
	}
	`  // ColCamel, ColLower, ColLowerCamel, ColLower, ColLowerCamel, ColCamel, Abbr, ColCamel, ColLowerCamel
	MANAGER_PUT_TIME_NULL_ASSIGN = `// %s
	%s, ok%s := jParsed.Search("%s").Data().(string)
	if ok%s {
		%sTime, errParse := time.Parse(time.RFC3339, %s)
		if errParse != nil {
			return ae.ParseError("%s: unable to parse time")
		}
		%s.%s.Scan(%sTime)
	}
	`  // ColCamel, ColLowerCamel, ColCamel, ColLower, ColCamel, ColLowerCamel, ColLowerCamel, ColCamel, Abbr, ColCamel, ColLowerCamel
	DATA_LAST_ID = `var lastId int64
	if rows.Next() {
		rows.Scan(&lastId)
	}
	%s.%s = int(lastId)
	`  // Abbr, ColCamel

	// TESTS
	HANDLER_TEST_INT_FAILURE = `
	func Test%sHandlerGetFailureInvalidInt(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/{{.GetDeleteUrl}}")
		c.SetParamNames("%s")
		c.SetParamValues("a")
	
		man := &MockManager%s{}
		h := NewHandler%s(man)
	
		h.Get(c)
	
		be := ae.BodyError{}
		json.Unmarshal(rec.Body.Bytes(), &be)
	
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "Invalid param value, not a number", be.Detail)
	}`  // camel, get_delete_url, col_lower, camel, camel
	HANDLER_TEST_INT_ZERO = `
	func Test%sHandlerGetFailureZeroInt(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/{{.GetDeleteUrl}}")
		c.SetParamNames("%s")
		c.SetParamValues("0")
	
		man := &MockManager%s{}
		h := NewHandler%s(man)
	
		h.Get(c)
	
		be := ae.BodyError{}
		json.Unmarshal(rec.Body.Bytes(), &be)
	
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "zero value", be.Detail)
	}`  // camel, get_delete_url, col_lower, camel, camel

	MAIN_COMMON_PATH = `{{.ProjectFile.SubPackage}} "{{.ProjectFile.ProjectPathEncoded}}\/internal\/{{.ProjectFile.SubPackage}}"`

	SERVER_ROUTE = `{{.ProjectFile.SubPackage}}.Setup{{.Name.Camel}}(routeGroup)
	\/\/ --- replace server text - do not remove ---
`
	COMMON_IMPORT = `
import (
	"github.com\/labstack\/echo"
	
	\/\/ --- replace header text - do not remove ---
)`
	COMMON_HEADER = `{{.Name.Abbr}} "{{.ProjectFile.ProjectPathEncoded}}\/{{.ProjectFile.SubDirEncoded}}\/{{.Name.Lower}}"
	\/\/ --- replace header text - do not remove ---`

	COMMON_SECTION = `\/\/ {{.Camel}}
func Setup{{.Camel}}(eg *echo.Group) {
	sl := {{.Abbr}}.InitStorage()
	ml := {{.Abbr}}.NewManager{{.Camel}}(sl)
	hl := {{.Abbr}}.NewHandler{{.Camel}}(ml)
	hl.Load{{.Camel}}Routes(eg)
}
	
\/\/ --- replace section text - do not remove ---`
)
