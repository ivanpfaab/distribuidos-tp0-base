package common

import (
	"encoding/csv"
	"io"
	"os"
)


type CSVChunkReader struct {
  reader *csv.Reader
  isEOF  bool
}

func NewCSVChunkReader(file *os.File) *CSVChunkReader {
  return &CSVChunkReader{
      reader: csv.NewReader(file),
      isEOF:  false,
  }
}

func (c *CSVChunkReader) ReadChunk(x int) ([][]string, error) {
  if c.isEOF {
      return nil, io.EOF
  }
  
  var lines [][]string
  
  for i := 0; i < x; i++ {
      record, err := c.reader.Read()
      if err == io.EOF {
          c.isEOF = true
          break
      }
      if err != nil {
          return nil, err
      }
      lines = append(lines, record)
  }
  
  return lines, nil
}

func (c *CSVChunkReader) HasMore() bool {
  return !c.isEOF
}