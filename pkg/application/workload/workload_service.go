package workload

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"cassandra-benchmark/pkg/domain/entities"
	"cassandra-benchmark/pkg/domain/repositories"
	"cassandra-benchmark/pkg/domain/services"
)

type WorkloadServiceImpl struct {
	writtenIDs []string
	idMutex    sync.Mutex
}

func NewWorkloadService() services.WorkloadService {
	return &WorkloadServiceImpl{
		writtenIDs: make([]string, 0),
	}
}

func (ws *WorkloadServiceImpl) addWrittenID(id string) {
	ws.idMutex.Lock()
	defer ws.idMutex.Unlock()
	ws.writtenIDs = append(ws.writtenIDs, id)
}

func (ws *WorkloadServiceImpl) getRandomWrittenID() string {
	ws.idMutex.Lock()
	defer ws.idMutex.Unlock()
	if len(ws.writtenIDs) == 0 {
		return ""
	}
	return ws.writtenIDs[rand.Intn(len(ws.writtenIDs))]
}

func (ws *WorkloadServiceImpl) ExecuteWorkload(
	ctx context.Context,
	config entities.BenchmarkConfig,
	employeeRepo repositories.EmployeeRepository,
	loggingRepo repositories.LoggingRepository,
	dataRepo repositories.DataRepository,
) error {
	templates, err := dataRepo.LoadEmployeeTemplates()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	startTime := time.Now()
	warmupCutoff := startTime.Add(time.Duration(config.WarmupSeconds) * time.Second)
	endTime := warmupCutoff.Add(time.Duration(config.DurationSeconds) * time.Second)

	if config.Mode == "open-loop" && config.RequestsPerSecond > 0 {
		return ws.executeOpenLoopWorkload(ctx, config, employeeRepo, loggingRepo, dataRepo, templates, warmupCutoff, endTime, &wg)
	} else {
		return ws.executeClosedLoopWorkload(ctx, config, employeeRepo, loggingRepo, dataRepo, templates, warmupCutoff, endTime, &wg)
	}
}

func (ws *WorkloadServiceImpl) executeOpenLoopWorkload(
	ctx context.Context,
	config entities.BenchmarkConfig,
	employeeRepo repositories.EmployeeRepository,
	loggingRepo repositories.LoggingRepository,
	dataRepo repositories.DataRepository,
	templates []entities.EmployeeTemplate,
	warmupCutoff, endTime time.Time,
	wg *sync.WaitGroup,
) error {
	var currentRate int
	var ticker *time.Ticker

	if config.RatePattern.Enabled {
		currentRate = config.RatePattern.MinRate
	} else {
		currentRate = config.RequestsPerSecond
	}

	ticker = time.NewTicker(time.Second / time.Duration(currentRate))
	defer ticker.Stop()

	var phaseStart time.Time
	var inPeakPhase bool
	var lastRateChange time.Time

	if config.RatePattern.Enabled {
		if config.RatePattern.Mode == "cycles" {
			phaseStart = time.Now()
			inPeakPhase = false
		} else if config.RatePattern.Mode == "random" {
			lastRateChange = time.Now()
		}
	}

	for time.Now().Before(endTime) {
		select {
		case <-ctx.Done():
			wg.Wait()
			return ctx.Err()
		default:
		}

		if config.RatePattern.Enabled {
			currentRate, ticker = ws.updateRatePattern(config.RatePattern, currentRate, ticker, &phaseStart, &inPeakPhase, &lastRateChange)
		}

		<-ticker.C
		wg.Add(1)
		go func() {
			defer wg.Done()
			ws.executeOperation(config, employeeRepo, loggingRepo, dataRepo, templates, warmupCutoff)
		}()
	}

	wg.Wait()
	return nil
}

