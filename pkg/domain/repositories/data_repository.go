package repositories

import (
	"cassandra-benchmark/pkg/domain/entities"
)

type DataRepository interface {
	LoadEmployeeTemplates() ([]entities.EmployeeTemplate, error)
	GenerateEmployee(id int, templates []entities.EmployeeTemplate) entities.Employee
}