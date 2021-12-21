package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

func (p *Project) Prompt() {
	clearScreen()
	cont := AskYesOrNo(p.Reader, "This will create a prompt for the table structure. Is this what you want (y/n)? ")
	if !cont {
		return
	}
	for {
		endPoint := EndPoint{}
		name := ""
		for {
			fmt.Println("")
			fmt.Print("What name would you like to call this endpoint (need to be > 2 characters)? ")
			name = ParseInput(p.Reader)
			if len(name) < 3 {
				fmt.Println("Longer name than that!")
			} else {
				break
			}
		}
		// check if any endpoint doesn't have the same name
		names := p.GetNames()
		foundName := false
		for _, n := range names {
			if n == name {
				foundName = true
			}
		}
		if foundName {
			fmt.Println("Endpoint already exists")
			continue
		}
		colName := Name{RawName: name}
		colName.NameConverter()
		endPoint.Name = colName
		cols := []Column{}
		for {
			clearScreen()
			p.PrintSqlColumns(cols)
			col := Column{}
			fmt.Print("Column Name: ")
			name := ParseInput(p.Reader)
			col.ColumnName.RawName = name
			col.ColumnName.NameConverter()
			messages := []string{"Column DB Type:"}
			prompts := []string{"(1) Varchar", "(2) Decimal", "(3) Integer", "(4) Timestamp", "(5) Boolean", "(6) Json", "(7) UUID", "(8) Auto Increment", "(9) Text", "(10) Char", "(11) Date"}
			acceptablePrompt := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "12"}

			sel := BasicPrompt(p.Reader, messages, prompts, acceptablePrompt, "")
			if strings.ToLower(sel) == "" {
				fmt.Println("Empty, try again!  Press enter to continue")
				ParseInput(p.Reader)
				continue
			}
			switch sel {
			case "1":
				col.DBType = "varchar"
				col.GoType = "null.String"
				col.GoTypeNonSql = "string"
				col.Length = AskLength("What is the varchar length? ", p.Reader)
				nullAble, useDefault := NullDefaultQuestion(true, true, p.Reader)
				if useDefault {
					fmt.Println("What is your varchar default value? ")
					col.DefaultValue = ParseInput(p.Reader)
				}
				if nullAble {
					col.Null = true
				}
				endPoint.HaveNullColumns = true
			case "2":
				col.DBType = "numeric"
				col.GoType = "null.Float"
				col.GoTypeNonSql = "float64"
				nullAble, useDefault := NullDefaultQuestion(true, true, p.Reader)
				if useDefault {
					for {
						fmt.Println("What is your numeric default value? ")
						numStr := ParseInput(p.Reader)
						_, errParse := strconv.ParseFloat(numStr, 64)
						if errParse != nil {
							fmt.Println("Not a numeric/float, you can do better than that!")
							continue
						}
						col.DefaultValue = numStr
						break
					}
				}
				if nullAble {
					col.Null = true
				}
				endPoint.HaveNullColumns = true
			case "3":
				col.DBType = "int"
				col.GoType = "null.Int"
				col.GoTypeNonSql = "int"
				nullAble, useDefault := NullDefaultQuestion(true, true, p.Reader)
				if useDefault {
					for {
						fmt.Printf("What is your integer default value? ")
						intStr := ParseInput(p.Reader)
						_, errParse := strconv.ParseInt(intStr, 10, 64)
						if errParse != nil {
							fmt.Println("Not a integer, you can do better than that!")
							continue
						}
						col.DefaultValue = intStr
						break
					}
				}
				if nullAble {
					col.Null = true
				}
				endPoint.HaveNullColumns = true
			case "4":
				col.DBType = "timestamp"
				col.GoType = "null.Time"
				col.GoTypeNonSql = "time.Time"
				nullAble, useDefault := NullDefaultQuestion(true, true, p.Reader)
				if useDefault {
					for {
						if AskYesOrNo(p.Reader, "Default timestamp only make sense for 'Now', is that what you want (y/n)? ") {
							col.DefaultValue = "now()"
						}
						break
					}
				}
				if nullAble {
					col.Null = true
				}
				endPoint.HaveNullColumns = true
			case "5":
				col.DBType = "bool"
				col.GoType = "null.Bool"
				col.GoTypeNonSql = "bool"
				nullAble, useDefault := NullDefaultQuestion(true, true, p.Reader)
				if useDefault {
					defAnswer := AskYesOrNo(p.Reader, "What is your bool default value (y/n)? ")
					col.DefaultValue = fmt.Sprintf("%t", defAnswer)
				}
				if nullAble {
					col.Null = true
				}
				endPoint.HaveNullColumns = true
			case "6":
				col.DBType = "json"
				col.GoType = "*json.RawMessage"
				col.GoTypeNonSql = "string"
				_, useDefault := NullDefaultQuestion(false, true, p.Reader)
				if useDefault {
					for {
						fmt.Println("What is your json default value? ")
						jsonStr := ParseInput(p.Reader)
						if !json.Valid([]byte(jsonStr)) {
							fmt.Println("Not valid json string, you can do better than that, try. (Note: try '{}')")
							continue
						}
						col.DefaultValue = jsonStr
						break
					}
				}
			case "7":
				col.DBType = "uuid"
				col.GoType = "string"
				col.GoTypeNonSql = "string"
				col.Null = false
			case "8":
				col.DBType = "autoincrement"
				col.GoType = "int"
				col.GoTypeNonSql = "int"
				col.Null = false
			case "9":
				col.DBType = "text"
				col.GoType = "null.String"
				col.GoTypeNonSql = "string"
				nullAble, _ := NullDefaultQuestion(true, false, p.Reader)
				if nullAble {
					col.Null = true
				}
				endPoint.HaveNullColumns = true
			case "10":
				col.DBType = "char"
				col.GoType = "null.String"
				col.GoTypeNonSql = "string"
				col.Length = AskLength("What is the char length? ", p.Reader)
				nullAble, useDefault := NullDefaultQuestion(true, true, p.Reader)
				if useDefault {
					fmt.Println("What is your char default value? ")
					col.DefaultValue = ParseInput(p.Reader)
				}
				if nullAble {
					col.Null = true
				}
				endPoint.HaveNullColumns = true
			case "11":
				col.DBType = "date"
				col.GoType = "null.Time"
				col.GoTypeNonSql = "time.Time"
				nullAble, _ := NullDefaultQuestion(true, false, p.Reader)
				if nullAble {
					col.Null = true
				}
				endPoint.HaveNullColumns = true
			}
			col.PrimaryKey = AskYesOrNo(p.Reader, "This column a primary key (y/n)? ")

			cols = append(cols, col)
			addAnother := AskYesOrNo(p.Reader, "Add another column (y/n)? ")
			if !addAnother {
				break
			}
		}
		endPoint.Columns = cols
		p.EndPoints = append(p.EndPoints, endPoint)
		addAnotherEP := AskYesOrNo(p.Reader, "Add another EndPoint (y/n)? ")
		if !addAnotherEP {
			break
		}
		clearScreen()
	}
	p.SaveOutSql()
}

