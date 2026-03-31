package backup

import (
	"bytes"
	"io"
	"log"
	"os"
	"time"
)

var (
	// LogBuf accumulates log output to be saved or emailed.
	LogBuf    bytes.Buffer
	Hostname  string
	Filestamp string

	// logger is a package-local logger so we don't touch the global stdlib logger.
	logger *log.Logger
)

func init() {
	h, err := os.Hostname()
	if err != nil {
		h = "unknown"
	}
	Hostname = h
	Filestamp = time.Now().Format("20060102_1504")

	logger = log.New(&LogBuf, "", log.Ldate|log.Ltime)
}

// InitLog enables stdout output in addition to the buffer when verbose is true.
func InitLog(verbose bool) {
	if verbose {
		logger.SetOutput(io.MultiWriter(&LogBuf, os.Stdout))
	}
}

// LogMsg writes a message to the log buffer (and stdout when verbose).
func LogMsg(msg string) {
	logger.Println(msg)
}

// SaveLog writes the accumulated log buffer to a file on disk.
func SaveLog(logFile string) {
	if err := os.WriteFile(logFile, LogBuf.Bytes(), 0644); err != nil {
		os.Stderr.WriteString("ERROR saving log file: " + err.Error() + "\n")
	}
}