func (ws *WorkloadServiceImpl) executeClosedLoopWorkload(
	ctx context.Context,
	config entities.BenchmarkConfig,
	employeeRepo repositories.EmployeeRepository,
	loggingRepo repositories.LoggingRepository,
	dataRepo repositories.DataRepository,
	templates []entities.EmployeeTemplate,
	warmupCutoff, endTime time.Time,
	wg *sync.WaitGroup,
) error {
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for time.Now().Before(endTime) {
				select {
				case <-ctx.Done():
					return
				default:
				}

				ws.executeOperation(config, employeeRepo, loggingRepo, dataRepo, templates, warmupCutoff)

				if config.Mode == "closed-loop" {
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()
	return nil
}

func (ws *WorkloadServiceImpl) executeOperation(
	config entities.BenchmarkConfig,
	employeeRepo repositories.EmployeeRepository,
	loggingRepo repositories.LoggingRepository,
	dataRepo repositories.DataRepository,
	templates []entities.EmployeeTemplate,
	warmupCutoff time.Time,
) {
	workerID := rand.Intn(config.Concurrency)
	nodeID := rand.Intn(1) // This would come from session manager

	if rand.Float64() < config.WriteRatio {
		ws.executeWriteOperation(employeeRepo, loggingRepo, dataRepo, templates, workerID, nodeID, warmupCutoff)
	} else {
		ws.executeReadOperation(employeeRepo, loggingRepo, workerID, nodeID, warmupCutoff)
	}
}

func (ws *WorkloadServiceImpl) executeWriteOperation(
	employeeRepo repositories.EmployeeRepository,
	loggingRepo repositories.LoggingRepository,
	dataRepo repositories.DataRepository,
	templates []entities.EmployeeTemplate,
	workerID, nodeID int,
	warmupCutoff time.Time,
) {
	employee := dataRepo.GenerateEmployee(rand.Int(), templates)
	
	start := time.Now()
	err := employeeRepo.Create(context.Background(), employee)
	duration := time.Since(start)

	if time.Now().After(warmupCutoff) {
		result := entities.WriteResult{
			OperationResult: entities.OperationResult{
				WorkerID:  workerID,
				NodeID:    nodeID,
				Operation: "write",
				Duration:  duration,
				Success:   err == nil,
				Error:     err,
				Timestamp: time.Now(),
			},
			Employee: employee,
		}
		loggingRepo.LogWrite(result)
	}

	if err == nil {
		ws.addWrittenID(employee.ID)
	}
}

func (ws *WorkloadServiceImpl) executeReadOperation(
	employeeRepo repositories.EmployeeRepository,
	loggingRepo repositories.LoggingRepository,
	workerID, nodeID int,
	warmupCutoff time.Time,
) {
	id := ws.getRandomWrittenID()
	if id == "" {
		return
	}

	start := time.Now()
	employee, err := employeeRepo.GetByID(context.Background(), id)
	duration := time.Since(start)

	if time.Now().After(warmupCutoff) {
		result := entities.ReadResult{
			OperationResult: entities.OperationResult{
				WorkerID:  workerID,
				NodeID:    nodeID,
				Operation: "read",
				Duration:  duration,
				Success:   err == nil,
				Error:     err,
				Timestamp: time.Now(),
			},
			EmployeeID: id,
			Employee:   employee,
		}
		loggingRepo.LogRead(result)
	}
}

func (ws *WorkloadServiceImpl) updateRatePattern(
	ratePattern entities.RatePatternConfig,
	currentRate int,
	ticker *time.Ticker,
	phaseStart *time.Time,
	inPeakPhase *bool,
	lastRateChange *time.Time,
) (int, *time.Ticker) {
	now := time.Now()

	if ratePattern.Mode == "cycles" {
		phaseDuration := now.Sub(*phaseStart)

		var shouldSwitchPhase bool
		var newRate int

		if *inPeakPhase {
			shouldSwitchPhase = phaseDuration >= time.Duration(ratePattern.PeakDuration)*time.Second
			newRate = ratePattern.MinRate
		} else {
			shouldSwitchPhase = phaseDuration >= time.Duration(ratePattern.ValleyDuration)*time.Second
			newRate = ratePattern.MaxRate
		}

		if shouldSwitchPhase {
			*inPeakPhase = !*inPeakPhase
			*phaseStart = now
			currentRate = newRate
			ticker.Stop()
			ticker = time.NewTicker(time.Second / time.Duration(currentRate))
		}
	} else if ratePattern.Mode == "random" {
		timeSinceLastChange := now.Sub(*lastRateChange)

		if timeSinceLastChange >= time.Duration(ratePattern.ChangeInterval*1000)*time.Millisecond {
			rateRange := ratePattern.MaxRate - ratePattern.MinRate
			newRate := ratePattern.MinRate + rand.Intn(rateRange+1)

			if newRate != currentRate {
				currentRate = newRate
				ticker.Stop()
				ticker = time.NewTicker(time.Second / time.Duration(currentRate))
			}
			*lastRateChange = now
		}
	}

	return currentRate, ticker
}