package result

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var (
	logFile    *os.File
	flushDone  chan bool
	needsComma bool = false
)

const (
	flushInterval  = 3 * time.Second
	batchFlushSize = 500
	resultFileName = "result.json"
)

func StartFlusher() {
	var err error
	logFile, err = os.Create(resultFileName)
	if err != nil {
		fmt.Println("Failed to create result.json:", err)
		os.Exit(1)
	}

	// Write the opening bracket for JSON array
	_, err = logFile.Write([]byte("[\n"))
	if err != nil {
		fmt.Println("Failed to write opening [ :", err)
		os.Exit(1)
	}

	flushDone = make(chan bool)
	go backgroundFlusher()
}

func backgroundFlusher() {
	buffer := make([]interface{}, 0, batchFlushSize)
	ticker := time.NewTicker(flushInterval)

	for {
		select {
		case entry := <-LogChannel:
			buffer = append(buffer, entry)
			if len(buffer) >= batchFlushSize {
				flushBuffer(buffer)
				buffer = buffer[:0]
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				flushBuffer(buffer)
				buffer = buffer[:0]
			}
		case <-flushDone:
			if len(buffer) > 0 {
				flushBuffer(buffer)
			}

			// Write the closing bracket for JSON array
			logFile.Write([]byte("\n]"))
			logFile.Close()
			return
		}
	}
}

func flushBuffer(entries []interface{}) {
	for _, entry := range entries {
		data, err := json.Marshal(entry)
		if err != nil {
			continue
		}
		if needsComma {
			logFile.Write([]byte(",\n"))
		}
		logFile.Write(data)
		needsComma = true
	}
}

func StopFlusher() {
	flushDone <- true
}
