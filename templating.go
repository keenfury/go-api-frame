package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

func (ep *EndPoint) BuildTemplateParts() {
	// model
	ep.BuildModelTemplate()
	// handler
	ep.BuildHandlerTemplate()
	// manager
	ep.BuildManagerTemplate()
	// data
	ep.BuildDataTemplate()
	// grpc
	ep.BuildGrpc()
	// server & common hooks
	ep.BuildAPIHooks()
}

func (ep *EndPoint) BuildModelTemplate() {
	cArray := []string{}
	for _, c := range ep.Columns {
		cArray = append(cArray, fmt.Sprintf(MODEL_COLUMN_WO_GORM, c.ColumnName.Camel, c.GoType, c.ColumnName.Lower, c.ColumnName.Camel))
		if ep.ProjectFile.UseORM {
			cArray = append(cArray, fmt.Sprintf(MODEL_COLUMN_W_GORM, c.ColumnName.Camel, c.GoType, c.ColumnName.Lower, c.ColumnName.Camel, c.ColumnName.Lower))
		}
	}
	ep.ModelRows = strings.Join(cArray, "\n")
	if ep.HaveNullColumns {
		ep.ModelIncludeNull = MODEL_INCLUDE_NULL
	}
	initStorage := []string{}
	split := strings.Split(ep.ProjectFile.Storages, " ")
	for _, s := range split {
		if string(s[0]) == "s" {
			initStorage = append(initStorage, fmt.Sprint("\tif config.StorageSQL {\n\t\treturn InitSQL()\n\t}"))
		}
		if string(s[0]) == "f" {
			initStorage = append(initStorage, fmt.Sprintf("\tif config.StorageFile {\n\t\treturn &File%s{}\n\t}", ep.Camel))
		}
		if string(s[0]) == "m" {
			initStorage = append(initStorage, fmt.Sprintf("\tif config.StorageMongo {\n\t\treturn InitMongo()\n\t}"))
		}
	}
	ep.InitStorage = strings.Join(initStorage, "\n")
}

func (ep *EndPoint) BuildHandlerTemplate() {
	// build get/delete url
	getDeleteUrl := ""
	foundOne := false
	for _, c := range ep.Columns {
		if c.PrimaryKey {
			if foundOne {
				getDeleteUrl += fmt.Sprintf("/%s/:%s", c.ColumnName.Lower, c.ColumnName.Lower)
			} else {
				getDeleteUrl = fmt.Sprintf(":%s", c.ColumnName.Lower)
				foundOne = true
			}
		}
	}
	ep.HandlerGetDeleteUrl = getDeleteUrl
	// build get/delete assign and args
	getDeleteAssign := ""
	setArgs := ""
	foundOne = false
	for _, c := range ep.Columns {
		if c.PrimaryKey {
			if c.GoType == "string" {
				if foundOne {
					getDeleteAssign += "\n"
					setArgs += ", "
				}
				getDeleteAssign += fmt.Sprintf(HANDLER_PRIMARY_STR, c.ColumnName.Lower, c.ColumnName.Lower)
				setArgs += fmt.Sprintf("%s: %s", c.ColumnName.Camel, c.ColumnName.Lower)
				foundOne = true
			}
			if c.GoType == "int" {
				if foundOne {
					getDeleteAssign += "\n"
					setArgs += ", "
				}
				getDeleteAssign += fmt.Sprintf(HANDLER_PRIMARY_INT, c.ColumnName.Lower, c.ColumnName.Lower, c.ColumnName.Lower, c.ColumnName.Lower)
				setArgs += fmt.Sprintf("%s: int(%s)", c.ColumnName.Camel, c.ColumnName.Lower)
				foundOne = true
				ep.HandlerStrConv = "\n\t\"strconv\""
			}
		}
	}
	ep.HandlerGetDeleteAssign = getDeleteAssign
	ep.HandlerArgSet = setArgs
}

