package repositories

import (
	"cassandra-benchmark/pkg/domain/entities"
)

type LoggingRepository interface {
	LogWrite(result entities.WriteResult) error
	LogRead(result entities.ReadResult) error
	Start() error
	Stop() error
}