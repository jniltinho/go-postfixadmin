package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"time"
)

var (
	logBuf    bytes.Buffer
	hostname  string
	filestamp string
)

func init() {
	h, err := os.Hostname()
	if err != nil {
		h = "unknown"
	}
	hostname = h
	filestamp = time.Now().Format("20060102_1504")

	log.SetFlags(log.Ldate | log.Ltime)
	log.SetOutput(&logBuf)
}

// initLog must be called after flag parsing to wire up stdout when verbose.
func initLog(verbose bool) {
	if verbose {
		log.SetOutput(io.MultiWriter(&logBuf, os.Stdout))
	}
}

func logMsg(msg string) {
	log.Println(msg)
}

func saveLog(logFile string) {
	if err := os.WriteFile(logFile, logBuf.Bytes(), 0644); err != nil {
		os.Stderr.WriteString("ERROR saving log file: " + err.Error() + "\n")
	}
}
