package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/tabwriter"
	"text/template"
)

var tmplFiles = []string{"model", "handler", "manager", "grpc"} // TODO: sql is optional depending on which data storages they want to template, dynamically build this

func (p *Project) LoadProjectFile() bool {
	if _, errStat := os.Stat("./.scaffold"); os.IsNotExist(errStat) {
		return false
	}
	bContent, errRead := os.ReadFile("./.scaffold")
	if errRead != nil {
		fmt.Printf("Error reading from .scaffold: %s", errRead)
		return true
	}
	errUnmarshal := json.Unmarshal(bContent, &p.ProjectFile)
	if errUnmarshal != nil {
		fmt.Printf("Error extracting data from .scaffold: %s", errUnmarshal)
		return true
	}
	return true
}

func (p *Project) SaveProjectFile() {
	bContent, errMarshal := json.MarshalIndent(p.ProjectFile, "", "    ")
	if errMarshal != nil {
		fmt.Println("Saving project file: unable to save -", errMarshal)
		return
	}
	errSave := os.WriteFile("./.scaffold", bContent, 0644)
	if errSave != nil {
		fmt.Println("Saving project file: unable to save -", errSave)
	}
}

func (p *Project) CreateProjectFile() bool {
	pwd, errPwd := os.Getwd()
	if errPwd != nil {
		fmt.Println("Unable to get present working directory:", errPwd)
		return false
	}
	fmt.Printf("Your current directory is: %s\n", pwd)
	msg := fmt.Sprint("Is this the project directory you want (y/n)? ")
	result := AskYesOrNo(p.Reader, msg)
	if !result {
		fmt.Println("You choose not to use the current diretory")
		return false
	}
	p.ProjectFile.Message = "This is used for the api frame program for convenience"
	p.ProjectFile.FullPath = pwd
	p.ProjectFile.AppName = path.Base(pwd)
	// ask for subdir
	fmt.Print("Which sub folder to save endpoint files (v1, routes, etc)? ")
	p.ProjectFile.SubDir = ParseInput(p.Reader)
	// create projectpath and subpackage
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		fmt.Println("***** GOPATH is not set, some of your paths will not be correct! *****")
	} else {
		idx := len(fmt.Sprintf("%s/src/", goPath))
		p.ProjectFile.ProjectPath = p.ProjectFile.FullPath[idx:]
	}
	subPath := strings.Split(p.ProjectFile.SubDir, "/")
	p.ProjectFile.SubPackage = subPath[len(subPath)-1]
	// encode paths
	p.ProjectFile.ProjectPathEncoded = strings.Replace(p.ProjectFile.ProjectPath, `/`, `\/`, -1)
	p.ProjectFile.SubDirEncoded = strings.Replace(p.ProjectFile.SubDir, `/`, `\/`, -1)
	return true
}

func (p *Project) GetNames() (names []string) {
	subDir := fmt.Sprintf("%s/%s", p.ProjectFile.ProjectPath, p.ProjectFile.SubDir)
	files, errReadDir := os.ReadDir(subDir)
	if errReadDir != nil {
		fmt.Println("Unable to read sub dir")
		return
	}
	for _, f := range files {
		names = append(names, f.Name())
	}
	return
}

func (p *Project) DetermineMenu() {
	sqlType := ""
	storages := strings.Split(p.ProjectFile.Storages, " ")
	for _, storage := range storages {
		if string(storage[0]) == "s" {
			sqlType = string(storage[1])
			break
		}
	}
	if sqlType != "" {
		SqlMenu(p, sqlType)
	} else {
		PromptMenu(p)
	}
}

func (p *Project) PrintBasicColumns(cols []Column) {
	fmt.Println("--- Saved Columns ---")
	tab := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(tab, "|Name\t|Type\t|Primary Key")
	fmt.Fprintln(tab, "+----\t+----\t+-----------")
	for _, c := range cols {
		fmt.Fprintf(tab, " %s\t %s\t %t\n", c.ColumnName.Camel, c.GoType, c.PrimaryKey)
	}
	tab.Flush()
	fmt.Println("")
}

