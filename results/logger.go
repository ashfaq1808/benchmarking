package result

import (
	"benchmarking/workload"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type WriteLog struct {
	WorkerID int               `json:"worker_id"`
	Time     string            `json:"timestamp"`
	Action   string            `json:"action"`
	Employee workload.Employee `json:"employee"`
	Duration string            `json:"duration"`
	Success  bool              `json:"success"`
	Error    string            `json:"error,omitempty"`
}

type ReadLog struct {
	WorkerID int    `json:"worker_id"`
	Time     string `json:"timestamp"`
	Action   string `json:"action"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	Dept     string `json:"dept"`
	Salary   int    `json:"salary"`
	Duration string `json:"duration"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

var (
	logFile *os.File
	logMux  sync.Mutex
)

func init() {
	var err error
	logFile, err = os.OpenFile("result.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Could not open result file:", err)
		os.Exit(1)
	}
}

func LogWrite(workerID int, emp workload.Employee, duration time.Duration, err error) {
	logMux.Lock()
	defer logMux.Unlock()
	entry := WriteLog{
		WorkerID: workerID,
		Time:     time.Now().Format(time.RFC3339),
		Action:   "write",
		Employee: emp,
		Duration: duration.String(),
		Success:  err == nil,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	json.NewEncoder(logFile).Encode(entry)
}

func LogRead(workerID int, id, name, dept string, salary int, duration time.Duration, err error) {
	logMux.Lock()
	defer logMux.Unlock()
	entry := ReadLog{
		WorkerID: workerID,
		Time:     time.Now().Format(time.RFC3339),
		Action:   "read",
		ID:       id,
		Name:     name,
		Dept:     dept,
		Salary:   salary,
		Duration: duration.String(),
		Success:  err == nil,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	json.NewEncoder(logFile).Encode(entry)
}

func CloseLogFile() {
	logFile.Close()
}
