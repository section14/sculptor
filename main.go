package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Entries []map[string]string

type Class struct {
	Name string
	Val  string
}

type Directions struct {
	Base   Class
	Top    Class
	Bottom Class
	Left   Class
	Right  Class
}

type CssFile struct {
	Vars  map[string]string
	Lines []string
}

func (c *CssFile) AddLine(line string) {
	c.Lines = append(c.Lines, line)
}

type Config struct {
	Theme   Entries `json:"theme"`
	Margin  Entries `json:"padding"`
	Padding Entries `json:"margin"`
	Border  Entries `json:"border"`
}

func main() {
	configName := os.Args[1]
    outputName := os.Args[2]

	file, err := os.Open(configName)
	if err != nil {
		log.Fatal(fmt.Sprintf("couldn't open %s", configName), err)
	}

	var config Config
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("couldn't decode config, check JSON format")
	}

	file.Close()

	css := &CssFile{Vars: make(map[string]string)}
	buildCss(config, css)

	//generate css file
	output, err := os.Create(outputName)
	if err != nil {
		log.Fatal("couldn't create CSS file", err)
	}
	defer output.Close()

	for _, c := range css.Lines {
		output.WriteString(c)
	}
}

func buildCss(c Config, css *CssFile) {
	padding := Directions{
		Base:   Class{Name: "p", Val: "padding"},
		Top:    Class{Name: "pt", Val: "padding-top"},
		Bottom: Class{Name: "pb", Val: "padding-bottom"},
		Left:   Class{Name: "pl", Val: "padding-left"},
		Right:  Class{Name: "pr", Val: "padding-right"},
	}

	margin := Directions{
		Base:   Class{Name: "m", Val: "margin"},
		Top:    Class{Name: "mt", Val: "margin-top"},
		Bottom: Class{Name: "mb", Val: "margin-bottom"},
		Left:   Class{Name: "ml", Val: "margin-left"},
		Right:  Class{Name: "mr", Val: "margin-right"},
	}

	border := Directions{
		Base:   Class{Name: "border", Val: "border-width"},
		Top:    Class{Name: "border-top", Val: "border-top-width"},
		Bottom: Class{Name: "border-bottom", Val: "border-bottom-width"},
		Left:   Class{Name: "border-left", Val: "border-left-width"},
		Right:  Class{Name: "border-right", Val: "border-right-width"},
	}

	buildVars(c.Theme, css)
	buildTheme(c.Theme, css)
	buildDirections(c.Margin, margin, css)
	buildDirections(c.Padding, padding, css)
	buildDirections(c.Border, border, css)
}

func buildVars(e Entries, css *CssFile) {
	css.AddLine(":root {\n")
	for _, row := range e {
		for key, value := range row {
			css.Vars[key] = fmt.Sprintf("var(--%s)", key)
			css.AddLine(fmt.Sprintf("    --%s: %s;\n", key, value))
		}
	}
	css.AddLine("}\n\n")
}

func buildTheme(e Entries, css *CssFile) {
	//background colors
	for _, row := range e {
		for key := range row {
			line := fmt.Sprintf(".bg-%s {\n    background-color: %s;\n}\n\n", key, css.Vars[key])
			css.AddLine(line)
		}

	}

	//text colors
	for _, row := range e {
		for key := range row {
			line := fmt.Sprintf(".text-%s {\n    color: %s;\n}\n\n", key, css.Vars[key])
			css.AddLine(line)
		}
	}

	//border colors
	for _, row := range e {
		for key := range row {
			line := fmt.Sprintf(".border-%s {\n    border-color: %s;\n}\n\n", key, css.Vars[key])
			css.AddLine(line)
		}
	}
}

func buildDirections(e Entries, d Directions, css *CssFile) {
	//base
	for _, row := range e {
		for key, value := range row {
			line := fmt.Sprintf(".%s-%s {\n    %s: %s;\n}\n\n", d.Base.Name, key, d.Base.Val, value)
			css.AddLine(line)
		}
	}

	//top
	for _, row := range e {
		for key, value := range row {
			line := fmt.Sprintf(".%s-%s {\n    %s: %s;\n}\n\n", d.Top.Name, key, d.Top.Val, value)
			css.AddLine(line)
		}
	}

	//bottom
	for _, row := range e {
		for key, value := range row {
			line := fmt.Sprintf(".%s-%s {\n    %s: %s;\n}\n\n", d.Bottom.Name, key, d.Bottom.Val, value)
			css.AddLine(line)
		}
	}

	//left
	for _, row := range e {
		for key, value := range row {
			line := fmt.Sprintf(".%s-%s {\n    %s: %s;\n}\n\n", d.Left.Name, key, d.Left.Val, value)
			css.AddLine(line)
		}
	}

	//right
	for _, row := range e {
		for key, value := range row {
			line := fmt.Sprintf(".%s-%s {\n    %s: %s;\n}\n\n", d.Right.Name, key, d.Right.Val, value)
			css.AddLine(line)
		}
	}
}
