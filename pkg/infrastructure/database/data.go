package database

import (
	"encoding/csv"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"cassandra-benchmark/pkg/domain/entities"
	"cassandra-benchmark/pkg/domain/repositories"
	"github.com/gocql/gocql"
)

type CsvDataRepository struct {
	csvFilePath string
	templates   []entities.EmployeeTemplate
	once        sync.Once
	loaded      bool
}

func NewCsvDataRepository(csvFilePath string) repositories.DataRepository {
	return &CsvDataRepository{
		csvFilePath: csvFilePath,
	}
}

func (cdr *CsvDataRepository) LoadEmployeeTemplates() ([]entities.EmployeeTemplate, error) {
	var err error
	cdr.once.Do(func() {
		err = cdr.loadTemplates()
	})
	return cdr.templates, err
}

func (cdr *CsvDataRepository) loadTemplates() error {
	file, err := os.Open(cdr.csvFilePath)
	if err != nil {
		cdr.templates = cdr.getDefaultTemplates()
		cdr.loaded = true
		return nil
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		cdr.templates = cdr.getDefaultTemplates()
		cdr.loaded = true
		return nil
	}

	cdr.templates = make([]entities.EmployeeTemplate, 0, len(records)-1)

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

		cdr.templates = append(cdr.templates, entities.EmployeeTemplate{
			Name:      record[0],
			Dept:      record[1],
			MinSalary: minSalary,
			MaxSalary: maxSalary,
		})
	}

	if len(cdr.templates) == 0 {
		cdr.templates = cdr.getDefaultTemplates()
	}

	cdr.loaded = true
	return nil
}

func (cdr *CsvDataRepository) getDefaultTemplates() []entities.EmployeeTemplate {
	return []entities.EmployeeTemplate{
		{"Alice", "Engineering", 75000, 120000},
		{"Bob", "HR", 45000, 80000},
		{"Charlie", "Finance", 60000, 100000},
		{"Diana", "Sales", 50000, 95000},
		{"Eve", "Legal", 70000, 110000},
	}
}

func (cdr *CsvDataRepository) GenerateEmployee(id int, templates []entities.EmployeeTemplate) entities.Employee {
	if len(templates) == 0 {
		return entities.Employee{
			ID:        gocql.TimeUUID().String(),
			Category:  "general",
			Data:      "default_employee",
			Timestamp: time.Now(),
			Value:     50000,
		}
	}

	template := templates[rand.Intn(len(templates))]
	salaryRange := template.MaxSalary - template.MinSalary
	salary := template.MinSalary + rand.Intn(salaryRange+1)

	return entities.Employee{
		ID:        gocql.TimeUUID().String(),
		Category:  template.Dept,
		Data:      template.Name,
		Timestamp: time.Now(),
		Value:     int64(salary),
	}
}