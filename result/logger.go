package result

import (
	"time"
)

// Employee struct
type Employee struct {
	ID        string `json:"id"`
	Category  string `json:"category"`
	Data      string `json:"data"`
	Timestamp string `json:"timestamp"`
	Value     int64  `json:"value"`
}

// WriteLog struct
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

// ReadLog struct
type ReadLog struct {
	WorkerID         int       `json:"worker_id"`
	Time             string    `json:"timestamp"`
	Action           string    `json:"action"`
	NodeID           int       `json:"node_id"`
	ID               string    `json:"id"`
	Category         string    `json:"category"`
	Data             string    `json:"data"`
	RecordTimestamp  string    `json:"record_timestamp"`
	Value            int64     `json:"value"`
	Duration         string    `json:"duration"`
	Success          bool      `json:"success"`
	Error            string    `json:"error,omitempty"`
	ReturnedEmployee *Employee `json:"returned_employee,omitempty"`
}

// LogChannel is the buffered channel where all log entries are pushed
var LogChannel chan interface{}

// InitializeLogger initializes the log channel
func InitializeLogger(bufferSize int) {
	LogChannel = make(chan interface{}, bufferSize)
}

// LogWrite sends a write operation log into channel
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
	LogChannel <- entry
}

// LogRead sends a read operation log into channel
func LogRead(workerID, nodeID int, id, category, data, recordTimestamp string, value int64, duration time.Duration, err error) {
	entry := ReadLog{
		WorkerID:        workerID,
		Time:            time.Now().Format(time.RFC3339),
		Action:          "read",
		NodeID:          nodeID,
		ID:              id,
		Category:        category,
		Data:            data,
		RecordTimestamp: recordTimestamp,
		Value:           value,
		Duration:        duration.String(),
		Success:         err == nil,
	}
	if err == nil {
		entry.ReturnedEmployee = &Employee{
			ID:        id,
			Category:  category,
			Data:      data,
			Timestamp: recordTimestamp,
			Value:     value,
		}
	} else {
		entry.Error = err.Error()
	}
	LogChannel <- entry
}