func (ep *EndPoint) BuildManagerTemplate() {
	getDeleteRow := ""
	postRow := "\t"
	putRow := ""
	patchRow := ""
	patchSearch := ""
	setArgs := ""
	foundOneKey := false
	getDeleteKeyTestSuccessful := []string{}
	getDeleteKeyTestFailure := []string{}
	for _, c := range ep.Columns {
		if c.PrimaryKey {
			if c.GoType == "string" {
				if foundOneKey {
					setArgs += ", "
					patchSearch += "\n"
					getDeleteRow += "\n"
				}
				getDeleteRow += fmt.Sprintf(MANAGER_GET_STRING, ep.Abbr, c.ColumnName.Camel, c.ColumnName.Camel)
				patchSearch += fmt.Sprintf(MANAGER_PATCH_SEARCH_STRING, c.ColumnName.Lower, c.ColumnName.Camel, c.ColumnName.Camel, c.ColumnName.Camel, c.ColumnName.Camel)
				getDeleteKeyTestSuccessful = append(getDeleteKeyTestSuccessful, fmt.Sprintf(`%s: "test id"`, c.ColumnName.Camel))
				getDeleteKeyTestFailure = append(getDeleteKeyTestFailure, fmt.Sprintf(`%s: ""`, c.ColumnName.Camel))
				foundOneKey = true
			}
			if c.GoType == "int" {
				if foundOneKey {
					setArgs += ", "
					patchSearch += "\n"
				}
				getDeleteRow += fmt.Sprintf(MANAGER_GET_INT, ep.Abbr, c.ColumnName.Camel, c.ColumnName.Camel)
				patchSearch += fmt.Sprintf(MANAGER_PATCH_SEARCH_INT, c.ColumnName.Lower, c.ColumnName.Camel, c.ColumnName.Camel, c.ColumnName.Camel, c.ColumnName.Camel, c.ColumnName.Lower, c.ColumnName.Lower)
				getDeleteKeyTestSuccessful = append(getDeleteKeyTestSuccessful, fmt.Sprintf(`%s: 1`, c.ColumnName.Camel))
				getDeleteKeyTestFailure = append(getDeleteKeyTestFailure, fmt.Sprintf(`%s: 0`, c.ColumnName.Camel))
				foundOneKey = true
			}
			setArgs += fmt.Sprintf("%s: %s", c.ColumnName.Camel, c.ColumnName.Lower)
		}
	}
	patchRow = patchSearch + fmt.Sprintf(MANAGER_PATCH_STRUCT_STMT, ep.Abbr, ep.Camel, setArgs) + fmt.Sprintf(MANAGER_PATCH_GET_STMT, ep.Abbr)
	putTests := []PostPutTest{{Name: "successful", Failure: false}}
	postTests := []PostPutTest{{Name: "successful", Failure: false}}
	InitializeColumnTests()
	for _, c := range ep.Columns {
		columnTestStrAdded := false
		// put rows
		if c.PrimaryKey {
			putRow += fmt.Sprintf(MANAGER_GET_INT, ep.Abbr, c.ColumnName.Camel, c.ColumnName.Camel)
			AppendColumnTest(c.ColumnName.Camel, c.GoType, true)
			putTests = append(putTests, PostPutTest{Name: fmt.Sprintf("invalid %s", c.ColumnName.LowerCamel), Failure: true, ForColumn: c.ColumnName.Camel})
			columnTestStrAdded = true
		}
		// post rows
		if c.GoType == "string" || c.GoType == "null.String" {
			if c.DBType == "uuid" {
				postRow += fmt.Sprintf(MANAGER_POST_UUID, ep.Abbr, c.ColumnName.Camel, c.ColumnName.Camel)
				putRow += fmt.Sprintf(MANAGER_POST_UUID, ep.Abbr, c.ColumnName.Camel, c.ColumnName.Camel)
				AppendColumnTest(c.ColumnName.Camel, c.GoType, false)
				postTests = append(postTests, PostPutTest{Name: "invalid UUID", Failure: true, ForColumn: c.ColumnName.Camel})
				putTests = append(putTests, PostPutTest{Name: "invalid UUID", Failure: true, ForColumn: c.ColumnName.Camel})
				columnTestStrAdded = true
			} else {
				if !c.Null {
					postRow += fmt.Sprintf(MANAGER_POST_NULL, ep.Abbr, c.ColumnName.Camel, c.ColumnName.Camel)
					putRow += fmt.Sprintf(MANAGER_POST_NULL, ep.Abbr, c.ColumnName.Camel, c.ColumnName.Camel)
					AppendColumnTest(c.ColumnName.Camel, c.GoType, false)
					postTests = append(postTests, PostPutTest{Name: fmt.Sprintf("invalid %s", c.ColumnName.LowerCamel), Failure: true, ForColumn: c.ColumnName.Camel})
					putTests = append(putTests, PostPutTest{Name: fmt.Sprintf("invalid %s", c.ColumnName.LowerCamel), Failure: true, ForColumn: c.ColumnName.Camel})
					columnTestStrAdded = true
				}
				if c.Length > 0 {
					postRow += fmt.Sprintf(MANAGER_POST_VARCHAR_LEN, ep.Abbr, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel, c.Length, c.ColumnName.Camel, c.Length)
					putRow += fmt.Sprintf(MANAGER_POST_VARCHAR_LEN, ep.Abbr, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel, c.Length, c.ColumnName.Camel, c.Length)
					AppendColumnTest(c.ColumnName.Camel, c.GoType, false)
					postTests = append(postTests, PostPutTest{Name: fmt.Sprintf("length %s", c.ColumnName.LowerCamel), Failure: true, ForColumn: c.ColumnName.Camel, ColumnLength: int(c.Length)})
					putTests = append(putTests, PostPutTest{Name: fmt.Sprintf("length %s", c.ColumnName.LowerCamel), Failure: true, ForColumn: c.ColumnName.Camel, ColumnLength: int(c.Length)})
					columnTestStrAdded = true
				}
			}
		} else {
			if !c.Null && !c.PrimaryKey {
				postRow += fmt.Sprintf(MANAGER_POST_NULL, ep.Abbr, c.ColumnName.Camel, c.ColumnName.Camel)
				putRow += fmt.Sprintf(MANAGER_POST_NULL, ep.Abbr, c.ColumnName.Camel, c.ColumnName.Camel)
				AppendColumnTest(c.ColumnName.Camel, c.GoType, false)
				postTests = append(postTests, PostPutTest{Name: fmt.Sprintf("invalid %s", c.ColumnName.LowerCamel), Failure: true, ForColumn: c.ColumnName.Camel})
				putTests = append(putTests, PostPutTest{Name: fmt.Sprintf("invalid %s", c.ColumnName.LowerCamel), Failure: true, ForColumn: c.ColumnName.Camel})
				columnTestStrAdded = true
			}
		}
		if !columnTestStrAdded && c.Null && !c.PrimaryKey {
			// add column to all the other tests with good data
			AppendColumnTest(c.ColumnName.Camel, c.GoType, false)
		}
		// patch rows
		if c.GoType == "null.String" {
			if !c.PrimaryKey {
				patchLenCheck := ""
				if c.Length > 0 {
					patchLenCheck = fmt.Sprintf(MANAGER_PATCH_VARCHAR_LEN, ep.Abbr, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel, c.Length, c.ColumnName.Camel, c.Length)
				}
				patchRow += fmt.Sprintf(MANAGER_PATCH_STRING_NULL_ASSIGN, c.ColumnName.Camel, c.ColumnName.LowerCamel, c.ColumnName.Camel, c.ColumnName.Camel, c.ColumnName.Camel, patchLenCheck, ep.Abbr, c.ColumnName.Camel, c.ColumnName.LowerCamel)
			}
		}
		if c.GoType == "int" || c.GoType == "null.Int" {
			if !c.PrimaryKey {
				patchRow += fmt.Sprintf(MANAGER_PATCH_INT_NULL_ASSIGN, c.ColumnName.Camel, c.ColumnName.LowerCamel, c.ColumnName.Camel, c.ColumnName.Camel, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel, c.ColumnName.LowerCamel)
			}
		}
		if c.GoType == "null.Float" {
			patchRow += fmt.Sprintf(MANAGER_PATCH_FLOAT_NULL_ASSIGN, c.ColumnName.Camel, c.ColumnName.LowerCamel, c.ColumnName.Camel, c.ColumnName.Camel, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel, c.ColumnName.LowerCamel)
		}
		if c.GoType == "null.Bool" {
			patchRow += fmt.Sprintf(MANAGER_PATCH_BOOL_NULL_ASSIGN, c.ColumnName.Camel, c.ColumnName.LowerCamel, c.ColumnName.Camel, c.ColumnName.Camel, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel, c.ColumnName.LowerCamel)
		}
		if c.GoType == "*json.RawMessage" {
			patchRow += fmt.Sprintf(MANAGER_PATCH_JSON_NULL_ASSIGN, c.ColumnName.Camel, c.ColumnName.Camel, c.ColumnName.LowerCamel, c.ColumnName.Camel, c.ColumnName.LowerCamel, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel, c.ColumnName.LowerCamel)
		}
		if c.GoType == "null.Time" {
			ep.ManagerImportTest = "\n\t\"time\""
			patchRow += fmt.Sprintf(MANAGER_PATCH_TIME_NULL_ASSIGN, c.ColumnName.Camel, c.ColumnName.LowerCamel, c.ColumnName.Camel, c.ColumnName.Camel, c.ColumnName.Camel, c.ColumnName.LowerCamel, c.ColumnName.LowerCamel, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel, c.ColumnName.LowerCamel)
			ep.ManagerTime = "\n\t\"time\""
		}
	}
	uuidColumn := ""
	for _, c := range ep.Columns {
		if c.DBType == "UUID" {
			uuidColumn = c.ColumnName.Camel
		}
	}
	if uuidColumn != "" {
		postRow += fmt.Sprintf(`%s.%s = util.GenerateUUID()`, ep.Abbr, uuidColumn)
		ep.ManagerUtilPath = fmt.Sprintf(`"%s/util"`, ep.ProjectPath)
	}
	ep.ManagerGetRow = strings.TrimRight(getDeleteRow, "\n")
	ep.ManagerPostRows = strings.TrimRight(postRow, "\n")
	ep.ManagerPutRows = strings.TrimRight(putRow, "\n")
	ep.ManagerPatchRows = strings.TrimRight(patchRow, "\n")
	managerGetSuccessfulRow := ""
	managerGetFailureRow := ""
	managerDeleteSuccessfulRow := ""
	managerDeleteFailureRow := ""
	if len(getDeleteKeyTestSuccessful) > 0 {
		managerGetSuccessfulRow = fmt.Sprintf("{\n\t\t\t\"successful\",\n\t\t\t&%s{%s},\n\t\t\tfalse,\n\t\t\t[]*gomock.Call{\n\t\t\t\tmockData%s.EXPECT().Read(gomock.Any()).Return(nil),\n\t\t\t},\n\t\t},", ep.Camel, strings.Join(getDeleteKeyTestSuccessful, ", "), ep.Camel)
		managerDeleteSuccessfulRow = fmt.Sprintf("{\n\t\t\t\"successful\",\n\t\t\t&%s{%s},\n\t\t\tfalse,\n\t\t\t[]*gomock.Call{\n\t\t\t\tmockData%s.EXPECT().Delete(gomock.Any()).Return(nil),\n\t\t\t},\n\t\t},", ep.Camel, strings.Join(getDeleteKeyTestSuccessful, ", "), ep.Camel)
	}
	if len(getDeleteKeyTestFailure) > 0 {
		managerGetFailureRow = fmt.Sprintf("{\n\t\t\t\"invalid id\",\n\t\t\t&%s{%s},\n\t\t\ttrue,\n\t\t\t[]*gomock.Call{},\n\t\t},", ep.Camel, strings.Join(getDeleteKeyTestFailure, ", "))
		managerDeleteFailureRow = fmt.Sprintf("{\n\t\t\t\"invalid id\",\n\t\t\t&%s{%s},\n\t\t\ttrue,\n\t\t\t[]*gomock.Call{},\n\t\t},", ep.Camel, strings.Join(getDeleteKeyTestFailure, ", "))
	}
	ep.ManagerGetTestRow = fmt.Sprintf("%s\n\t\t%s", managerGetSuccessfulRow, managerGetFailureRow)
	ep.ManagerDeleteTestRow = fmt.Sprintf("%s\n\t\t%s", managerDeleteSuccessfulRow, managerDeleteFailureRow)
	managerPostTestRow := []string{}
	managerPutTestRow := []string{}

	for _, postTest := range postTests {
		call := ""
		if !postTest.Failure {
			call = fmt.Sprintf("mockData%s.EXPECT().Create(gomock.Any()).Return(nil).AnyTimes(),\n\t\t\t", ep.Camel)
		}
		columnStr := []string{}
		for name, column := range PostTests {
			columnValid := true
			columnLength := 0
			if postTest.ForColumn == name {
				columnValid = false
				if postTest.ColumnLength > 0 {
					columnLength = postTest.ColumnLength
				}
			}
			columnStr = append(columnStr, TranslateType(name, column.GoType, columnLength, columnValid))
		}
		managerPostTestRow = append(managerPostTestRow, fmt.Sprintf("{\n\t\t\t\"%s\",\n\t\t\t&%s{%s},\n\t\t\t%t,\n\t\t\t[]*gomock.Call{%s},\n\t\t},", postTest.Name, ep.Camel, strings.Join(columnStr, ", "), postTest.Failure, call))
	}
	for _, putTest := range putTests {
		call := ""
		if !putTest.Failure {
			call = fmt.Sprintf("mockData%s.EXPECT().Update(gomock.Any()).Return(nil).AnyTimes(),\n\t\t\t", ep.Camel)
		}
		columnStr := []string{}
		for name, column := range PutTests {
			columnValid := true
			columnLength := 0
			if putTest.ForColumn == name {
				columnValid = false
				if putTest.ColumnLength > 0 {
					columnLength = putTest.ColumnLength
				}
			}
			columnStr = append(columnStr, TranslateType(name, column.GoType, columnLength, columnValid))
		}
		managerPutTestRow = append(managerPutTestRow, fmt.Sprintf("{\n\t\t\t\"%s\",\n\t\t\t&%s{%s},\n\t\t\t%t,\n\t\t\t[]*gomock.Call{%s},\n\t\t},", putTest.Name, ep.Camel, strings.Join(columnStr, ", "), putTest.Failure, call))
	}
	ep.ManagerPostTestRow = strings.Join(managerPostTestRow, "\n\t\t")
	ep.ManagerPutTestRow = strings.Join(managerPutTestRow, "\n\t\t")
}

