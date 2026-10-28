package service

import (
	"errors"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

var (
	ErrInvalidPDF = errors.New("invalid pdf")
	ErrEmptyPDF   = errors.New("no text found in pdf")
)

func ExtractText(file io.ReaderAt, size int64) (string, error) {
	reader, err := pdf.NewReader(file, size)
	if err != nil {
		return "", ErrInvalidPDF
	}

	plainText, err := reader.GetPlainText()
	if err != nil {
		return "", ErrInvalidPDF
	}

	content, err := io.ReadAll(plainText)
	if err != nil {
		return "", ErrInvalidPDF
	}

	text := strings.TrimSpace(string(content))
	if text == "" {
		return "", ErrEmptyPDF
	}

	return text, nil
}
