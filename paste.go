package main

import (
	"fmt"
)

func (p *Project) Paste() {
	clearScreen()
	cont := AskYesOrNo(p.Reader, "This will create a prompt to paste a table schema. Is this what you want (y/n)? ")
	if !cont {
		return
	}
	for {
		endPoint := EndPoint{}
		sql := Sql{}
		clearScreen()
		fmt.Println("Paste your table structure below: ")
		for {
			line := ParseInput(p.Reader)
			if line == "" {
				break
			} else if line[:] == ")" || line[:] == ");" || line[len(line)-1:] == ";" {
				sql.RawSql = append(sql.RawSql, line)
				break
			}
			sql.RawSql = append(sql.RawSql, line)
		}
		columns := []Column{}
		tableName := ""
		errParse, foundNull := sql.ParseLine(&tableName, &columns)
		if errParse != nil {
			fmt.Println(errParse)
			break
		}
		endPoint.Name.RawName = tableName
		endPoint.Name.NameConverter()
		endPoint.Columns = columns
		endPoint.SqlLines = sql
		endPoint.HaveNullColumns = foundNull
		p.EndPoints = append(p.EndPoints, endPoint)
		cont := AskYesOrNo(p.Reader, "Paste another table schema (y/n)? ")
		if !cont {
			break
		}
	}
}