func (ep *EndPoint) BuildDataTemplate() {
	if ep.DynamicSchema {
		ep.DataTablePrefix = "fmt.Sprintf("
		ep.DataTable = fmt.Sprintf("%s.%s", ep.Schema, ep.Lower)
		ep.DataTablePostfix = fmt.Sprintf(", %s)", ep.DynamicSchemaPostfix)
	} else {
		ep.DataTable = ep.Lower
	}
	SqlGetColumns := ""
	foundOneKey := false
	foundOnePatch := false
	foundOnePost := false
	keys := ""
	patchKeys := ""
	keyCount := 1
	values := ""
	listOrder := ""
	postColumn := ""
	postColumnNames := ""
	patchColumn := ""
	keysCount := 1
	foundSerial := ""
	foundSerialDB := ""
	fileKey := []string{}
	fileGetColumn := []string{}
	filePostIncrInit := []string{}
	filePostIncrCheck := []string{}
	filePostIncr := []string{}
	for i, c := range ep.Columns {
		fileGetColumn = append(fileGetColumn, fmt.Sprintf("%s.%s = %sObj.%s", ep.Abbr, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel))
		if c.DBType == "autoincrement" {
			foundSerialDB = c.ColumnName.Lower
			foundSerial = c.ColumnName.Camel
		}
		if c.DBType != "autoincrement" {
			if foundOnePost {
				postColumn += fmt.Sprintf(",\n\t\t\t%s", c.ColumnName.Lower)
				postColumnNames += fmt.Sprintf(",\n\t\t\t:%s", c.ColumnName.Lower)
			} else {
				postColumn += fmt.Sprintf("%s", c.ColumnName.Lower)
				postColumnNames += fmt.Sprintf(":%s", c.ColumnName.Lower)
				foundOnePost = true
			}
		}
		if i == 0 {
			SqlGetColumns += fmt.Sprintf("%s", c.ColumnName.Lower)
		} else {
			SqlGetColumns += fmt.Sprintf(",\n\t\t\t%s", c.ColumnName.Lower)
		}
		if c.PrimaryKey {
			if foundOneKey {
				keys += " and "
				values += ", "
				listOrder += ", "
			}
			foundOneKey = true
			if ep.SQLProvider == MYSQL {
				keys += fmt.Sprintf("%s = ?", c.ColumnName.Lower)
			} else {
				keys += fmt.Sprintf("%s = $%d", c.ColumnName.Lower, keyCount)
			}
			patchKeys += fmt.Sprintf("%s = :%s", c.ColumnName.Lower, c.ColumnName.Lower)
			keyCount++
			values += fmt.Sprintf("%s.%s", ep.Name.Abbr, c.ColumnName.Camel)
			listOrder += fmt.Sprintf("%s", c.ColumnName.Lower)
			keysCount++
			fileKey = append(fileKey, fmt.Sprintf("%sObj.%s == %s.%s", ep.Abbr, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel))
			if c.DBType == "autoincrement" || c.DBType == "int" {
				filePostIncrInit = append(filePostIncrInit, fmt.Sprintf("max%s := 0", c.ColumnName.Camel))
				filePostIncrCheck = append(filePostIncrCheck, fmt.Sprintf("\t\tif %sObj.%s > max%s {\n\t\t\tmax%s = %sObj.%s\n\t\t}", ep.Abbr, c.ColumnName.Camel, c.ColumnName.Camel, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel))
				filePostIncr = append(filePostIncr, fmt.Sprintf("\t%s.%s = max%s + 1", ep.Abbr, c.ColumnName.Camel, c.ColumnName.Camel))
			}
		} else {
			if foundOnePatch {
				patchColumn += ",\n\t\t\t"
			}
			patchColumn += fmt.Sprintf("%s = :%s", c.ColumnName.Lower, c.ColumnName.Lower)
			foundOnePatch = true
		}
	}
	ep.SqlGetColumns = strings.TrimRight(SqlGetColumns, "\n")
	ep.SqlTableKeyKeys = keys
	ep.SqlTableKeyValues = values
	ep.SqlTableKeyListOrder = listOrder
	ep.SqlPostColumns = strings.TrimRight(postColumn, "\n")
	ep.SqlPostColumnsNamed = strings.TrimRight(postColumnNames, "\n")
	ep.SqlPatchColumns = strings.TrimRight(patchColumn, "\n")
	ep.SqlPatchWhere = patchKeys
	if foundSerial != "" {
		if ep.SQLProvider == POSTGRESQL {
			ep.SqlPostReturning = fmt.Sprintf(" returning %s", foundSerialDB)
		}
		ep.SqlPostLastId = fmt.Sprintf(DATA_LAST_ID, ep.Abbr, foundSerial)
	}
	ep.FileKeys = strings.Join(fileKey, " && ")
	ep.FileGetColumns = strings.Join(fileGetColumn, "\n\t\t\t")
	ep.FilePostIncr = fmt.Sprintf("%s\n\tfor _, %sObj := range %ss {\n%s\n\t}\n%s", strings.Join(filePostIncrInit, "\n"), ep.Abbr, ep.Abbr, strings.Join(filePostIncrCheck, "\n"), strings.Join(filePostIncr, "\n"))
}

