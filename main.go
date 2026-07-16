package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Test this with "/index.html", "/etc/passwd", and "/../etc/passwd"
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("The -Cwd-:", cwd)

	isSafePath(cwd, "/index.html")
	isSafePath(cwd, "/etc/passwd")
	isSafePath(cwd, "/../etc/passwd")
	isSafePath(cwd, "../http-engine-go-secrets/flag.txt")
}

func isSafePath(cwd string, targetPath string) bool {
	filePath := targetPath
	log.Println("Target path:", filePath)

	fullPath := filepath.Join(cwd, filePath)
	log.Println("Full path:", fullPath)

	basePrefix := cwd + string(filepath.Separator)
	log.Println("Base prefix:", basePrefix)

	if !strings.HasPrefix(fullPath, basePrefix) && fullPath != cwd {
		log.Println("Error: File path is outside the current working directory.")
		return false
	}

	log.Println("Success! File path is safe to use.")
	log.Println()
	return true
}
