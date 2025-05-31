package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// Untar takes a destination path and a reader; a tar reader loops over the tar file
// creating the file structure at 'dst' along the way, and writing any files.
func Untar(destDir string, reader io.Reader) error {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer func() {
		_ = gzReader.Close()
	}()

	tarReader := tar.NewReader(gzReader)

	for {
		var header *tar.Header
		header, err = tarReader.Next()

		switch {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(target, 0755); err != nil && !os.IsExist(err) {
				return err
			}
		case tar.TypeReg:
			if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil && !os.IsExist(err) {
				return err
			}

			var file *os.File
			file, err = os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err = io.Copy(file, tarReader); err != nil {
				return err
			}

			_ = file.Close()
		}
	}
}

// CountLines takes a reader and counts the number of lines in the reader.
func CountLines(reader io.Reader) (uint, error) {
	var count uint
	const lineBreak = '\n'

	buf := make([]byte, bufio.MaxScanTokenSize)

	for {
		bufferSize, err := reader.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, err
		}

		var buffPosition int
		for {
			i := bytes.IndexByte(buf[buffPosition:], lineBreak)
			if i == -1 || bufferSize == buffPosition {
				break
			}
			buffPosition += i + 1
			count++
		}
		if errors.Is(err, io.EOF) {
			break
		}
	}

	return count, nil
}
