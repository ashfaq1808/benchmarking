package core

import (
	"log"
	"math/rand"

	"github.com/gocql/gocql"
)

type CassandraEngine struct {
	Config   *Config
	Sessions []*gocql.Session
}

func (e *CassandraEngine) Connect() error {
	cluster := gocql.NewCluster("155.98.38.146", "155.98.38.130") // Add your actual IP addresses
	cluster.Keyspace = "benchmark"
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal("Connection failed: ", err)
	}
	defer session.Close()

}

func (e *CassandraEngine) GetRandomSession() *gocql.Session {
	return e.Sessions[rand.Intn(len(e.Sessions))]
}

func (e *CassandraEngine) Close() {
	for _, session := range e.Sessions {
		session.Close()
	}
}
