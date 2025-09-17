package benchmark

import (
	"cassandra-benchmark/pkg/domain/entities"
	"cassandra-benchmark/pkg/domain/repositories"
	"cassandra-benchmark/pkg/domain/services"
	"context"
	"fmt"
)

type BenchmarkServiceImpl struct {
	workloadService services.WorkloadService
	employeeRepo    repositories.EmployeeRepository
	loggingRepo     repositories.LoggingRepository
	dataRepo        repositories.DataRepository
}

func NewBenchmarkService(
	workloadService services.WorkloadService,
	employeeRepo repositories.EmployeeRepository,
	loggingRepo repositories.LoggingRepository,
	dataRepo repositories.DataRepository,
) services.BenchmarkService {
	return &BenchmarkServiceImpl{
		workloadService: workloadService,
		employeeRepo:    employeeRepo,
		loggingRepo:     loggingRepo,
		dataRepo:        dataRepo,
	}
}

func (bs *BenchmarkServiceImpl) RunBenchmark(ctx context.Context, config entities.BenchmarkConfig) error {
	err := bs.loggingRepo.Start()
	if err != nil {
		return fmt.Errorf("failed to start logging: %w", err)
	}
	defer bs.loggingRepo.Stop()

	err = bs.workloadService.ExecuteWorkload(ctx, config, bs.employeeRepo, bs.loggingRepo, bs.dataRepo)
	if err != nil {
		return fmt.Errorf("workload execution failed: %w", err)
	}

	return nil
}