func (p *Project) PrintSqlColumns(cols []Column) {
	fmt.Println("--- Saved Columns ---")
	tab := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(tab, "|Name\t|Type\t|Default Value\t|Null\t|Primary Key")
	fmt.Fprintln(tab, "+----\t+----\t+-------------\t+----\t+-----------")
	for _, c := range cols {
		fmt.Fprintf(tab, " %s\t %s\t %s\t %t\t %t\n", c.ColumnName.Camel, c.DBType, c.DefaultValue, c.Null, c.PrimaryKey)
	}
	tab.Flush()
	fmt.Println("")
}

func (p *Project) SaveOutSql() {
	sqlProvider := ""
	split := strings.Split(p.ProjectFile.Storages, " ")
	for _, s := range split {
		switch string(s[0]) {
		case "s":
			switch string(s[1]) {
			case "m":
				sqlProvider = MYSQL
			case "p":
				sqlProvider = POSTGRESQL
			case "s":
				sqlProvider = SQLITE3
			}
		}
	}
	fileName := "./prompt_schema"
	lines := []string{}
	for e, ep := range p.EndPoints {
		primaryKeys := []string{}
		lines = append(lines, fmt.Sprintf("create table if not exists %s (", ep.Name.Lower))
		for i, c := range ep.Columns {
			null := " null"
			defaultValue := ""
			length := ""
			if !c.Null {
				null = " not null"
			}
			dbType := c.DBType
			if dbType == "autoincrement" || dbType == "serial" {
				null = ""
				if sqlProvider == SQLITE3 {
					dbType = "integer primary key autoincrement"
				}
				if sqlProvider == MYSQL {
					dbType = "integer auto_increment"
				}
				if sqlProvider == POSTGRESQL {
					dbType = "serial"
				}
			}
			if c.PrimaryKey && !(dbType == "autoincrement" && ep.SQLProvider == SQLITE3) {
				primaryKeys = append(primaryKeys, c.ColumnName.Lower)
			}
			if c.DefaultValue != "" {
				if c.DBType == "varchar" || c.DBType == "char" || c.DBType == "text" {
					defaultValue = fmt.Sprintf(" default '%s'", c.DefaultValue)
				} else {
					defaultValue = fmt.Sprintf(" default %s", c.DefaultValue)
				}
			}
			if c.Length > 0 {
				length = fmt.Sprintf("(%d)", c.Length)
			}
			if i < len(ep.Columns)-1 || (i == len(ep.Columns)-1 && len(primaryKeys) > 0) {
				lines = append(lines, fmt.Sprintf("\t%s %s%s%s%s,", c.ColumnName.Lower, dbType, length, null, defaultValue))
			} else {
				lines = append(lines, fmt.Sprintf("\t%s %s%s%s%s", c.ColumnName.Lower, dbType, length, null, defaultValue))
			}
		}
		if len(primaryKeys) > 0 {
			lines = append(lines, fmt.Sprintf("\tprimary key(%s)", strings.Join(primaryKeys, ",")))
		}
		lines = append(lines, ");")
		if e > 0 {
			lines = append(lines, "")
		}
	}
	lines = append(lines, "\n")
	if len(lines) > 0 {
		file, errOpen := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if errOpen != nil {
			fmt.Println("Unable to save schema to:", fileName)
			return
		}
		defer file.Close()
		if _, errWrite := file.WriteString(strings.Join(lines, "\n")); errWrite != nil {
			fmt.Println("Unable to write lines to:", fileName)
		}
	}
}
