package scanner

import (
	"os"
	"path/filepath"
	"strings"
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
	rootDir      string
	files        []string
	maxFiles     int
	maxFileSizeKB int
}

func NewLocalFileReader(rootDir string, maxFiles, maxFileSizeKB int) *LocalFileReader {
	return &LocalFileReader{
		rootDir:      rootDir,
		maxFiles:     maxFiles,
		maxFileSizeKB: maxFileSizeKB,
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
