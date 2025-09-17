package core

import (
	"math/rand"

	"github.com/gocql/gocql"
)

type CassandraEngine struct {
	Config   *Config
	Sessions []*gocql.Session
}

// Connect establishes connections to the Cassandra cluster.
func (e *CassandraEngine) Connect() error {
	cluster := gocql.NewCluster("155.98.38.146", "155.98.38.130") // Use the correct IPs for your nodes
	cluster.Keyspace = "benchmark"
	cluster.Consistency = gocql.Quorum

	// Create multiple sessions if needed (one per node)
	for i := 0; i < 2; i++ { // Assuming 2 nodes are available
		session, err := cluster.CreateSession()
		if err != nil {
			return err // return error so it can be handled higher up the stack
		}
		e.Sessions = append(e.Sessions, session)
	}

	return nil
}

// GetRandomSession returns a random session from the pool of sessions
func (e *CassandraEngine) GetRandomSession() *gocql.Session {
	if len(e.Sessions) == 0 {
		return nil // Handle case where no session is available
	}
	return e.Sessions[rand.Intn(len(e.Sessions))]
}

// Close closes all the active Cassandra sessions
func (e *CassandraEngine) Close() {
	for _, session := range e.Sessions {
		session.Close()
	}
}
