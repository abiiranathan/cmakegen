package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

type Config struct {
	ProjectName string
	Description string
	Language    string
	Linklibs    string
	Installable bool
	Library     bool
	Configure   bool
	Generator   string
}

var config Config = Config{
	ProjectName: "",
	Language:    "C",
	Linklibs:    "",
	Installable: false,
	Library:     false,
	Configure:   false,
	Generator:   "Ninja",
}

func validate() {
	if config.ProjectName == "" {
		// use current directory name, removing all whitespaces
		// and converting to lower case
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		name := filepath.Base(cwd)
		name = strings.ToLower(strings.ReplaceAll(name, " ", "_"))
		config.ProjectName = name
	}

	if config.Language != "C" && config.Language != "CXX" {
		panic("lang must be either C or CXX")
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.StringVar(&config.ProjectName, "name", "", "Name of the project")
	flag.StringVar(&config.Description, "desc", "", "Description of the project")
	flag.StringVar(&config.Language, "lang", "C", "Language of the project")
	flag.StringVar(&config.Linklibs, "linklibs", "", "Linker libraries")
	flag.BoolVar(&config.Installable, "install", true, "Is the project installable")
	flag.BoolVar(&config.Library, "lib", false, "If the project is a library")
	flag.BoolVar(&config.Configure, "configure", false, "Run cmake after generating the files")
	flag.StringVar(&config.Generator, "generator", "Ninja", "CMake generator to use")

	flag.Parse()
	validate()

	// create directories
	check(os.MkdirAll("include", 0755))
	check(os.MkdirAll("src", 0755))
	check(os.MkdirAll("build", 0755))

	// render the template
	err := renderTemplate(&config)
	check(err)

	if config.Library && config.Installable {
		// generate the .pc file
		generatePCFile(&config)

		// generate the cmake config file
		generateCmakeConfig(&config)
	}

	if config.Configure {
		// run cmake
		check(runCmake(&config))
	}
}

func runCmake(config *Config) error {
	os.RemoveAll("build")
	check(os.MkdirAll("build", 0755))

	validGenerators := []string{
		"Ninja", "Unix Makefiles", "Ninja Multi-Config",
		"Watcom WMake", "CodeBlocks - Ninja", "CodeBlocks - Unix Makefiles",
		"Green Hills MULTI",
	}

	if !slices.Contains(validGenerators, config.Generator) {
		return fmt.Errorf("invalid generator %s. Must be one of %v", config.Generator, validGenerators)
	}

	args := []string{
		"-S", ".",
		"-B", "build",
		"-G", config.Generator,
	}

	cmd := exec.Command("cmake", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
