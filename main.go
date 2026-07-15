package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Test this with "/index.html", "/etc/passwd", and "/../etc/passwd"
	maliciousPath := "/malicious_shortcut/passwd"
	log.Println("Malicious path:", maliciousPath)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("The -Cwd-:", cwd)

	fullPath := filepath.Join(cwd, maliciousPath)
	log.Println("Full path:", fullPath)


	basePrefix := cwd + string(filepath.Separator)
	log.Println("Base prefix:", basePrefix)

	if !strings.HasPrefix(fullPath, basePrefix) && fullPath != cwd {
		log.Fatal("Access to files outside the allowed directory is not permitted")
	}

	log.Println("Success! File path is safe to use.")
}
