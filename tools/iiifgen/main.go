// Command iiifgen vendors canonical IIIF machine-readable artifacts from
// upstream community repositories into this repo.
//
// Vendored files are pinned by commit SHA and written alongside a `.source`
// manifest recording the upstream repo, SHA, original path, and content hash.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type fileSpec struct {
	Src string
	Dst string
}

type dirSpec struct {
	SrcPrefix string
	DstDir    string
	Suffixes  []string
}

type source struct {
	Repo  string
	SHA   string
	Files []fileSpec
	Dirs  []dirSpec
}

var sources = []source{
	{
		Repo: "IIIF/api",
		SHA:  "c57f35b39511b3a458d317c0d0f2f4dd277bd590",
		Files: []fileSpec{
			{
				Src: "source/image/3/context.json",
				Dst: "upstream/iiif-api/source/image/3/context.json",
			},
			{
				Src: "source/image/3/info_frame.json",
				Dst: "upstream/iiif-api/source/image/3/info_frame.json",
			},
			{
				Src: "source/image/3/level0.json",
				Dst: "upstream/iiif-api/source/image/3/level0.json",
			},
			{
				Src: "source/image/3/level1.json",
				Dst: "upstream/iiif-api/source/image/3/level1.json",
			},
			{
				Src: "source/image/3/level2.json",
				Dst: "upstream/iiif-api/source/image/3/level2.json",
			},
			{
				Src: "source/presentation/3/context.json",
				Dst: "upstream/iiif-api/source/presentation/3/context.json",
			},
			{
				Src: "source/auth/2/context.json",
				Dst: "upstream/iiif-api/source/auth/2/context.json",
			},
			{
				Src: "source/search/2/context.json",
				Dst: "upstream/iiif-api/source/search/2/context.json",
			},
			{
				Src: "source/extension/text-granularity/context.json",
				Dst: "upstream/iiif-api/source/extension/text-granularity/context.json",
			},
			{
				Src: "source/extension/navplace/context.json",
				Dst: "upstream/iiif-api/source/extension/navplace/context.json",
			},
		},
	},
	{
		Repo: "IIIF/presentation-validator",
		SHA:  "5fccfbc01ac6627ecdb81f845a595b5ea295f86d",
		Files: []fileSpec{
			{
				Src: "schema/iiif_3_0.json",
				Dst: "upstream/presentation-validator/schema/iiif_3_0.json",
			},
		},
		Dirs: []dirSpec{
			{
				SrcPrefix: "schema/v4/",
				DstDir:    "upstream/presentation-validator/schema/v4",
				Suffixes:  []string{".json"},
			},
			{
				SrcPrefix: "fixtures/3/",
				DstDir:    "upstream/presentation-validator/fixtures/3",
				Suffixes:  []string{".json"},
			},
		},
	},
	{
		Repo: "IIIF/image-validator",
		SHA:  "1740893f1fb22960142071a9f3d1c99122a190c7",
		Dirs: []dirSpec{
			{
				SrcPrefix: "tests/json/",
				DstDir:    "upstream/image-validator/tests/json",
				Suffixes:  []string{".json"},
			},
		},
	},
}

type treeResponse struct {
	Tree []treeEntry `json:"tree"`
}

type treeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type lockRecord struct {
	Repo   string
	SHA    string
	Source string
	Dest   string
	SHA256 string
}

func main() {
	dryRun := flag.Bool("dry-run", false, "print actions without writing files")
	timeout := flag.Duration("timeout", 30*time.Second, "per-request HTTP timeout")
	flag.Parse()

	client := &http.Client{Timeout: *timeout}

	moduleRoot, err := findModuleRoot()
	if err != nil {
		die("locate module root: %v", err)
	}

	var lock []lockRecord
	for _, s := range sources {
		if s.SHA == "" {
			die("missing SHA for %s", s.Repo)
		}
		for _, f := range s.Files {
			rec, err := vendorFile(client, moduleRoot, s.Repo, s.SHA, f.Src, f.Dst, *dryRun)
			if err != nil {
				die("%s@%s %s: %v", s.Repo, shortSHA(s.SHA), f.Src, err)
			}
			lock = append(lock, rec)
		}
		for _, d := range s.Dirs {
			records, err := vendorDir(client, moduleRoot, s.Repo, s.SHA, d, *dryRun)
			if err != nil {
				die("%s@%s %s: %v", s.Repo, shortSHA(s.SHA), d.SrcPrefix, err)
			}
			lock = append(lock, records...)
		}
	}
	if !*dryRun {
		if err := writeLockFile(moduleRoot, lock); err != nil {
			die("write upstream.lock: %v", err)
		}
	}
}

