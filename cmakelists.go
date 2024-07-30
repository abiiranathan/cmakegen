package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

const templateString = `
cmake_minimum_required(VERSION 3.5.0)
project({{ .ProjectName }} VERSION 0.1.0 LANGUAGES {{ .Language }} DESCRIPTION "{{ .Description }}")

set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)
set(CMAKE_C_EXTENSIONS ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
list(APPEND CMAKE_MODULE_PATH "${CMAKE_CURRENT_SOURCE_DIR}/cmake")

# Find cmake modules here
# e.g. find_package(solidc REQUIRED)

include(GNUInstallDirs)
include(CTest)
enable_testing()

file(GLOB_RECURSE SOURCES "src/*.c")
file(GLOB_RECURSE HEADERS "include/*.h")

{{ if .Library -}}
add_library(${PROJECT_NAME} ${SOURCES})
{{ else }}
add_executable(${PROJECT_NAME} ${SOURCES})
{{ end -}}

# Add the compile options(CFLAGS)
target_compile_options(${PROJECT_NAME} PUBLIC -Wall -Wextra -Werror )
{{ if .Linklibs -}}
target_link_libraries(${PROJECT_NAME} PUBLIC {{ .Linklibs }})
{{ end }}

if(CMAKE_C_COMPILER_ID STREQUAL "GNU")
    # If in release mode, add some security flags
    if(CMAKE_BUILD_TYPE STREQUAL "Release")
        message(STATUS "Using GCC, adding security flags")
        target_compile_options(${PROJECT_NAME} PUBLIC 
            -fstack-protector-strong 
            -D_FORTIFY_SOURCE=2 
            -fPIE 
            -fPIC 
            -O2)
    endif()
endif()

# Include the include directory
target_include_directories(${PROJECT_NAME} PUBLIC
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
    $<INSTALL_INTERFACE:include/${PROJECT_NAME}>
)

# Set the public macros for the library, especially for the implementation
# of header-only libraries and overriding the default macros
# target_compile_definitions(${PROJECT_NAME} PUBLIC MY_MACRO=1 STB_IMAGE_IMPLEMENTATION)

{{ if and .Library .Installable -}}
# Install targets
install(FILES ${HEADERS} DESTINATION ${CMAKE_INSTALL_INCLUDEDIR}/${PROJECT_NAME})

install(TARGETS ${PROJECT_NAME} 
    EXPORT ${PROJECT_NAME}_export
    LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR} 
    PUBLIC_HEADER DESTINATION ${CMAKE_INSTALL_INCLUDEDIR}
)

# Install export targets
install(EXPORT ${PROJECT_NAME}_export
    FILE ${PROJECT_NAME}Targets.cmake
    NAMESPACE ${PROJECT_NAME}::
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/${PROJECT_NAME})

install(FILES ${CMAKE_CURRENT_BINARY_DIR}/${PROJECT_NAME}Config.cmake
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/${PROJECT_NAME})

# Generate the configuration file for the library
include(CMakePackageConfigHelpers)
write_basic_package_version_file(
    ${CMAKE_CURRENT_BINARY_DIR}/${PROJECT_NAME}ConfigVersion.cmake
    VERSION ${PROJECT_VERSION}
    COMPATIBILITY AnyNewerVersion
)

configure_package_config_file(
    ${CMAKE_CURRENT_SOURCE_DIR}/${PROJECT_NAME}Config.cmake.in
    ${CMAKE_CURRENT_BINARY_DIR}/${PROJECT_NAME}Config.cmake
    INSTALL_DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/${PROJECT_NAME}
)

install(FILES ${CMAKE_CURRENT_BINARY_DIR}/${PROJECT_NAME}ConfigVersion.cmake
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/${PROJECT_NAME})

# Generate the .pc file for pkg-config
configure_file(${CMAKE_CURRENT_SOURCE_DIR}/${PROJECT_NAME}.pc.in
    ${CMAKE_CURRENT_BINARY_DIR}/${PROJECT_NAME}.pc
    @ONLY)

# Install the .pc file
install(FILES ${CMAKE_CURRENT_BINARY_DIR}/${PROJECT_NAME}.pc
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/pkgconfig)
{{ end }}
`

func renderTemplate(cfg *Config) error {
	tmpl, err := template.New("CMakeLists").Parse(templateString)
	if err != nil {
		return err
	}

	f, err := os.Create("CMakeLists.txt")
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, cfg)
}

func generatePCFile(config *Config) error {
	if !config.Installable || !config.Library {
		return nil
	}

	linkLibs := strings.Split(config.Linklibs, " ")
	// append -l to each library
	for i, lib := range linkLibs {
		linkLibs[i] = "-l" + strings.TrimSpace(lib)
	}

	// pc file
	pcFile := `prefix=@CMAKE_INSTALL_PREFIX@
exec_prefix=${prefix}
libdir=${exec_prefix}/@CMAKE_INSTALL_LIBDIR@
includedir=${prefix}/@CMAKE_INSTALL_INCLUDEDIR@

Name: @PROJECT_NAME@
Description: @PROJECT_DESCRIPTION@
Version: @PROJECT_VERSION@
Cflags: -I${includedir}
Libs: -L${libdir} -l@PROJECT_NAME@ `

	if len(linkLibs) > 0 {
		pcFile += strings.Join(linkLibs, " ")
	}

	return os.WriteFile(config.ProjectName+".pc.in", []byte(pcFile), 0644)

}

func generateCmakeConfig(config *Config) error {
	configFile := fmt.Sprintf(`@PACKAGE_INIT@

include("${CMAKE_CURRENT_LIST_DIR}/%sTargets.cmake")
`, config.ProjectName)

	return os.WriteFile(config.ProjectName+"Config.cmake.in", []byte(configFile), 0644)

}
