package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	MaxArchiveSize = 500 * 1024 * 1024
)

type ArchiveType string

const (
	ArchiveTypeZip   ArchiveType = "zip"
	ArchiveTypeTarGz ArchiveType = "tar.gz"
)

func ValidateArchive(path string) (ArchiveType, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("stat file: %w", err)
	}

	if info.Size() > MaxArchiveSize {
		return "", fmt.Errorf("file too large: %d bytes (max %d bytes)", info.Size(), MaxArchiveSize)
	}

	ext := strings.ToLower(filepath.Ext(path))
	name := strings.ToLower(filepath.Base(path))

	if ext == ".zip" {
		if err := validateZip(path); err != nil {
			return "", fmt.Errorf("invalid zip file: %w", err)
		}
		return ArchiveTypeZip, nil
	}

	if ext == ".gz" || strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz") {
		if err := validateTarGz(path); err != nil {
			return "", fmt.Errorf("invalid tar.gz file: %w", err)
		}
		return ArchiveTypeTarGz, nil
	}

	return "", fmt.Errorf("unsupported file type: %s (supported: .zip, .tar.gz, .tgz)", ext)
}

func validateZip(path string) error {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer reader.Close()

	if len(reader.File) == 0 {
		return fmt.Errorf("archive is empty")
	}

	return nil
}

func validateTarGz(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("invalid gzip format: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	_, err = tarReader.Next()
	if err == io.EOF {
		return fmt.Errorf("archive is empty")
	}
	if err != nil {
		return fmt.Errorf("invalid tar format: %w", err)
	}

	return nil
}

func IsArchive(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	name := strings.ToLower(filepath.Base(path))

	return ext == ".zip" ||
		ext == ".gz" ||
		strings.HasSuffix(name, ".tar.gz") ||
		strings.HasSuffix(name, ".tgz")
}