func vendorDir(client *http.Client, moduleRoot, repo, sha string, dir dirSpec, dryRun bool) ([]lockRecord, error) {
	paths, err := listTreePaths(client, repo, sha)
	if err != nil {
		return nil, fmt.Errorf("list tree: %w", err)
	}
	var records []lockRecord
	for _, src := range paths {
		if !strings.HasPrefix(src, dir.SrcPrefix) {
			continue
		}
		if !hasAllowedSuffix(src, dir.Suffixes) {
			continue
		}
		rel := strings.TrimPrefix(src, dir.SrcPrefix)
		dst := filepath.ToSlash(filepath.Join(dir.DstDir, rel))
		rec, err := vendorFile(client, moduleRoot, repo, sha, src, dst, dryRun)
		if err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

func hasAllowedSuffix(path string, suffixes []string) bool {
	if len(suffixes) == 0 {
		return true
	}
	for _, suffix := range suffixes {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}
	return false
}

func listTreePaths(client *http.Client, repo, sha string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/git/trees/%s?recursive=1", repo, sha)
	body, err := fetch(client, url)
	if err != nil {
		return nil, err
	}
	var tr treeResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, err
	}
	var out []string
	for _, entry := range tr.Tree {
		if entry.Type == "blob" {
			out = append(out, entry.Path)
		}
	}
	slices.Sort(out)
	return out, nil
}

func vendorFile(client *http.Client, moduleRoot, repo, sha, src, dst string, dryRun bool) (lockRecord, error) {
	absDst := filepath.Join(moduleRoot, filepath.FromSlash(dst))
	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", repo, sha, src)
	body, err := fetch(client, rawURL)
	if err != nil {
		return lockRecord{}, fmt.Errorf("fetch %s: %w", rawURL, err)
	}
	sum := sha256.Sum256(body)
	rec := lockRecord{
		Repo:   repo,
		SHA:    sha,
		Source: src,
		Dest:   dst,
		SHA256: hex.EncodeToString(sum[:]),
	}
	if dryRun {
		fmt.Printf("would write %s (%d bytes, sha256=%s)\n", dst, len(body), rec.SHA256)
		return rec, nil
	}
	if err := writeWithManifest(absDst, body, repo, sha, src, sum); err != nil {
		return lockRecord{}, fmt.Errorf("write %s: %w", dst, err)
	}
	fmt.Printf("wrote %s (from %s@%s)\n", dst, repo, shortSHA(sha))
	return rec, nil
}

func fetch(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func writeWithManifest(dst string, body []byte, repo, sha, src string, sum [32]byte) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(dst, body, 0o644); err != nil {
		return err
	}
	manifest := fmt.Sprintf(
		"source: https://github.com/%s\nsha: %s\npath: %s\nsha256: %s\n",
		repo,
		sha,
		src,
		hex.EncodeToString(sum[:]),
	)
	return os.WriteFile(dst+".source", []byte(manifest), 0o644)
}

func writeLockFile(moduleRoot string, records []lockRecord) error {
	slices.SortFunc(records, func(a, b lockRecord) int {
		return strings.Compare(a.Dest, b.Dest)
	})

	var b strings.Builder
	b.WriteString("# Generated by go run ./tools/iiifgen; do not edit by hand.\n")
	b.WriteString("version: 1\n")
	b.WriteString("artifacts:\n")
	for _, rec := range records {
		fmt.Fprintf(&b, "  - dest: %s\n", rec.Dest)
		fmt.Fprintf(&b, "    source: https://github.com/%s\n", rec.Repo)
		fmt.Fprintf(&b, "    sha: %s\n", rec.SHA)
		fmt.Fprintf(&b, "    path: %s\n", rec.Source)
		fmt.Fprintf(&b, "    sha256: %s\n", rec.SHA256)
	}
	return os.WriteFile(filepath.Join(moduleRoot, "upstream.lock"), []byte(b.String()), 0o644)
}

func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod found above %q", dir)
		}
		dir = parent
	}
}

func shortSHA(sha string) string {
	if len(sha) < 8 {
		return sha
	}
	return sha[:8]
}

func die(format string, args ...any) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
