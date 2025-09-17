package repositories

import (
	"context"

	"cassandra-benchmark/pkg/domain/entities"
)

type EmployeeRepository interface {
	Create(ctx context.Context, employee entities.Employee) error
	GetByID(ctx context.Context, id string) (*entities.Employee, error)
	Close() error
}

type SessionManager interface {
	GetSession(nodeIndex int) Session
	GetSessionCount() int
	Close() error
}

type Session interface {
	Query(stmt string, values ...interface{}) Query
	Close()
}

type Query interface {
	Exec() error
	Scan(dest ...interface{}) error
}