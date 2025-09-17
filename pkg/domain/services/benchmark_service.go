package services

import (
	"context"
	"cassandra-benchmark/pkg/domain/entities"
	"cassandra-benchmark/pkg/domain/repositories"
)

type BenchmarkService interface {
	RunBenchmark(ctx context.Context, config entities.BenchmarkConfig) error
}

type WorkloadService interface {
	ExecuteWorkload(ctx context.Context, config entities.BenchmarkConfig, employeeRepo repositories.EmployeeRepository, loggingRepo repositories.LoggingRepository, dataRepo repositories.DataRepository) error
}

type EmployeeService interface {
	CreateEmployee(ctx context.Context, employee entities.Employee) error
	GetEmployee(ctx context.Context, id string) (*entities.Employee, error)
}