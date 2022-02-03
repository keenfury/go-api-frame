package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

func (p *Project) File() {
	clearScreen()
	cont := AskYesOrNo(p.Reader, "This will create a file to read in the table schemas. Is this what you want (y/n)? ")
	if !cont {
		return
	}
	files, errReadDir := ioutil.ReadDir(p.ProjectFile.FullPath)
	if errReadDir != nil {
		fmt.Println(errReadDir)
		return
	}
	sqlFiles := []string{}
	for _, file := range files {
		if strings.Contains(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sqlFile := ""
	if len(sqlFiles) > 1 {
		fmt.Println("")
		for {
			for i, file := range sqlFiles {
				fmt.Printf("%d - %s\n", i+1, file)
			}
			fmt.Println("q - quit")
			fmt.Println("")
			fmt.Print("Choose your file: ")
			selection := p.ParseInput(p.Reader)
			if strings.ToLower(selection) == "q" {
				clearScreen()
				return
			}
			selectionNumber, errParse := strconv.ParseInt(selection, 10, 64)
			if errParse != nil {
				fmt.Println("Not a number, you can do better than that!")
				continue
			}
			if selectionNumber < 1 {
				fmt.Println("Not a valid selection, you can do better than that!")
				continue
			}
			if selectionNumber > int64(len(sqlFiles)) {
				fmt.Println("Not a valid seleciton, you can do better than that!")
				continue
			}
			sqlFile = sqlFiles[selectionNumber-1]
			break
		}
	} else if len(sqlFiles) == 1 {
		sqlFile = sqlFiles[0]
	}
	if sqlFile == "" {
		//clearScreen()
		fmt.Print("Please enter full path to your sql file to parse or q - quit: ")
		sqlFile = p.ParseInput(p.Reader)
		if strings.ToLower(sqlFile) == "q" {
			return
		}
	}
	if sqlFile != "" {
		bContent, errRead := ioutil.ReadFile(sqlFile)
		if errRead != nil {
			fmt.Println(errRead)
			return
		}
		bArray := bytes.Split(bContent, []byte("\n"))
		sql := Sql{}
		tableCount := 0
		for _, bLine := range bArray {
			if len(bLine) != 0 {
				sql.RawSql = append(sql.RawSql, string(bLine))
			} else {
				p.ProcessSql(sql)
				tableCount++
				sql = Sql{}
			}
		}
		p.ProcessSql(sql)
		tableCount++
		fmt.Println("")
		fmt.Printf("Processed %d tables, press any key to continue", tableCount)
		p.ParseInput(p.Reader)
		return
	}
	fmt.Println("We didn't do anything!")
}

func (p *Project) ProcessSql(sql Sql) error {
	endPoint := EndPoint{}
	columns := []Column{}
	tableName := ""
	errParse, foundNull := sql.ParseLine(&tableName, &columns)
	if errParse != nil {
		fmt.Println(errParse)
		return errParse
	}
	endPoint.Name.RawName = tableName
	endPoint.Name.NameConverter()
	endPoint.Columns = columns
	endPoint.SqlLines = sql
	endPoint.HaveNullColumns = foundNull
	p.EndPoints = append(p.EndPoints, endPoint)
	return nil
}
