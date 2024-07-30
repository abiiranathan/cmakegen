# cmakegen

A simple command line tool to generate a cmake project structure.

It automatically generates the configuration for pkg-config, install targets, and a standard CMakeLists.txt file.

## Installation

```bash
go install github.com/abiiranathan/cmakegen
```

## Usage

```bash
cmakegen -h
Usage of cmakegen:
  -configure
        Run cmake after generating the files
  -desc string
        Description of the project
  -generator string
        CMake generator to use (default "Ninja")
  -install
        Is the project installable (default true)
  -lang string
        Language of the project (default "C")
  -lib
        Is the project a library
  -linklibs string
        Linker libraries
  -name string
        Name of the project
```

If the project name is not provided, the current directory name is used as the project name, replacing spaces with underscores.

## Example

```bash
cmakegen -name my_project -desc "My awesome project" -lang C -linklibs "pthread m" -configure
```

This will generate the following files in the working directory:

```txt
build/
include/
src/
    main.c
CMakeLists.txt
my_project.pc.in
my_projectConfig.cmake.in
```

If the `-configure` flag is provided, it will also run `cmake` in the `build` directory and generate the build files and targets.

## License

MIT