func (ep *EndPoint) BuildGrpc() {
	protoFile := fmt.Sprintf("%s/pkg/proto/%s.proto", ep.ProjectFile.FullPath, ep.AppName)
	if _, err := os.Stat(protoFile); os.IsNotExist(err) {
		fmt.Printf("%s: file not found\n", protoFile)
		return
	}
	lines := []string{}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("message %s {", ep.Camel))
	translateInLines := []string{}
	translateOutLines := []string{}
	for i, column := range ep.Columns {
		idx := i + 1 // start the count at 1
		typeValue := "string"
		var inLine, outLine string
		switch column.GoType {
		case "float64", "null.Float":
			typeValue = "double"
			outLine = fmt.Sprintf("\tproto%s.%s = float64(%s.%s)", ep.Camel, column.ColumnName.Camel, ep.Abbr, column.ColumnName.Camel)
			inLine = fmt.Sprintf("\t%s.%s = in.%s", ep.Abbr, column.ColumnName.Camel, column.ColumnName.Camel)
			if column.GoType == "null.Float" {
				outLine = fmt.Sprintf("\tproto%s.%s = %s.%s.Float64", ep.Camel, column.ColumnName.Camel, ep.Abbr, column.ColumnName.Camel)
				inLine = fmt.Sprintf("\t%s.%s.Scan(in.%s)", ep.Camel, column.ColumnName.Camel, column.ColumnName.Camel)
			}
		case "float32":
			typeValue = "float"
			outLine = fmt.Sprintf("\tproto%s.%s = float32(%s.%s)", ep.Camel, column.ColumnName.Camel, ep.Abbr, column.ColumnName.Camel)
			inLine = fmt.Sprintf("\t%s.%s = in.%s", ep.Abbr, column.ColumnName.Camel, column.ColumnName.Camel)
		case "int32":
			typeValue = "int32"
			outLine = fmt.Sprintf("\tproto%s.%s = int32(%s.%s)", ep.Camel, column.ColumnName.Camel, ep.Abbr, column.ColumnName.Camel)
			inLine = fmt.Sprintf("\t%s.%s = in.%s", ep.Abbr, column.ColumnName.Camel, column.ColumnName.Camel)
		case "int64", "int", "null.Int":
			typeValue = "int64"
			outLine = fmt.Sprintf("\tproto%s.%s = int64(%s.%s)", ep.Camel, column.ColumnName.Camel, ep.Abbr, column.ColumnName.Camel)
			inLine = fmt.Sprintf("\t%s.%s = int(in.%s)", ep.Abbr, column.ColumnName.Camel, column.ColumnName.Camel)
			if column.GoType == "null.Int" {
				outLine = fmt.Sprintf("\tproto%s.%s = %s.%s.Int64", ep.Camel, column.ColumnName.Camel, ep.Abbr, column.ColumnName.Camel)
				inLine = fmt.Sprintf("\t%s.%s.Scan(in.%s)", ep.Abbr, column.ColumnName.Camel, column.ColumnName.Camel)
			}
		case "uint32":
			typeValue = "uint32"
			outLine = fmt.Sprintf("\tproto%s.%s = uint32(%s.%s)", ep.Camel, column.ColumnName.Camel, ep.Abbr, column.ColumnName.Camel)
			inLine = fmt.Sprintf("\t%s.%s = in.%s", ep.Abbr, column.ColumnName.Camel, column.ColumnName.Camel)
		case "uint64":
			typeValue = "uint64"
			outLine = fmt.Sprintf("\tproto%s.%s = uint64(%s.%s)", ep.Camel, column.ColumnName.Camel, ep.Abbr, column.ColumnName.Camel)
			inLine = fmt.Sprintf("\t%s.%s = in.%s", ep.Abbr, column.ColumnName.Camel, column.ColumnName.Camel)
		case "bool", "null.Bool":
			typeValue = "bool"
			outLine = fmt.Sprintf("\tproto%s.%s = %s.%s", ep.Camel, column.ColumnName.Camel, ep.Abbr, column.ColumnName.Camel)
			inLine = fmt.Sprintf("\t%s.%s = in.%s", ep.Abbr, column.ColumnName.Camel, column.ColumnName.Camel)
			if column.GoType == "null.Bool" {
				outLine = fmt.Sprintf("\tproto%s.%s = %s.%s.Bool", ep.Camel, column.ColumnName.Camel, ep.Abbr, column.ColumnName.Camel)
				inLine = fmt.Sprintf("\t%s.%s.Scan(in.%s)", ep.Abbr, column.ColumnName.Camel, column.ColumnName.Camel)
			}
		case "[]byte":
			typeValue = "bytes"
			outLine = fmt.Sprintf("\tproto%s.%s = %s.%s", ep.Camel, column.ColumnName.Camel, ep.Abbr, column.ColumnName.Camel)
			inLine = fmt.Sprintf("\t%s.%s = in.%s", ep.Abbr, column.ColumnName.Camel, column.ColumnName.Camel)
		default:
			typeValue = "string"
			outLine = fmt.Sprintf("\tproto%s.%s = %s.%s", ep.Camel, column.ColumnName.Camel, ep.Abbr, column.ColumnName.Camel)
			inLine = fmt.Sprintf("\t%s.%s = in.%s", ep.Abbr, column.ColumnName.Camel, column.ColumnName.Camel)
			if column.GoType == "null.String" {
				outLine = fmt.Sprintf("\tproto%s.%s = %s.%s.String", ep.Camel, column.ColumnName.Camel, ep.Abbr, column.ColumnName.Camel)
				inLine = fmt.Sprintf("\t%s.%s.Scan(in.%s)", ep.Abbr, column.ColumnName.Camel, column.ColumnName.Camel)
			}
			if column.GoType == "time.Time" {
				outLine = fmt.Sprintf("\tproto%s.%s = %s.%s.Time.Format(time.RFC3339)", ep.Camel, column.ColumnName.Camel, ep.Abbr, column.ColumnName.Camel)
			}
			if column.GoType == "null.Time" {
				inLine = fmt.Sprintf("\t%s.%s.Scan(in.%s)", ep.Abbr, column.ColumnName.Camel, column.ColumnName.Camel)
			}
		}
		lines = append(lines, fmt.Sprintf("\t%s %s = %d;", typeValue, column.ColumnName.Camel, idx))
		translateInLines = append(translateInLines, inLine)
		translateOutLines = append(translateOutLines, outLine)
	}
	lines = append(lines, "}")
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("message %sResponse {", ep.Camel))
	lines = append(lines, fmt.Sprintf("\t%s %s = 1;", ep.Name.Camel, ep.Name.Camel))
	lines = append(lines, fmt.Sprintf("\tResult result = 2;"))
	lines = append(lines, "}")
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("message %sRepeatResponse {", ep.Camel))
	lines = append(lines, fmt.Sprintf("\trepeated %s %s = 1;", ep.Name.Camel, ep.Name.Camel))
	lines = append(lines, fmt.Sprintf("\tResult result = 2;"))
	lines = append(lines, "}")
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("service %sService {", ep.Name.Camel))
	lines = append(lines, fmt.Sprintf("\trpc Get%s(IDIn) returns (%sResponse);", ep.Name.Camel, ep.Name.Camel))
	lines = append(lines, fmt.Sprintf("\trpc List%s(%s) returns (%sRepeatResponse);", ep.Name.Camel, ep.Name.Camel, ep.Name.Camel))
	lines = append(lines, fmt.Sprintf("\trpc Post%s(%s) returns (%sResponse);", ep.Name.Camel, ep.Name.Camel, ep.Name.Camel))
	lines = append(lines, fmt.Sprintf("\trpc Put%s(%s) returns (Result);", ep.Name.Camel, ep.Name.Camel))
	lines = append(lines, fmt.Sprintf("\trpc Patch%s(%s) returns (Result);", ep.Name.Camel, ep.Name.Camel))
	lines = append(lines, fmt.Sprintf("\trpc Delete%s(IDIn) returns (Result);", ep.Name.Camel))
	lines = append(lines, "}")
	lines = append(lines, "")

	file, err := os.OpenFile(protoFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("%s: unable to open file with error: %s\n", protoFile, err)
		return
	}
	defer file.Close()
	_, err = file.WriteString(strings.Join(lines, "\n"))
	if err != nil {
		fmt.Printf("%s: unable to write to file with error: %s\n", protoFile, err)
	}
	// save off translateIn/Out
	ep.GrpcTranslateIn = strings.Join(translateInLines, "\n\t")
	ep.GrpcTranslateOut = strings.Join(translateOutLines, "\n\t")
}

