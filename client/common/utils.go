package common

import (
  "os"
  "path/filepath"
)

// OpenFileForRead Opens a file for reading and returns the file handle
func OpenFileForRead(path string) (*os.File, error) {
  abs, err := filepath.Abs(path)
  if err != nil {
    return nil, err
  }
  f, err := os.Open(abs) // read-only
  if err != nil {
    return nil, err
  }
  return f, nil
}
