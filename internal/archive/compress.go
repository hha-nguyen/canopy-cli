package archive

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var defaultIgnorePatterns = []string{
	".git",
	".svn",
	".hg",
	"node_modules",
	".gradle",
	"build",
	"DerivedData",
	"Pods",
	".idea",
	".vscode",
	"*.xcworkspace",
	".DS_Store",
	"*.log",
	"__pycache__",
	"*.pyc",
	"*.class",
	"target",
	"dist",
	"*.tmp",
	"*.temp",
}

type CompressOptions struct {
	IgnorePatterns []string
	MaxSize        int64
	ShowProgress   bool
}

func DefaultCompressOptions() *CompressOptions {
	return &CompressOptions{
		IgnorePatterns: defaultIgnorePatterns,
		MaxSize:        500 * 1024 * 1024,
		ShowProgress:   true,
	}
}

func CompressDirectory(srcDir, destFile string, opts *CompressOptions) error {
	if opts == nil {
		opts = DefaultCompressOptions()
	}

	ignorePatterns := append([]string{}, opts.IgnorePatterns...)
	canopyIgnore := filepath.Join(srcDir, ".canopyignore")
	if userPatterns, err := loadIgnoreFile(canopyIgnore); err == nil {
		ignorePatterns = append(ignorePatterns, userPatterns...)
	}

	file, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("create archive file: %w", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		if shouldIgnore(relPath, info, ignorePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("create tar header: %w", err)
		}

		header.Name = relPath

		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return nil
			}
			if !filepath.IsAbs(link) {
				absLink := filepath.Join(filepath.Dir(path), link)
				if !strings.HasPrefix(absLink, srcDir) {
					return nil
				}
			}
			header.Linkname = link
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("write tar header: %w", err)
		}

		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("open file: %w", err)
			}
			defer f.Close()

			if _, err := io.Copy(tarWriter, f); err != nil {
				return fmt.Errorf("write file to tar: %w", err)
			}
		}

		return nil
	})
}

func shouldIgnore(path string, info os.FileInfo, patterns []string) bool {
	name := info.Name()

	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "*") {
			ext := strings.TrimPrefix(pattern, "*")
			if strings.HasSuffix(name, ext) {
				return true
			}
		} else if name == pattern {
			return true
		} else if strings.Contains(path, pattern) {
			return true
		}
	}

	return false
}

func loadIgnoreFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}

	return patterns, scanner.Err()
}
