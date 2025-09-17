package entities

import (
	"time"
)

type Employee struct {
	ID        string    `json:"id"`
	Category  string    `json:"category"`
	Data      string    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	Value     int64     `json:"value"`
}

type EmployeeTemplate struct {
	Name      string
	Dept      string
	MinSalary int
	MaxSalary int
}