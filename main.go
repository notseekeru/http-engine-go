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

	pathMap := map[string]bool{
		"/index.html":                        isSafePath(cwd, "/index.html"),                        //true
		"/etc/passwd":                        isSafePath(cwd, "/etc/passwd"),                        //true
		"/../etc/passwd":                     isSafePath(cwd, "/../etc/passwd"),                     //false
		"../http-engine-go-secrets/flag.txt": isSafePath(cwd, "../http-engine-go-secrets/flag.txt"), //false
	}

	for path, isSafe := range pathMap {
		log.Printf("Path: %s, Is Safe: %t", path, isSafe)
	}
}

func isSafePath(cwd string, targetPath string) bool {
	filePath := targetPath
	fullPath := filepath.Join(cwd, filePath)
	basePrefix := cwd + string(filepath.Separator)

	if !strings.HasPrefix(fullPath, basePrefix) && fullPath != cwd {
		return false
	}

	return true
}
