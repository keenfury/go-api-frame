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
	// server & common hooks
	ep.BuildAPIHooks()
}

func (ep *EndPoint) BuildModelTemplate() {
	cArray := []string{}
	for _, c := range ep.Columns {
		cArray = append(cArray, fmt.Sprintf(MODEL_COLUMN, c.ColumnName.Camel, c.GoType, c.ColumnName.Lower, c.ColumnName.Lower))
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
			initStorage = append(initStorage, fmt.Sprintf("\tif config.StorageMongo {\n\t\treturn &Mongo%s{}\n\t}", ep.Camel))
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
	putSearch := ""
	setArgs := ""
	foundOneKey := false
	for _, c := range ep.Columns {
		if c.PrimaryKey {
			if c.GoType == "string" {
				if foundOneKey {
					setArgs += ", "
					putSearch += "\n"
					getDeleteRow += "\n"
				}
				getDeleteRow += fmt.Sprintf(MANAGER_GET_STRING, ep.Abbr, c.ColumnName.Camel, c.ColumnName.Camel)
				putSearch += fmt.Sprintf(MANAGER_PUT_SEARCH_STRING, c.ColumnName.Lower, c.ColumnName.Camel, c.ColumnName.Lower, c.ColumnName.Camel, c.ColumnName.Camel)
				foundOneKey = true
			}
			if c.GoType == "int" {
				if foundOneKey {
					setArgs += ", "
					putSearch += "\n"
				}
				getDeleteRow += fmt.Sprintf(MANAGER_GET_INT, ep.Abbr, c.ColumnName.Camel, c.ColumnName.Camel)
				putSearch += fmt.Sprintf(MANAGER_PUT_SEARCH_INT, c.ColumnName.Lower, c.ColumnName.Camel, c.ColumnName.Lower, c.ColumnName.Camel, c.ColumnName.Camel, c.ColumnName.Lower, c.ColumnName.Lower)
				foundOneKey = true
			}
			setArgs += fmt.Sprintf("%s: %s", c.ColumnName.Camel, c.ColumnName.Lower)
		}
	}
	putRow = putSearch + fmt.Sprintf(MANAGER_PUT_STRUCT_STMT, ep.Abbr, ep.Camel, setArgs) + fmt.Sprintf(MANAGER_PUT_GET_STMT, ep.Abbr)
	for _, c := range ep.Columns {
		// post rows
		// if c.DBType == "autoincrement" || c.DBType == "serial" {
		// 	postRow += fmt.Sprintf(MANAGER_POST_AUTOINCREMENT, ep.Abbr, c.ColumnName.Lower, c.ColumnName.Camel)
		// } else {
		if c.GoType == "string" || c.GoType == "null.String" {
			if c.DBType == "uuid" {
				postRow += fmt.Sprintf(MANAGER_POST_UUID, ep.Abbr, c.ColumnName.Camel, c.ColumnName.Camel)
			} else {
				postRow += fmt.Sprintf(MANAGER_POST_VARCHAR_LEN, ep.Abbr, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel, c.Length, c.ColumnName.Camel, c.Length)
			}
		} else {
			if !c.Null && !c.PrimaryKey {
				postRow += fmt.Sprintf(MANAGER_POST_NULL, ep.Abbr, c.ColumnName.Camel, c.ColumnName.Camel)
			}
		}
		// }
		// put rows
		if c.GoType == "null.String" {
			if !c.PrimaryKey {
				putRow += fmt.Sprintf(MANAGER_PUT_STRING_NULL_ASSIGN, c.ColumnName.Camel, c.ColumnName.LowerCamel, c.ColumnName.Camel, c.ColumnName.Lower, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel, c.ColumnName.LowerCamel)
			}
		}
		if c.GoType == "int" || c.GoType == "null.Int" {
			if !c.PrimaryKey {
				putRow += fmt.Sprintf(MANAGER_PUT_INT_NULL_ASSIGN, c.ColumnName.Camel, c.ColumnName.LowerCamel, c.ColumnName.Camel, c.ColumnName.Lower, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel, c.ColumnName.LowerCamel)
			}
		}
		if c.GoType == "null.Float" {
			putRow += fmt.Sprintf(MANAGER_PUT_FLOAT_NULL_ASSIGN, c.ColumnName.Camel, c.ColumnName.LowerCamel, c.ColumnName.Camel, c.ColumnName.Lower, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel, c.ColumnName.LowerCamel)
		}
		if c.GoType == "null.Bool" {
			putRow += fmt.Sprintf(MANAGER_PUT_BOOL_NULL_ASSIGN, c.ColumnName.Camel, c.ColumnName.LowerCamel, c.ColumnName.Camel, c.ColumnName.Lower, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel, c.ColumnName.LowerCamel)
		}
		if c.GoType == "*json.RawMessage" {
			putRow += fmt.Sprintf(MANAGER_PUT_JSON_NULL_ASSIGN, c.ColumnName.Camel, c.ColumnName.Lower, c.ColumnName.LowerCamel, c.ColumnName.Lower, c.ColumnName.LowerCamel, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel, c.ColumnName.LowerCamel)
		}
		if c.GoType == "null.Time" {
			putRow += fmt.Sprintf(MANAGER_PUT_TIME_NULL_ASSIGN, c.ColumnName.Camel, c.ColumnName.LowerCamel, c.ColumnName.Camel, c.ColumnName.Lower, c.ColumnName.Camel, c.ColumnName.LowerCamel, c.ColumnName.LowerCamel, c.ColumnName.Camel, ep.Abbr, c.ColumnName.Camel, c.ColumnName.LowerCamel)
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
	foundOnePut := false
	foundOnePost := false
	keys := ""
	putKeys := ""
	keyCount := 1
	values := ""
	listOrder := ""
	postColumn := ""
	postColumnNames := ""
	putColumn := ""
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
			putKeys += fmt.Sprintf("%s = :%s", c.ColumnName.Lower, c.ColumnName.Lower)
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
			if foundOnePut {
				putColumn += ",\n\t\t\t"
			}
			putColumn += fmt.Sprintf("%s = :%s", c.ColumnName.Lower, c.ColumnName.Lower)
			foundOnePut = true
		}
	}
	ep.SqlGetColumns = strings.TrimRight(SqlGetColumns, "\n")
	ep.SqlTableKeyKeys = keys
	ep.SqlTableKeyValues = values
	ep.SqlTableKeyListOrder = listOrder
	ep.SqlPostColumns = strings.TrimRight(postColumn, "\n")
	ep.SqlPostColumnsNamed = strings.TrimRight(postColumnNames, "\n")
	ep.SqlPutColumns = strings.TrimRight(putColumn, "\n")
	ep.SqlPutWhere = putKeys
	if foundSerial != "" {
		ep.SqlPostReturning = fmt.Sprintf(" returning %s", foundSerialDB)
		ep.SqlPostLastId = fmt.Sprintf(DATA_LAST_ID, ep.Abbr, foundSerial)
	}
	ep.FileKeys = strings.Join(fileKey, " && ")
	ep.FileGetColumns = strings.Join(fileGetColumn, "\n\t\t\t")
	ep.FilePostIncr = fmt.Sprintf("%s\n\tfor _, %sObj := range %ss {\n%s\n\t}\n%s", strings.Join(filePostIncrInit, "\n"), ep.Abbr, ep.Abbr, strings.Join(filePostIncrCheck, "\n"), strings.Join(filePostIncr, "\n"))
}

func (ep *EndPoint) BuildAPIHooks() {
	// hook into server file
	serverFile := fmt.Sprintf("%s/cmd/%s/main.go", ep.ProjectFile.FullPath, ep.ProjectFile.AppName)
	if _, err := os.Stat(serverFile); os.IsNotExist(err) {
		fmt.Printf("%s is missing unable to write in hooks\n", serverFile)
	} else {
		var serverReplace bytes.Buffer
		tServer := template.Must(template.New("server").Parse(SERVER_ROUTE))
		errServer := tServer.Execute(&serverReplace, ep)
		if errServer != nil {
			fmt.Printf("%s: template error [%s]\n", serverFile, errServer)
		} else {
			cmdServer := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace server text - do not remove ---/%s/g' %s`, serverReplace.String(), serverFile)
			execServer := exec.Command("bash", "-c", cmdServer)
			errServerCmd := execServer.Run()
			if errServerCmd != nil {
				fmt.Printf("%s: error in replace for server [%s]\n", serverFile, errServerCmd)
			}
		}
		var mainReplace bytes.Buffer
		tMain := template.Must(template.New("server").Parse(MAIN_COMMON_PATH))
		errServer = tMain.Execute(&mainReplace, ep)
		if errServer != nil {
			fmt.Printf("%s: template error [%s]\n", serverFile, errServer)
		} else {
			cmdServer := fmt.Sprintf(`perl -pi -e 's/\/\/ --- replace server header text ---/%s/g' %s`, mainReplace.String(), serverFile)
			execServer := exec.Command("bash", "-c", cmdServer)
			errServerCmd := execServer.Run()
			if errServerCmd != nil {
				fmt.Printf("%s: error in replace for main [%s]\n", serverFile, errServerCmd)
			}
		}
	}
	// hook into common file
	commonFile := fmt.Sprintf("%s/%s/common.go", ep.ProjectFile.FullPath, ep.SubDir)
	if _, err := os.Stat(commonFile); os.IsNotExist(err) {
		commonSrc := fmt.Sprintf("%s/templates/common.go", os.Getenv("SCAFFOLDING_PATH"))
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
}
