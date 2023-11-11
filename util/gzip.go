package util

import (
    "bytes"
    "compress/gzip"
    "fmt"
    "io"
    "os"
    "path/filepath"
)

func GzipFileToBytes(path string) ([]byte, error) {
    var buffer bytes.Buffer;

    writer := gzip.NewWriter(&buffer);
    writer.Name = filepath.Base(path);

    file, err := os.Open(path);
    if (err != nil) {
        return nil, fmt.Errorf("Could not open source file for gzip '%s': '%w'.", path, err);
    }
    defer file.Close();

    _, err = io.Copy(writer, file);
    if (err != nil) {
        return nil, fmt.Errorf("Could not copy file into gzip '%s': '%w'.", path, err);
    }

    err = writer.Close();
    if (err != nil) {
        return nil, fmt.Errorf("Failed to close gzip writer for '%s': '%w'.", path, err);
    }

    return buffer.Bytes(), nil;
}

func GzipBytesToFile(data []byte, path string) error {
    reader, err := gzip.NewReader(bytes.NewBuffer(bytes.Clone(data)));
    if (err != nil) {
        return fmt.Errorf("Failed to create gzip read for data to go in '%s': '%w'.", path, err);
    }

    clearData, err := io.ReadAll(reader);
    if (err != nil) {
        return fmt.Errorf("Failed to read gzip contents to go in '%s': '%w'.", path, err);
    }

    return WriteBinaryFile(clearData, path);
}
