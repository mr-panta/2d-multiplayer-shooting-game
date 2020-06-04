package util

import (
	"fmt"
	"log"
	"os"
	"time"
)

type LogPrinter interface {
	Printf(format string, args ...interface{})
}

type logPrinter struct {
	file *os.File
}

func NewLogPrinter(fileName string) (LogPrinter, error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &logPrinter{file: file}, nil
}

func (l *logPrinter) Printf(format string, args ...interface{}) {
	format = fmt.Sprintf("%s%s\n", time.Now().Format("2006-01-02 03:04:05.000000"), format)
	_, err := l.file.WriteString(fmt.Sprintf(format, args...))
	if err != nil {
		log.Fatal(err)
	}
}
