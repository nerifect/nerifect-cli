package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

// Binary file extensions to skip when scanning.
var binaryExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".pdf": true, ".zip": true, ".tar": true, ".gz": true, ".tgz": true,
	".jar": true, ".exe": true, ".dmg": true, ".woff": true, ".woff2": true,
	".ttf": true, ".ico": true, ".svg": true, ".mp3": true, ".mp4": true,
	".mov": true, ".avi": true, ".so": true, ".dylib": true, ".dll": true,
	".a": true, ".o": true, ".pyc": true, ".class": true, ".wasm": true,
	".bmp": true, ".eot": true, ".otf": true, ".db": true, ".sqlite": true,
}

// Directories to always skip.
var skipDirs = map[string]bool{
	"node_modules": true, ".git": true, "__pycache__": true, ".venv": true,
	"venv": true, "vendor": true, "dist": true, "build": true, ".next": true,
	".nuxt": true, "target": true, ".idea": true, ".vscode": true, "coverage": true,
	".cache": true, ".tox": true, ".mypy_cache": true, ".pytest_cache": true,
	"env": true, ".env": true, ".terraform": true,
}

// LocalFileReader implements ai.FileReader for local directories.
type LocalFileReader struct {
	rootDir        string
	files          []string
	maxFiles       int
	maxFileSizeKB  int
	ignorePatterns []glob.Glob
	allowList      map[string]bool // if non-nil, only include these relative paths
}

func NewLocalFileReader(rootDir string, maxFiles, maxFileSizeKB int) *LocalFileReader {
	return &LocalFileReader{
		rootDir:        rootDir,
		maxFiles:       maxFiles,
		maxFileSizeKB:  maxFileSizeKB,
		ignorePatterns: loadIgnorePatterns(rootDir),
	}
}

// NewLocalFileReaderWithAllowList creates a reader that only includes files in the allow list.
func NewLocalFileReaderWithAllowList(rootDir string, maxFiles, maxFileSizeKB int, allowList map[string]bool) *LocalFileReader {
	return &LocalFileReader{
		rootDir:        rootDir,
		maxFiles:       maxFiles,
		maxFileSizeKB:  maxFileSizeKB,
		ignorePatterns: loadIgnorePatterns(rootDir),
		allowList:      allowList,
	}
}

func (r *LocalFileReader) ListFiles() ([]string, error) {
	if r.files != nil {
		return r.files, nil
	}

	var files []string
	err := filepath.Walk(r.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}

		// Skip directories
		if info.IsDir() {
			name := filepath.Base(path)
			if skipDirs[name] {
				return filepath.SkipDir
			}
			// Check ignore patterns against directory paths
			if len(r.ignorePatterns) > 0 && path != r.rootDir {
				relDir, _ := filepath.Rel(r.rootDir, path)
				if relDir != "." {
					for _, pat := range r.ignorePatterns {
						if pat.Match(relDir) || pat.Match(relDir+"/") || pat.Match(name) {
							return filepath.SkipDir
						}
					}
				}
			}
			return nil
		}

		// Skip binary files
		ext := strings.ToLower(filepath.Ext(path))
		if binaryExts[ext] {
			return nil
		}

		// Skip files that are too large
		if r.maxFileSizeKB > 0 && info.Size() > int64(r.maxFileSizeKB)*1024 {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(r.rootDir, path)
		if err != nil {
			relPath = path
		}

		// Check allow list (diff-based scanning)
		if r.allowList != nil && !r.allowList[relPath] {
			return nil
		}

		// Check ignore patterns
		for _, pat := range r.ignorePatterns {
			if pat.Match(relPath) || pat.Match(filepath.Base(relPath)) {
				return nil
			}
		}

		files = append(files, relPath)

		// Cap total files
		if r.maxFiles > 0 && len(files) >= r.maxFiles {
			return filepath.SkipAll
		}

		return nil
	})

	r.files = files
	return files, err
}

func (r *LocalFileReader) ReadFile(path string) (string, error) {
	fullPath := filepath.Join(r.rootDir, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadFilesContents reads contents of all text files up to limits.
func (r *LocalFileReader) ReadFilesContents() (map[string]string, error) {
	files, err := r.ListFiles()
	if err != nil {
		return nil, err
	}

	contents := make(map[string]string)
	for _, path := range files {
		content, err := r.ReadFile(path)
		if err != nil {
			continue
		}
		contents[path] = content
	}
	return contents, nil
}

// IsTextFile checks if a file extension indicates a text file.
func IsTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return !binaryExts[ext]
}

// loadIgnorePatterns reads a .nerifectignore file and compiles glob patterns.
func loadIgnorePatterns(rootDir string) []glob.Glob {
	f, err := os.Open(filepath.Join(rootDir, ".nerifectignore"))
	if err != nil {
		return nil
	}
	defer f.Close()

	var patterns []glob.Glob
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip trailing slash for directory patterns — match both with and without
		g, err := glob.Compile(strings.TrimRight(line, "/"))
		if err != nil {
			continue
		}
		patterns = append(patterns, g)
	}
	return patterns
}
