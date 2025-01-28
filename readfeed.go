package main

import (
	"archive/zip"
	"bufio"
	"encoding/csv"
	"io"
	"strings"

	"github.com/jszwec/csvutil"
)

type trimReader struct {
	csvReader *csv.Reader
}

func (tr trimReader) Read() ([]string, error) {
	row, err := tr.csvReader.Read()
	if err != nil {
		return nil, err
	}
	//trim spaces on feeds (should fix Renfe feed shenanigans)
	for i, v := range row {
		row[i] = strings.TrimSpace(v)
	}
	return row, nil
}

type objectReader[T any] struct {
	decoder    *csvutil.Decoder
	header     []string
	trimReader trimReader
	csvWriter  *csv.Writer //may be nil if no output feed given
}

func (or *objectReader[T]) read(val *T) error {
	err := or.decoder.Decode(&val)
	if err != nil {
		return err
	}
	return nil
}

func newObjectReader[T any](inputFeed *zip.ReadCloser, feedFilename string, writer io.Writer) (objectReader[T], error) {
	file, err := inputFeed.Open(feedFilename)
	if err != nil {
		return objectReader[T]{}, err
	}
	rd := bufio.NewReader(file)
	//check for a BOM
	r, _, err := rd.ReadRune()
	if err != nil {
		return objectReader[T]{}, err
	}
	if r != '\uFEFF' {
		if err := rd.UnreadRune(); err != nil {
			//should never ever happen according to docs
			panic(err)
		}
	}
	csvReader := csv.NewReader(rd)
	trimReader := trimReader{csvReader}
	header, err := trimReader.Read()
	if err != nil {
		return objectReader[T]{}, err
	}
	decoder, err := csvutil.NewDecoder(trimReader, header...)
	if err != nil {
		return objectReader[T]{}, err
	}

	var csvWriter *csv.Writer
	if writer != nil {
		csvWriter = csv.NewWriter(writer)
		csvWriter.Write(header) //header needs to be written
	}
	return objectReader[T]{
		decoder:    decoder,
		header:     header,
		trimReader: trimReader,
		csvWriter:  csvWriter,
	}, nil
}
