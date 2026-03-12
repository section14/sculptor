package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Entries []map[string]string

type OkLch struct {
	L string `yaml:"l"`
	C string `yaml:"c"`
	H string `yaml:"h"`
}

type ThemeEntry struct {
	Color          map[string]string `yaml:"color"`
	GenerateShades bool              `yaml:"generateShades"`
	UserShades     map[string]OkLch  `yaml:"userShades"`
}

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
	Vars       map[string]string
	Shades     []string
	UserShades []string
	Lines      []string
}

func (c *CssFile) AddLine(line string) {
	c.Lines = append(c.Lines, line)
}

type Config struct {
	Theme   []ThemeEntry `yaml:"theme"`
	Shades  []string     `yaml:"shades"`
	Spacing Entries      `yaml:"spacing"`
	Margin  Entries      `yaml:"padding"`
	Padding Entries      `yaml:"margin"`
	Border  Entries      `yaml:"border"`
}

func main() {
	//file flags
	configFlag := flag.String("config", "", "name of YAML config file")
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
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)

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
		Left:   Class{Name: "ps", Val: "padding-left"},
		Right:  Class{Name: "pe", Val: "padding-right"},
	}

	margin := Directions{
		Base:   Class{Name: "m", Val: "margin"},
		Top:    Class{Name: "mt", Val: "margin-top"},
		Bottom: Class{Name: "mb", Val: "margin-bottom"},
		Left:   Class{Name: "ms", Val: "margin-left"},
		Right:  Class{Name: "me", Val: "margin-right"},
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

// predetermined shade generation
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

// user based shade generation
func buildShadeOffset(key string, ok OkLch) string {
	l := "l"
	c := "c"
	h := "h"

	if ok.L != "" {
		l = fmt.Sprintf("calc(l %s)", ok.L)
	}

	if ok.C != "" {
		c = fmt.Sprintf("calc(c %s)", ok.C)
	}

	if ok.H != "" {
		h = fmt.Sprintf("calc(h %s)", ok.H)
	}

	return fmt.Sprintf("oklch(from var(--%s) %s %s %s);", key, l, c, h)
}

func buildUserShade(color string, shades map[string]OkLch) ([]string, []string) {
	userShades := make([]string, 0)
	shadeNames := make([]string, 0)

	for k, v := range shades {
		c := buildShadeOffset(color, v)
		shade := fmt.Sprintf("    --%s-%s: %s\n", color, k, c)
		shadeNames = append(shadeNames, fmt.Sprintf("%s-%s", color, k))
		userShades = append(userShades, shade)
	}

	return shadeNames, userShades
}

func (css *CssFile) buildVars(th []ThemeEntry, shades []string, spacing Entries) {
	var shadeNames []string
	var userShades []string

	css.AddLine(":root {\n")

	//add root variables + shades to build later
	for _, entry := range th {
		for key, value := range entry.Color {
			if len(entry.UserShades) > 0 {
				names, shades := buildUserShade(key, entry.UserShades)
				userShades = append(userShades, shades...)
				css.UserShades = append(css.UserShades, names...)
			}

			if entry.GenerateShades && len(entry.UserShades) == 0 {
				shadeNames = append(shadeNames, key)
			}
			css.Vars[key] = fmt.Sprintf("var(--%s)", key)
			css.AddLine(fmt.Sprintf("    --%s: %s;\n", key, value))
		}
	}

	css.AddLine("\n")

	// build user provided shades
	for _, s := range userShades {
		css.AddLine(s)
	}

	css.AddLine("\n")

	// build shade variables
	for _, c := range shadeNames {
		for _, s := range shades {
			shadeName := fmt.Sprintf("--%s-%s", c, s)
			newHsl := shadeBuilder(c, s)
			newShade := fmt.Sprintf("    %s: %s\n", shadeName, newHsl)
			css.Shades = append(css.Shades, shadeName)
			css.AddLine(newShade)
		}
		css.AddLine("\n")
	}

	// spacing
	for _, row := range spacing {
		for key, value := range row {
			css.Vars[key] = fmt.Sprintf("var(--spacing-%s)", key)
			css.AddLine(fmt.Sprintf("    --spacing-%s: %s;\n", key, value))
		}
	}
	css.AddLine("}\n\n")
}

func (css *CssFile) buildTheme(th []ThemeEntry, shades []string) {
	//background colors
	var bgOffset = 0
	for _, entry := range th {
		for key := range entry.Color {
			line := fmt.Sprintf(".bg-%s {\n    background-color: %s;\n}\n\n", key, css.Vars[key])
			css.AddLine(line)

			// add user defined shades
			if len(entry.UserShades) > 0 {
				for i, s := range css.UserShades {
					line := fmt.Sprintf(".bg-%s {\n    background-color: var(--%s);\n}\n\n",
						s, css.UserShades[i])
					css.AddLine(line)
				}
			}

			// add color shade classes
			if entry.GenerateShades {
				for i, s := range shades {
					line := fmt.Sprintf(".bg-%s-%s {\n    background-color: var(%s);\n}\n\n",
						key, s, css.Shades[i+(bgOffset*len(shades))])
					css.AddLine(line)
				}
				bgOffset++
			}

		}
	}

	//text colors
	var textOffset = 0
	for _, entry := range th {
		for key := range entry.Color {
			line := fmt.Sprintf(".text-%s {\n    color: %s;\n}\n\n", key, css.Vars[key])
			css.AddLine(line)

			// add user defined shades
			if len(entry.UserShades) > 0 {
				for i, s := range css.UserShades {
					line := fmt.Sprintf(".text-%s {\n    color: var(--%s);\n}\n\n", s, css.UserShades[i])
					css.AddLine(line)
				}
			}

			// add color shade classes
			if entry.GenerateShades {
				for i, s := range shades {
					line := fmt.Sprintf(".text-%s-%s {\n    color: var(%s);\n}\n\n",
						key, s, css.Shades[i+(textOffset*len(shades))])
					css.AddLine(line)
				}
				textOffset++
			}

		}
	}

	//border colors
	for _, entry := range th {
		for key := range entry.Color {
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
