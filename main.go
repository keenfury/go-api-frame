package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var (
	clearMap map[string]func()
)

func init() {
	clearMap = make(map[string]func())
	clearMap["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clearMap["darwin"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clearMap["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func main() {
	// main reader for the command line
	reader := bufio.NewReader(os.Stdin)
	clearScreen()
	fmt.Println("*** Go API Frame Helper ***")

	// this will be the main struct to hold all information passed to other processes
	project := Project{ParseInput: ParseInput, Reader: reader}

	// see if projectFile is able to load
	exists := project.LoadProjectFile()
	if !exists {
		// if not create new one, save values
		exists = project.CreateProjectFile()
	}
	if exists {
		// defer save project file
		defer project.SaveProjectFile()
	} else {
		fmt.Println("Quitting app, remove .frame is present, restart app")
		project.ParseInput(project.Reader)
	}

	if project.ProjectFile.SaveStorage == false {
		project.ProjectFile.UseORM, project.ProjectFile.SaveStorage, project.ProjectFile.Storages = storageMenu(reader)
	}
	project.DetermineMenu()
	project.Protoc()
	project.Fmt()
	project.Generate()
	fmt.Printf("\n*** Remember to 'go get -u all && go mod tidy' ***\n\n")
	fmt.Println("Bye!")
}

func ParseInput(reader *bufio.Reader) string {
	s, _ := reader.ReadString('\n')
	s = strings.TrimSpace(s)
	return s
}

func clearScreen() {
	clearFunc, ok := clearMap[runtime.GOOS]
	if !ok {
		fmt.Println("\n *** Your platform is not supported to clear the terminal screen ***")
		return
	}
	clearFunc()
}

func (n *Name) NameConverter() {
	lower := strings.ToLower(n.RawName) // CamelCase => camelcase; Snake_Case => snake_case;
	camel := lower
	i := strings.Index(camel, "_")
	if i == -1 {
		camel = fmt.Sprintf("%s%s", strings.ToUpper(string(camel[0])), camel[1:])
	} else {
		for i > -1 {
			firstCamel := ""
			firstRestOf := ""
			afterUnderscoreCamel := ""
			afterUnderscoreRestOf := ""
			if i == 0 {
				camel = camel[1:]
			} else {
				firstCamel = strings.ToUpper(string(camel[0]))
				firstRestOf = camel[1:i]
				if i+1 < len(camel) {
					afterUnderscoreCamel = strings.ToUpper(string(camel[i+1]))
					afterUnderscoreRestOf = camel[i+2:]
				}
				camel = fmt.Sprintf("%s%s%s%s", firstCamel, firstRestOf, afterUnderscoreCamel, afterUnderscoreRestOf)
			}
			i = strings.Index(camel, "_")
		} // snake_case => SnakeCase; _snake_case_ => SnakeCase
	}
	camelLower := fmt.Sprintf("%s%s", strings.ToLower(string(camel[0])), camel[1:]) // SnakeCase => snakeCase
	abbr := lower
	if len(lower) > 2 {
		abbr = lower[:3] // snake_case => sna
	}
	n.Abbr = abbr
	n.Camel = camel
	n.LowerCamel = camelLower
	n.Lower = lower
	n.AllLower = strings.ToLower(camel)
	n.Upper = strings.ToUpper(camel)
	n.EnvVar = strings.ToUpper(n.RawName)
}
