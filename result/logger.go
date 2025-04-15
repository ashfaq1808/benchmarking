package result

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

var logs []interface{}

type Employee struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Dept   string `json:"dept"`
	Salary int    `json:"salary"`
}

type WriteLog struct {
	WorkerID int      `json:"worker_id"`
	Time     string   `json:"timestamp"`
	Action   string   `json:"action"`
	Employee Employee `json:"employee"`
	Duration string   `json:"duration"`
	Success  bool     `json:"success"`
	Error    string   `json:"error,omitempty"`
	NodeID   int      `json:"node_id"`
}

type ReadLog struct {
	WorkerID         int       `json:"worker_id"`
	Time             string    `json:"timestamp"`
	Action           string    `json:"action"`
	NodeID           int       `json:"node_id"`
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Dept             string    `json:"dept"`
	Salary           int       `json:"salary"`
	Duration         string    `json:"duration"`
	Success          bool      `json:"success"`
	Error            string    `json:"error,omitempty"`
	ReturnedEmployee *Employee `json:"returned_employee,omitempty"`
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

func LogWrite(workerID, nodeID int, emp Employee, duration time.Duration, err error) {
	entry := WriteLog{
		WorkerID: workerID,
		NodeID:   nodeID,
		Time:     time.Now().Format(time.RFC3339),
		Action:   "write",
		Employee: emp,
		Duration: duration.String(),
		Success:  err == nil,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	logMux.Lock()
	logs = append(logs, entry)
	logMux.Unlock()
}

func LogRead(workerID, nodeID int, id, name, dept string, salary int, duration time.Duration, err error) {
	entry := ReadLog{
		WorkerID: workerID,
		Time:     time.Now().Format(time.RFC3339),
		Action:   "read",
		NodeID:   nodeID,
		ID:       id,
		Name:     name,
		Dept:     dept,
		Salary:   salary,
		Duration: duration.String(),
		Success:  err == nil,
	}
	if err == nil {
		entry.ReturnedEmployee = &Employee{
			ID:     id,
			Name:   name,
			Dept:   dept,
			Salary: salary,
		}
	} else {
		entry.Error = err.Error()
	}

	logMux.Lock()
	logs = append(logs, entry)
	logMux.Unlock()
}

func CloseLogFile() {
	logMux.Lock()
	defer logMux.Unlock()

	file, err := os.Create("result.json")
	if err != nil {
		fmt.Println("Could not open result.json:", err)
		return
	}
	defer file.Close()

	// encoder := json.NewEncoder(file)
	_, err = file.Write([]byte("[\n"))
	if err != nil {
		fmt.Println("Write error:", err)
		return
	}

	for i, entry := range logs {
		enc, _ := json.Marshal(entry)
		file.Write(enc)
		if i != len(logs)-1 {
			file.Write([]byte(",\n"))
		}
	}
	file.Write([]byte("\n]"))
}

func InitLogFile() {
	file, err := os.Create("result.json")
	if err != nil {
		fmt.Println("Failed to initialize result.json:", err)
		os.Exit(1)
	}
	file.Close()
}
