package main

import (
	"fmt"
)

func (p *Project) Blank() {
	clearScreen()
	fmt.Print("This will create a 'Blank' end-point file. Is this what you want (y/n)? ")
	cont := false
	for {
		sel := ParseInput(p.Reader)
		switch sel {
		case "y", "Y":
			cont = true
		case "n", "N":
			cont = false
		default:
			fmt.Println("Invalid value (y/n) only")
			continue
		}
		if !cont {
			return
		}
		break
	}
	name := ""
	for {
		fmt.Print("What name would you like to call this endpoint (need to be > 3 characters)? ")
		name = ParseInput(p.Reader)
		if len(name) < 3 {
			fmt.Println("Longer name than that!")
		} else {
			break
		}
	}
	// only creating one endpoint, so manually create and set one endpoint array
	p.EndPoints = []EndPoint{{Name: Name{RawName: name}}}
	p.EndPoints[0].Name.NameConverter()
	p.UseBlank = true
}
