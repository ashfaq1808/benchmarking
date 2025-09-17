package workload

import (
	"encoding/csv"
	"math/rand"
	"os"
	"strconv"
	"sync"
)

type Employee struct {
	ID        string
	Category  string
	Data      string
	Timestamp string
	Value     int64
}

type EmployeeTemplate struct {
	Name      string
	Dept      string
	MinSalary int
	MaxSalary int
}

var (
	employeeTemplates []EmployeeTemplate
	templatesOnce     sync.Once
	templatesLoaded   bool
)

func loadEmployeeTemplates() error {
	file, err := os.Open("employees_data.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	employeeTemplates = make([]EmployeeTemplate, 0, len(records)-1)
	
	for i, record := range records {
		if i == 0 {
			continue
		}
		
		minSalary, err := strconv.Atoi(record[2])
		if err != nil {
			continue
		}
		
		maxSalary, err := strconv.Atoi(record[3])
		if err != nil {
			continue
		}
		
		employeeTemplates = append(employeeTemplates, EmployeeTemplate{
			Name:      record[0],
			Dept:      record[1],
			MinSalary: minSalary,
			MaxSalary: maxSalary,
		})
	}
	
	templatesLoaded = true
	return nil
}

func GenerateEmployee(id int) Employee {
	templatesOnce.Do(func() {
		if err := loadEmployeeTemplates(); err != nil {
			employeeTemplates = []EmployeeTemplate{
				{"Alice", "Engineering", 75000, 120000},
				{"Bob", "HR", 45000, 80000},
				{"Charlie", "Finance", 60000, 100000},
				{"Diana", "Sales", 50000, 95000},
				{"Eve", "Legal", 70000, 110000},
			}
			templatesLoaded = true
		}
	})

	if !templatesLoaded || len(employeeTemplates) == 0 {
		return Employee{
			ID:        strconv.Itoa(id),
			Category:  "general",
			Data:      "default_employee",
			Timestamp: "2025-01-01T00:00:00Z",
			Value:     50000,
		}
	}

	template := employeeTemplates[rand.Intn(len(employeeTemplates))]
	salaryRange := template.MaxSalary - template.MinSalary
	salary := template.MinSalary + rand.Intn(salaryRange+1)

	return Employee{
		ID:        strconv.Itoa(id),
		Category:  template.Dept,
		Data:      template.Name,
		Timestamp: "2025-01-01T00:00:00Z",
		Value:     int64(salary),
	}
}
