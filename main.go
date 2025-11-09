package main

import (
	"encoding/json"
    "flag"
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
	Vars   map[string]string
	Shades []string
	Lines  []string
}

func (c *CssFile) AddLine(line string) {
	c.Lines = append(c.Lines, line)
}

type Config struct {
	Theme   Entries  `json:"theme"`
	Shades  []string `json:"shades"`
	Spacing Entries  `json:"spacing"`
	Margin  Entries  `json:"padding"`
	Padding Entries  `json:"margin"`
	Border  Entries  `json:"border"`
}

func main() {
    //file flags
    configFlag := flag.String("config", "", "name of JSON config file")
    outputFlag := flag.String("output", "", "name of CSS output file")

    //optional flag to print banner
    banner := flag.Bool("banner", false, "Print banner at the top of CSS file")
    flag.Parse()

    configName := *configFlag
    outputName := *outputFlag

	file, err := os.Open(configName)
	if err != nil {
		log.Fatal(fmt.Sprintf("couldn't open %s", configName), err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("couldn't decode config, check JSON format")
	}

	css := CssFile{Vars: make(map[string]string), Shades: make([]string, 0)}
	css.buildCss(config, *banner)

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

func (css *CssFile) buildCss(c Config, printBanner bool) {
    if printBanner {
        css.AddLine("/* Generated with Sculptor: https://github.com/section14/sculptor */\n\n")
    }

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

	css.buildVars(c.Theme, c.Shades, c.Spacing)
	css.buildTheme(c.Theme, c.Shades)
	css.buildDirections(c.Margin, margin)
	css.buildDirections(c.Padding, padding)
	css.buildDirections(c.Border, border)
}

//oklch(from var(--color-base-100) calc(l + 0.08) c h)

func shadeBuilder(color, shade string) string {
	switch shade {
	case "100":
		return fmt.Sprintf("oklch(from var(--%s) calc(l + 0.35) c h);", color)
	case "200":
		return fmt.Sprintf("oklch(from var(--%s) calc(l + 0.25) c h);", color)
	case "300":
		return fmt.Sprintf("oklch(from var(--%s) calc(l + 0.15) c h);", color)
	case "400":
		return fmt.Sprintf("oklch(from var(--%s) calc(l + 0.05) c h);", color)
	case "500":
		return fmt.Sprintf("var(--%s);", color)
	case "600":
		return fmt.Sprintf("oklch(from var(--%s) calc(l - 0.05) c h);", color)
	case "700":
		return fmt.Sprintf("oklch(from var(--%s) calc(l - 0.15) c h);", color)
	case "800":
		return fmt.Sprintf("oklch(from var(--%s) calc(l - 0.25) c h);", color)
	case "900":
		return fmt.Sprintf("oklch(from var(--%s) calc(l - 0.35) c h);", color)
	default:
		return fmt.Sprintf("var(--%s)", color)
	}
}

func (css *CssFile) buildVars(color Entries, shades []string, spacing Entries) {
	var colorNames []string

	css.AddLine(":root {\n")
	//colors
	for _, row := range color {
		for key, value := range row {
			colorNames = append(colorNames, key)
			css.Vars[key] = fmt.Sprintf("var(--%s)", key)
			css.AddLine(fmt.Sprintf("    --%s: %s;\n", key, value))
		}
	}

	css.AddLine("\n")

	//shades
	for _, c := range colorNames {
		for _, s := range shades {
			shadeName := fmt.Sprintf("--%s-%s", c, s)
			newHsl := shadeBuilder(c, s)
			newShade := fmt.Sprintf("    %s: %s\n", shadeName, newHsl)
            css.Shades = append(css.Shades, shadeName)
			css.AddLine(newShade)
		}
		css.AddLine("\n")
	}

	//spacing
	for _, row := range spacing {
		for key, value := range row {
			css.Vars[key] = fmt.Sprintf("var(--spacing-%s)", key)
			css.AddLine(fmt.Sprintf("    --spacing-%s: %s;\n", key, value))
		}
	}
	css.AddLine("}\n\n")
}

func (css *CssFile) buildTheme(e Entries, shades []string) {
	//background colors
    var bgOffset = 0
	for _, row := range e {
		for key := range row {
			line := fmt.Sprintf(".bg-%s {\n    background-color: %s;\n}\n\n", key, css.Vars[key])
			css.AddLine(line)

			//add color shade classes
			for i, s := range shades {
				line := fmt.Sprintf(".bg-%s-%s {\n    background-color: var(%s);\n}\n\n", 
                    key, s, css.Shades[i + (bgOffset * len(shades))])
				css.AddLine(line)
			}

            bgOffset++
		}
	}

	//text colors
    var textOffset = 0
	for _, row := range e {
		for key := range row {
			line := fmt.Sprintf(".text-%s {\n    color: %s;\n}\n\n", key, css.Vars[key])
			css.AddLine(line)

			//add color shade classes
			for i, s := range shades {
				line := fmt.Sprintf(".text-%s-%s {\n    color: var(%s);\n}\n\n", 
                    key, s, css.Shades[i + (textOffset * len(shades))])
				css.AddLine(line)
			}

            textOffset++
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

func (css *CssFile) buildDirections(e Entries, d Directions) {
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