func (p *Project) ProcessTemplates() {
	storageSavePath := fmt.Sprintf("%s/internal/storage", p.ProjectFile.FullPath)
	if errMakeAll := os.MkdirAll(storageSavePath, os.ModeDir|os.ModePerm); errMakeAll != nil {
		fmt.Println("New storage folder was not able to be made", errMakeAll)
		return
	}
	for _, ep := range p.EndPoints {
		storageFiles := []string{}
		savePath := fmt.Sprintf("%s/%s/%s", p.ProjectFile.FullPath, p.ProjectFile.SubDir, ep.AllLower)
		if _, err := os.Stat(savePath); !os.IsNotExist(err) {
			fmt.Println("Endpoint name already exists, skipping!")
			continue
		}
		if errMakeAll := os.MkdirAll(savePath, os.ModeDir|os.ModePerm); errMakeAll != nil {
			fmt.Println("New path was not able to be made", errMakeAll)
			return
		}
		templatePath := fmt.Sprintf("%s/templates", os.Getenv("FRAME_PATH"))
		blankInsert := ""
		if p.UseBlank {
			blankInsert = "_blank"
		}
		ep.ProjectFile = p.ProjectFile
		// determine which "storage" files to template too
		split := strings.Split(p.ProjectFile.Storages, " ")
		for _, s := range split {
			if ep.UseORM {
				tmplFiles = append(tmplFiles, "gorm")
				storageFiles = append(storageFiles, "gorm")
			}
			switch string(s[0]) {
			case "s":
				if !ep.UseORM {
					tmplFiles = append(tmplFiles, "sql")
				}
				switch string(s[1]) {
				case "m":
					ep.SQLProvider = MYSQL
					ep.SQLProviderLower = MYSQLLOWER
					ep.SQLProviderConnection = fmt.Sprintf("%sConnection", MYSQL)
					storageFiles = append(storageFiles, "mysql")
				case "p":
					ep.SQLProvider = POSTGRESQL
					ep.SQLProviderLower = POSTGRESQLLOWER
					ep.SQLProviderConnection = fmt.Sprintf("%sConnection", POSTGRESQL)
					storageFiles = append(storageFiles, "psql")
				case "s":
					ep.SQLProvider = SQLITE3
					ep.SQLProviderLower = SQLITE3LOWER
					ep.SQLProviderConnection = fmt.Sprintf("%sConnection", SQLITE3)
					storageFiles = append(storageFiles, "sqlite")
				}
			case "f":
				tmplFiles = append(tmplFiles, "file")
				storageFiles = append(storageFiles, "file")
			case "m":
				tmplFiles = append(tmplFiles, "mongo")
				storageFiles = append(storageFiles, "mongo")
			}
		}
		ep.BuildTemplateParts()
		for _, tmpl := range tmplFiles {
			tmplPath := fmt.Sprintf("%s/%s/%s%s.tmpl", templatePath, tmpl, tmpl, blankInsert)
			t, errParse := template.ParseFiles(tmplPath)
			if errParse != nil {
				fmt.Printf("Template could not parse file: %s; %s", tmplPath, errParse)
				fmt.Println("Exiting...")
				return
			}
			newFileName := fmt.Sprintf("%s/%s.go", savePath, tmpl)
			//fmt.Println("New file", newFileName, "creating...")
			file, err := os.Create(newFileName)
			if err != nil {
				fmt.Println("File:", tmpl, "was not able to be created", err)
				fmt.Println("Exiting...")
				return
			}
			err = t.Execute(file, ep)
			if err != nil {
				fmt.Println("Execution of template:", err)
			}
		}
		// save storage
		for _, tmpl := range storageFiles {
			tmplPath := fmt.Sprintf("%s/storage/%s.tmpl", templatePath, tmpl)
			t, errParse := template.ParseFiles(tmplPath)
			if errParse != nil {
				fmt.Printf("Template storage could not parse file: %s; %s", tmplPath, errParse)
				fmt.Println("Exiting...")
				return
			}
			newFileName := fmt.Sprintf("%s/%s.go", storageSavePath, tmpl)
			// don't over-write if already there
			if _, err := os.Stat(newFileName); !os.IsNotExist(err) {
				fmt.Printf("already exists: %s, skipping\n", newFileName)
				continue
			}
			file, err := os.Create(newFileName)
			if err != nil {
				fmt.Println("File:", tmpl, "was not able to be created", err)
				fmt.Println("Exiting...")
				return
			}
			err = t.Execute(file, ep)
			if err != nil {
				fmt.Println("Execution of template:", err)
			}
		}
	}
}

func (p *Project) Protoc() {
	cmd := fmt.Sprintf("cd pkg/proto && protoc --go_out=. --go_opt=paths=source_relative    --go-grpc_out=. --go-grpc_opt=paths=source_relative %s.proto", p.ProjectFile.AppName)
	execProto := exec.Command("bash", "-c", cmd)
	errProtoCmd := execProto.Run()
	if errProtoCmd != nil {
		fmt.Printf("Error executing protoc command: %s", errProtoCmd)
	}
}

func (p *Project) Fmt() {
	cmd := "go fmt ./..."
	execFmt := exec.Command("bash", "-c", cmd)
	errFmtCmd := execFmt.Run()
	if errFmtCmd != nil {
		fmt.Printf("Error executing fmt command: %s", errFmtCmd)
	}
}
