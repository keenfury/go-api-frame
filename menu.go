package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

func storageMenu(reader *bufio.Reader) (bool, bool, string) {
	useORM := false
	storage := []string{}
	messages := []string{"SQL Options", "Choice which SQL implementation"}
	prompts := []string{"(P)ostgres", "(M)ysql", "(S)qlite3"}
	acceptablePrompts := []string{"p", "m", "s"}
	exitEarly := false
StorageLoop:
	for {
		fmt.Println("")
		fmt.Println("Storage Types")
		fmt.Println("Enter choices on one line with spaces between choices")
		fmt.Println("")
		fmt.Println("(S)QL")
		fmt.Println("(F)ile")
		fmt.Println("(M)ongoDB")
		fmt.Println("(E)xit")
		fmt.Println("")
		fmt.Println("")
		fmt.Printf("Select choices: ")
		sel := ParseInput(reader)
		if strings.ToLower(sel) == "e" {
			exitEarly = true
			break
		}
		if sel == "" {
			fmt.Print("Please enter at least one choice, press 'enter' to continue")
			ParseInput(reader)
			clearScreen()
			continue
		}
		split := strings.Split(sel, " ")
		// validate each one and make unique
		storageOptions := map[string]bool{"s": false, "f": false, "m": false}
		for _, S := range split {
			s := strings.ToLower(S)
			if _, ok := storageOptions[s]; ok {
				storageOptions[s] = true
				continue
			}
			fmt.Printf("Choice '%s' in not valid, press 'enter' to continue", S)
			ParseInput(reader)
			clearScreen()
			continue StorageLoop
		}
		// check for sql, only ask once after all be validated
		for k, v := range storageOptions {
			if v {
				if strings.ToLower(k) == "s" {
					sqlType := BasicPrompt(reader, messages, prompts, acceptablePrompts, "e")
					if sqlType == "e" {
						exitEarly = true
						break StorageLoop
					}
					storage = append(storage, "s"+sqlType)
					msg := "Would you like to use GORM (y/n)? "
					useORM = AskYesOrNo(reader, msg)
				} else {
					storage = append(storage, k)
				}
			}
		}
		break
	}

	if exitEarly {
		return false, false, ""
	}

	fmt.Println("")
	// msg := "Would you like to save these SQL options for the project? "
	save := true // AskYesOrNo(reader, msg)
	return useORM, save, strings.Join(storage, " ")
}

func SqlMenu(project *Project, sqlType string) {
	clearScreen()
	messages := []string{"*** SQL Storage ***", "If other storage types are wanting, we will use the sql import matching the type as best it can"}
	prompts := []string{"(1) Load File", "(2) Paste Table Syntax", "(3) Prompt For Table Syntax", "(4) Blank Struct"}
	acceptablePrompts := []string{"1", "2", "3", "4"}
	selection := BasicPrompt(project.Reader, messages, prompts, acceptablePrompts, "e")

	switch selection {
	case "1":
		project.File()
	case "2":
		project.Paste()
	case "3":
		project.Prompt()
	case "4":
		project.Blank()
	case "e", "E":
		return
	}
	project.ProcessTemplates()
}

func PromptMenu(project *Project) {
	messages := []string{"*** File/MongoDB Storage ***", "", "Field Type:"}
	prompts := []string{"(1) String", "(2) Integer", "(3) Decimal", "(4) Timestamp", "(5) Boolean", "(6) UUID"}
	acceptablePrompts := []string{"1", "2", "3", "4", "5", "6"}

	for {
		fmt.Print("Endpoint name: (e) to exit")
		name := Name{}
		name.RawName = ParseInput(project.Reader)
		if strings.ToLower(name.RawName) == "e" {
			break
		}
		name.NameConverter()
		endpoint := EndPoint{Name: name}
		for {
			clearScreen()
			project.PrintBasicColumns(endpoint.Columns)
			column := Column{}
			fmt.Print("Field Name: (e) to exit")
			column.ColumnName.RawName = ParseInput(project.Reader)
			column.ColumnName.NameConverter()
			if strings.ToLower(column.ColumnName.RawName) == "e" {
				break
			}
			selection := BasicPrompt(project.Reader, messages, prompts, acceptablePrompts, "e")

			switch selection {
			case "1":
				column.GoType = "string"
			case "2":
				column.GoType = "int"
			case "3":
				column.GoType = "float64"
			case "4":
				column.GoType = "time.Time"
			case "5":
				column.GoType = "bool"
			case "6":
				column.GoType = "string"
			case "e", "E":
				break
			}
			endpoint.Columns = append(endpoint.Columns, column)
			anotherColumn := AskYesOrNo(project.Reader, "Add another field?")
			if !anotherColumn {
				break
			}
		}
		project.EndPoints = append(project.EndPoints, endpoint)
		anotherEndpoint := AskYesOrNo(project.Reader, "Add another endpoint?")
		if !anotherEndpoint {
			break
		}
	}
	project.ProcessTemplates()
}

// helper functions
func BasicPrompt(reader *bufio.Reader, mainMessage []string, prompts []string, acceptablePrompts []string, exitString string) string {
	for {
		fmt.Println("")
		for _, msg := range mainMessage {
			fmt.Println(msg)
		}
		fmt.Println("")
		for _, prompt := range prompts {
			fmt.Println(prompt)
		}
		if exitString != "" {
			// just in case you don't want to show this line
			fmt.Printf("(%s) to exit", exitString)
		}
		fmt.Println("")
		fmt.Println("")
		fmt.Print("Selection Choice: ")
		selection := ParseInput(reader)
		if strings.ToLower(selection) == exitString {
			return exitString
		}
		found := false
		for _, acceptablePrompt := range acceptablePrompts {
			if strings.ToLower(selection) == acceptablePrompt {
				found = true
				break
			}
		}
		if !found {
			fmt.Print("Invalid selection, try again, press 'enter' to continue:")
			ParseInput(reader)
			clearScreen()
			continue
		}
		return strings.ToLower(selection)
	}
}

func AskYesOrNo(reader *bufio.Reader, msg string) (answer bool) {
	for {
		fmt.Print(msg)
		def := ParseInput(reader)
		switch def {
		case "y", "Y":
			answer = true
		case "n", "N":
			answer = false
		default:
			fmt.Println("Invalid value, get it together (y or n)!")
			continue
		}
		break
	}
	return
}

func NullDefaultQuestion(askNull, askDefault bool, reader *bufio.Reader) (nullAble, useDefault bool) {
	if askNull {
		nullAble = AskYesOrNo(reader, "Can this column be null (y/n)? ")
	}
	if !nullAble && askDefault {
		useDefault = AskYesOrNo(reader, "Does this column have a default value (y/n)? ")
	}
	return
}

func AskLength(msg string, reader *bufio.Reader) int64 {
	for {
		fmt.Print(msg)
		length := ParseInput(reader)
		lenInt, errParse := strconv.ParseInt(length, 10, 64)
		if errParse != nil {
			fmt.Println("Not an integer, you can can do better than that!")
			continue
		}
		return lenInt
	}
}
