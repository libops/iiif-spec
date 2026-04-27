package iiifspec_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpstreamLockMatchesVendoredArtifacts(t *testing.T) {
	data, err := os.ReadFile("upstream.lock")
	if err != nil {
		t.Fatal(err)
	}
	records := parseLockRecords(string(data))
	if len(records) == 0 {
		t.Fatal("upstream.lock has no artifact records")
	}

	var sourceFiles []string
	if err := filepath.WalkDir("upstream", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".source") {
			sourceFiles = append(sourceFiles, path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	seen := make(map[string]bool, len(records))
	for _, rec := range records {
		dest := rec["dest"]
		if dest == "" {
			t.Fatalf("lock record missing dest: %#v", rec)
		}
		body, err := os.ReadFile(dest)
		if err != nil {
			t.Fatalf("read %s: %v", dest, err)
		}
		sum := sha256.Sum256(body)
		if got, want := hex.EncodeToString(sum[:]), rec["sha256"]; got != want {
			t.Fatalf("%s sha256 = %s, want %s", dest, got, want)
		}
		if _, err := os.Stat(dest + ".source"); err != nil {
			t.Fatalf("%s missing sidecar source manifest: %v", dest, err)
		}
		seen[filepath.ToSlash(dest)+".source"] = true
	}

	for _, sourceFile := range uniqueStrings(sourceFiles) {
		sourceFile = filepath.ToSlash(filepath.Clean(sourceFile))
		if !seen[sourceFile] {
			t.Fatalf("%s is not represented in upstream.lock", sourceFile)
		}
	}
}

func parseLockRecords(lock string) []map[string]string {
	var records []map[string]string
	var current map[string]string
	for _, line := range strings.Split(lock, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") {
			if current != nil {
				records = append(records, current)
			}
			current = make(map[string]string)
			line = strings.TrimPrefix(line, "- ")
		}
		if current == nil || !strings.Contains(line, ":") {
			continue
		}
		key, value, _ := strings.Cut(line, ":")
		current[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	if current != nil {
		records = append(records, current)
	}
	return records
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	var out []string
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
