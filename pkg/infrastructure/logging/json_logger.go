package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"cassandra-benchmark/pkg/domain/entities"
)

type WriteLog struct {
	WorkerID int               `json:"worker_id"`
	Time     string            `json:"timestamp"`
	Action   string            `json:"action"`
	Employee entities.Employee `json:"employee"`
	Duration string            `json:"duration"`
	Success  bool              `json:"success"`
	Error    string            `json:"error,omitempty"`
	NodeID   int               `json:"node_id"`
}

type ReadLog struct {
	WorkerID         int                `json:"worker_id"`
	Time             string             `json:"timestamp"`
	Action           string             `json:"action"`
	NodeID           int                `json:"node_id"`
	ID               string             `json:"id"`
	Category         string             `json:"category"`
	Data             string             `json:"data"`
	RecordTimestamp  string             `json:"record_timestamp"`
	Value            int64              `json:"value"`
	Duration         string             `json:"duration"`
	Success          bool               `json:"success"`
	Error            string             `json:"error,omitempty"`
	ReturnedEmployee *entities.Employee `json:"returned_employee,omitempty"`
}

type JsonLoggingRepository struct {
	logChannel chan interface{}
	logFile    *os.File
	flushDone  chan bool
	needsComma bool
}

const (
	flushInterval  = 3 * time.Second
	batchFlushSize = 500
	resultFileName = "result.json"
)

func NewJsonLoggingRepository(bufferSize int) *JsonLoggingRepository {
	return &JsonLoggingRepository{
		logChannel: make(chan interface{}, bufferSize),
		needsComma: false,
	}
}

func (jlr *JsonLoggingRepository) Start() error {
	var err error
	jlr.logFile, err = os.Create(resultFileName)
	if err != nil {
		return fmt.Errorf("failed to create result.json: %w", err)
	}

	_, err = jlr.logFile.Write([]byte("[\n"))
	if err != nil {
		return fmt.Errorf("failed to write opening bracket: %w", err)
	}

	jlr.flushDone = make(chan bool)
	go jlr.backgroundFlusher()
	return nil
}

func (jlr *JsonLoggingRepository) Stop() error {
	jlr.flushDone <- true
	return nil
}

func (jlr *JsonLoggingRepository) LogWrite(result entities.WriteResult) error {
	entry := WriteLog{
		WorkerID: result.WorkerID,
		NodeID:   result.NodeID,
		Time:     result.Timestamp.Format(time.RFC3339),
		Action:   "write",
		Employee: result.Employee,
		Duration: result.Duration.String(),
		Success:  result.Success,
	}
	if result.Error != nil {
		entry.Error = result.Error.Error()
	}
	
	select {
	case jlr.logChannel <- entry:
		return nil
	default:
		return fmt.Errorf("log channel is full")
	}
}

func (jlr *JsonLoggingRepository) LogRead(result entities.ReadResult) error {
	entry := ReadLog{
		WorkerID: result.WorkerID,
		Time:     result.Timestamp.Format(time.RFC3339),
		Action:   "read",
		NodeID:   result.NodeID,
		ID:       result.EmployeeID,
		Duration: result.Duration.String(),
		Success:  result.Success,
	}
	
	if result.Success && result.Employee != nil {
		entry.Category = result.Employee.Category
		entry.Data = result.Employee.Data
		entry.RecordTimestamp = result.Employee.Timestamp.Format(time.RFC3339)
		entry.Value = result.Employee.Value
		entry.ReturnedEmployee = result.Employee
	} else if result.Error != nil {
		entry.Error = result.Error.Error()
	}
	
	select {
	case jlr.logChannel <- entry:
		return nil
	default:
		return fmt.Errorf("log channel is full")
	}
}

func (jlr *JsonLoggingRepository) backgroundFlusher() {
	buffer := make([]interface{}, 0, batchFlushSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case entry := <-jlr.logChannel:
			buffer = append(buffer, entry)
			if len(buffer) >= batchFlushSize {
				jlr.flushBuffer(buffer)
				buffer = buffer[:0]
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				jlr.flushBuffer(buffer)
				buffer = buffer[:0]
			}
		case <-jlr.flushDone:
			if len(buffer) > 0 {
				jlr.flushBuffer(buffer)
			}
			jlr.logFile.Write([]byte("\n]"))
			jlr.logFile.Close()
			return
		}
	}
}

func (jlr *JsonLoggingRepository) flushBuffer(entries []interface{}) {
	for _, entry := range entries {
		data, err := json.Marshal(entry)
		if err != nil {
			continue
		}
		if jlr.needsComma {
			jlr.logFile.Write([]byte(",\n"))
		}
		jlr.logFile.Write(data)
		jlr.needsComma = true
	}
}