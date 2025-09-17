package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"cassandra-benchmark/pkg/domain/entities"
	"cassandra-benchmark/pkg/domain/repositories"
	"github.com/gocql/gocql"
)

type CassandraSessionManager struct {
	sessions []*gocql.Session
}

type CassandraSession struct {
	session *gocql.Session
}

type CassandraQuery struct {
	query *gocql.Query
}

func NewCassandraSessionManager(config entities.DatabaseConfig) (*CassandraSessionManager, error) {
	var sessions []*gocql.Session

	if len(config.Hosts) == 0 {
		return nil, fmt.Errorf("no Cassandra nodes specified in configuration")
	}

	for _, node := range config.Hosts {
		cluster := gocql.NewCluster(node)
		cluster.Keyspace = config.Keyspace
		cluster.Consistency = gocql.Quorum
		cluster.ProtoVersion = 4

		session, err := cluster.CreateSession()
		if err != nil {
			log.Printf("Warning: Failed to connect to %s: %v", node, err)
			continue
		}
		sessions = append(sessions, session)
	}

	if len(sessions) == 0 {
		return nil, fmt.Errorf("failed to connect to any Cassandra nodes")
	}

	log.Printf("Successfully connected to %d/%d Cassandra nodes", len(sessions), len(config.Hosts))
	return &CassandraSessionManager{sessions: sessions}, nil
}

func (csm *CassandraSessionManager) GetSession(nodeIndex int) repositories.Session {
	if nodeIndex >= len(csm.sessions) {
		nodeIndex = nodeIndex % len(csm.sessions)
	}
	return &CassandraSession{session: csm.sessions[nodeIndex]}
}

func (csm *CassandraSessionManager) GetSessionCount() int {
	return len(csm.sessions)
}

func (csm *CassandraSessionManager) Close() error {
	for _, session := range csm.sessions {
		session.Close()
	}
	return nil
}

func (cs *CassandraSession) Query(stmt string, values ...interface{}) repositories.Query {
	return &CassandraQuery{query: cs.session.Query(stmt, values...)}
}

func (cs *CassandraSession) Close() {
	cs.session.Close()
}

func (cq *CassandraQuery) Exec() error {
	return cq.query.Exec()
}

func (cq *CassandraQuery) Scan(dest ...interface{}) error {
	return cq.query.Scan(dest...)
}

type CassandraEmployeeRepository struct {
	sessionManager repositories.SessionManager
	tableName      string
}

func NewCassandraEmployeeRepository(sessionManager repositories.SessionManager, tableName string) *CassandraEmployeeRepository {
	return &CassandraEmployeeRepository{
		sessionManager: sessionManager,
		tableName:      tableName,
	}
}

func (cer *CassandraEmployeeRepository) Create(ctx context.Context, employee entities.Employee) error {
	uuid, err := gocql.ParseUUID(employee.ID)
	if err != nil {
		return err
	}

	nodeIndex := 0
	if cer.sessionManager.GetSessionCount() > 1 {
		nodeIndex = int(time.Now().UnixNano()) % cer.sessionManager.GetSessionCount()
	}
	
	session := cer.sessionManager.GetSession(nodeIndex)
	query := fmt.Sprintf(`INSERT INTO %s (id, category, data, timestamp, value) VALUES (?, ?, ?, ?, ?)`, cer.tableName)
	
	return session.Query(query, uuid, employee.Category, employee.Data, employee.Timestamp, employee.Value).Exec()
}

func (cer *CassandraEmployeeRepository) GetByID(ctx context.Context, id string) (*entities.Employee, error) {
	uuid, err := gocql.ParseUUID(id)
	if err != nil {
		return nil, err
	}

	nodeIndex := 0
	if cer.sessionManager.GetSessionCount() > 1 {
		nodeIndex = int(time.Now().UnixNano()) % cer.sessionManager.GetSessionCount()
	}
	
	session := cer.sessionManager.GetSession(nodeIndex)
	query := fmt.Sprintf(`SELECT category, data, timestamp, value FROM %s WHERE id = ?`, cer.tableName)
	
	var category, data string
	var timestamp time.Time
	var value int64
	
	err = session.Query(query, uuid).Scan(&category, &data, &timestamp, &value)
	if err != nil {
		return nil, err
	}
	
	return &entities.Employee{
		ID:        id,
		Category:  category,
		Data:      data,
		Timestamp: timestamp,
		Value:     value,
	}, nil
}

func (cer *CassandraEmployeeRepository) Close() error {
	return cer.sessionManager.Close()
}