func (ep *EndPoint) BuildAPIHooks() {
	// hook into rest main
	apiFile := fmt.Sprintf("%s/cmd/rest/main.go", ep.ProjectFile.FullPath)
	if _, err := os.Stat(apiFile); os.IsNotExist(err) {
		fmt.Printf("%s is missing unable to write in hooks\n", apiFile)
	} else {
		var serverReplace bytes.Buffer
		tServer := template.Must(template.New("server").Parse(SERVER_ROUTE))
		errServer := tServer.Execute(&serverReplace, ep)
		if errServer != nil {
			fmt.Printf("%s: template error [%s]\n", apiFile, errServer)
		} else {
			cmdServer := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace server text - do not remove ---/%s/g' %s`, serverReplace.String(), apiFile)
			execServer := exec.Command("bash", "-c", cmdServer)
			errServerCmd := execServer.Run()
			if errServerCmd != nil {
				fmt.Printf("%s: error in replace for server [%s]\n", apiFile, errServerCmd)
			}
		}
		onceReplace := `routeGroup := e.Group("v1") \/\/ change to match your uri prefix`
		cmdOnceServer := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace server once text - do not remove ---/%s/g' %s`, onceReplace, apiFile)
		execServerOnce := exec.Command("bash", "-c", cmdOnceServer)
		errServerOnceCmd := execServerOnce.Run()
		if errServerOnceCmd != nil {
			fmt.Printf("%s: error in replace for server once [%s]\n", apiFile, errServerOnceCmd)
		}
		var mainReplace bytes.Buffer
		tMain := template.Must(template.New("server").Parse(MAIN_COMMON_PATH))
		errServer = tMain.Execute(&mainReplace, ep)
		if errServer != nil {
			fmt.Printf("%s: template error [%s]\n", apiFile, errServer)
		} else {
			cmdServer := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace server header text ---/%s/g' %s`, mainReplace.String(), apiFile)
			execServer := exec.Command("bash", "-c", cmdServer)
			errServerCmd := execServer.Run()
			if errServerCmd != nil {
				fmt.Printf("%s: error in replace for main [%s]\n", apiFile, errServerCmd)
			}
		}
		if ep.SQLProvider != "" {
			// handling sql add migration
			migRestMain := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace migration once text - do not remove ---/%s/g' %s`, MIGRATION_CALL, apiFile)
			execRestMain := exec.Command("bash", "-c", migRestMain)
			errExecRestMain := execRestMain.Run()
			if errExecRestMain != nil {
				fmt.Printf("%s: error in replace migration main [%s]\n", apiFile, errExecRestMain)
			}
			migRestHeader := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace migration header once text - do not remove ---/%s/g' %s`, fmt.Sprintf(`mig "%s\/tools\/migration\/src"`, ep.ProjectPathEncoded), apiFile)
			execRestHeader := exec.Command("bash", "-c", migRestHeader)
			errExecRestHeader := execRestHeader.Run()
			if errExecRestHeader != nil {
				fmt.Printf("%s: error in replace migration main [%s]\n", apiFile, errExecRestMain)
			}
		}
	}
	// hook into common file
	commonFile := fmt.Sprintf("%s/%s/common.go", ep.ProjectFile.FullPath, ep.SubDir)
	if _, err := os.Stat(commonFile); os.IsNotExist(err) {
		// create file if not there
		commonSrc := fmt.Sprintf("%s/templates/common.go", os.Getenv("FRAME_PATH"))
		commonDest := fmt.Sprintf("%s/%s/common.go", ep.ProjectFile.FullPath, ep.ProjectFile.SubDir)
		bSrc, errSrc := ioutil.ReadFile(commonSrc)
		if errSrc != nil {
			fmt.Println("Unable to read common.go")
		} else {
			errWrite := ioutil.WriteFile(commonDest, bSrc, 0644)
			if errWrite != nil {
				fmt.Println("Unable to write common.go")
			}
		}
		var importReplace bytes.Buffer
		tImport := template.Must(template.New("import").Parse(COMMON_IMPORT))
		errImport := tImport.Execute(&importReplace, ep)
		if errImport != nil {
			fmt.Println("Import template error:", errImport)
		} else {
			cmdImport := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace import text - do not remove ---/%s/g' %s`, importReplace.String(), commonFile)
			execImport := exec.Command("bash", "-c", cmdImport)
			errImportCmd := execImport.Run()
			if errImportCmd != nil {
				fmt.Println("Error in replace for import:", errImportCmd)
			}
		}
	}
	// cont. with common.go
	// header
	var headerReplace bytes.Buffer
	tHeader := template.Must(template.New("header").Parse(COMMON_HEADER))
	errHeader := tHeader.Execute(&headerReplace, ep)
	if errHeader != nil {
		fmt.Println("Header template error:", errHeader)
		return
	}
	cmdHeader := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace header text - do not remove ---/%s/g' %s`, headerReplace.String(), commonFile)
	execHeader := exec.Command("bash", "-c", cmdHeader)
	errHeaderCmd := execHeader.Run()
	if errHeaderCmd != nil {
		fmt.Println("Error in replace for header:", errHeaderCmd)
	}

	// section
	var sectionReplace bytes.Buffer
	tSection := template.Must(template.New("section").Parse(COMMON_SECTION))
	errSection := tSection.Execute(&sectionReplace, ep)
	if errSection != nil {
		fmt.Println("Section template error:", errSection)
		return
	}
	cmdSection := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace section text - do not remove ---/%s/g' %s`, sectionReplace.String(), commonFile)
	execSection := exec.Command("bash", "-c", cmdSection)
	errSectionCmd := execSection.Run()
	if errSectionCmd != nil {
		fmt.Println("Error in replace for server:", errSectionCmd)
	}
	// hook into grpc file
	grpcFile := fmt.Sprintf("%s/cmd/grpc/main.go", ep.ProjectFile.FullPath)
	if _, err := os.Stat(grpcFile); os.IsNotExist(err) {
		fmt.Printf("%s is missing unable to write in hooks\n", grpcFile)
	} else {
		var grpcReplace bytes.Buffer
		tGrpc := template.Must(template.New("grpc").Parse(GRPC_TEXT))
		errGrpc := tGrpc.Execute(&grpcReplace, ep)
		if errGrpc != nil {
			fmt.Printf("%s: template error [%s]\n", grpcFile, errGrpc)
		} else {
			cmdGrpc := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace grpc text - do not remove ---/%s/g' %s`, grpcReplace.String(), grpcFile)
			execGrpc := exec.Command("bash", "-c", cmdGrpc)
			errGrpcCmd := execGrpc.Run()
			if errGrpcCmd != nil {
				fmt.Printf("%s: error in replace for grpc text [%s]\n", grpcFile, errGrpcCmd)
			}
		}
		var importReplace bytes.Buffer
		tImport := template.Must(template.New("grpc").Parse(GRPC_IMPORT))
		errGrpc = tImport.Execute(&importReplace, ep)
		if errGrpc != nil {
			fmt.Printf("%s: template error [%s]\n", grpcFile, errGrpc)
		} else {
			cmdGrpc := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace grpc import - do not remove ---/%s/g' %s`, importReplace.String(), grpcFile)
			execGrpc := exec.Command("bash", "-c", cmdGrpc)
			errGrpcCmd := execGrpc.Run()
			if errGrpcCmd != nil {
				fmt.Printf("%s: error in replace for grpc [%s]\n", grpcFile, errGrpcCmd)
			}
		}
		var importOnceReplace bytes.Buffer
		tOnce := template.Must(template.New("grpc").Parse(GRPC_IMPORT_ONCE))
		errGrpc = tOnce.Execute(&importOnceReplace, ep)
		if errGrpc != nil {
			fmt.Printf("%s: template error [%s]\n", grpcFile, errGrpc)
		} else {
			cmdGrpc := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace grpc import once - do not remove ---/%s/g' %s`, importOnceReplace.String(), grpcFile)
			execGrpc := exec.Command("bash", "-c", cmdGrpc)
			errGrpcCmd := execGrpc.Run()
			if errGrpcCmd != nil {
				fmt.Printf("%s: error in replace for grpc [%s]\n", grpcFile, errGrpcCmd)
			}
		}
		if ep.SQLProvider != "" {
			// handling sql add migration
			migGrpcMain := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace migration once text - do not remove ---/%s/g' %s`, MIGRATION_CALL, grpcFile)
			execGrpcMain := exec.Command("bash", "-c", migGrpcMain)
			errExecGrpcMain := execGrpcMain.Run()
			if errExecGrpcMain != nil {
				fmt.Printf("%s: error in replace migration main [%s]\n", grpcFile, errExecGrpcMain)
			}
			migGrpcHeader := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace migration header once text - do not remove ---/%s/g' %s`, fmt.Sprintf(`mig "%s\/tools\/migration\/src"`, ep.ProjectPathEncoded), grpcFile)
			execGrpcHeader := exec.Command("bash", "-c", migGrpcHeader)
			errExecGrpcHeader := execGrpcHeader.Run()
			if errExecGrpcHeader != nil {
				fmt.Printf("%s: error in replace migration main [%s]\n", grpcFile, errExecGrpcHeader)
			}
			migGrpcHeaderOs := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace migration header os once text - do not remove ---/%s/g' %s`, MIGRATION_GRPC_HEADER_ONCE, grpcFile)
			execGrpcHeaderOs := exec.Command("bash", "-c", migGrpcHeaderOs)
			errExecGrpcHeaderOs := execGrpcHeaderOs.Run()
			if errExecGrpcHeaderOs != nil {
				fmt.Printf("%s: error in replace migration main [%s]\n", grpcFile, errExecGrpcHeaderOs)
			}
		}
	}
	// hook into config.go
	configFile := fmt.Sprintf("%s/config/config.go", ep.ProjectFile.FullPath)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("%s is missing unable to write in hooks\n", configFile)
	} else {
		// env lines
		configLines := []string{}
		split := strings.Split(ep.ProjectFile.Storages, " ")
		for _, s := range split {
			if string(s[0]) == "s" {
				configLines = append(configLines, "StorageSQL = true")
				if string(s[1]) == "p" || string(s[1]) == "m" {
					configLines = append(configLines, "DBUser = getEnvOrDefault(\"{{.ProjectFile.Name.EnvVar}}_DB_USER\",\"\")")
					configLines = append(configLines, "DBPass = getEnvOrDefault(\"{{.ProjectFile.Name.EnvVar}}_DB_PASS\", \"\")")
					configLines = append(configLines, "DBDB = getEnvOrDefault(\"{{.ProjectFile.Name.EnvVar}}_DB_DB\", \"\")")
					configLines = append(configLines, "DBHost = getEnvOrDefault(\"{{.ProjectFile.Name.EnvVar}}_DB_HOST\", \"\")")
				} else {
					configLines = append(configLines, "SqlitePath = getEnvOrDefault(\"{{.ProjectFile.Name.EnvVar}}_SQLITE_PATH\",\"\")")
				}
			}
			if string(s[0]) == "f" {
				configLines = append(configLines, "StorageFile = true")
				configLines = append(configLines, "SqlitePath = getEnvOrDefault(\"{{.Name.Upper}}_SQLITE_PATH\", \"/tmp/{{.Name.Lower}}.db\")")
			}
			if string(s[0]) == "m" {
				configLines = append(configLines, "StorageMongo = true")
				configLines = append(configLines, "MongoHost = getEnvOrDefault(\"{{.Name.Upper}}_MONGO_HOST\", \"localhost\"")
				configLines = append(configLines, "MongoPort = getEnvOrDefault(\"{{.Name.Upper}}_MONGO_PORT\", \"27017\"")
			}
		}
		configLine := strings.Join(configLines, "\n\t")

		var configReplace bytes.Buffer
		tConfig := template.Must(template.New("config").Parse(configLine))
		errConfig := tConfig.Execute(&configReplace, ep)
		if errConfig != nil {
			fmt.Printf("%s: template error [%s]\n", configFile, errConfig)
		} else {
			cmdConfig := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace config text - do not remove ---/%s/g' %s`, configReplace.String(), configFile)
			execConfig := exec.Command("bash", "-c", cmdConfig)
			errConfigCmd := execConfig.Run()
			if errConfigCmd != nil {
				fmt.Printf("%s: error in replace for config text [%s]\n", configFile, errConfigCmd)
			}
		}
	}
}
