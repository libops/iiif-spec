package iiifspec_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestOpenAPIRefsResolve(t *testing.T) {
	files := []string{
		filepath.Join("openapi", "image", "v3", "openapi.yaml"),
	}
	refRE := regexp.MustCompile(`\$ref:\s+([^#\s][^\s]*)`)
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			matches := refRE.FindAllSubmatch(data, -1)
			if len(matches) == 0 {
				t.Fatal("no external refs found")
			}
			for _, match := range matches {
				ref := string(match[1])
				ref = strings.Trim(ref, `"'`)
				if strings.HasPrefix(ref, "#") {
					continue
				}
				target := ref
				if i := len(target); i > 0 {
					if hash := regexp.MustCompile(`#.*$`).FindStringIndex(target); hash != nil {
						target = target[:hash[0]]
					}
				}
				if target == "" {
					continue
				}
				if _, err := os.Stat(filepath.Clean(filepath.Join(filepath.Dir(file), target))); err != nil {
					t.Fatalf("%s ref %s does not resolve: %v", file, ref, err)
				}
			}
		})
	}
}
