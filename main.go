package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Testing grounds for my implementation of solutions

func main() {
	maliciousPath := "/index.html"
	cleanPath := filepath.Clean(maliciousPath)
	log.Println(cleanPath)

	cwd, err := os.Getwd()
	log.Println("Current working directory:", cwd)
	if err != nil {
		log.Fatal(err)
	}

	fullPath := filepath.Join(cwd, cleanPath)
	log.Println("Full path:", fullPath)

	if !strings.HasPrefix(fullPath, cwd) {
		log.Fatal("Access to files outside the allowed directory is not permitted")
	}

	// Dynamically build your base folder path
}
