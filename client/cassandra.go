package client

import (
	"log"

	"github.com/gocql/gocql"
)

func Connect(hosts []string, keyspace string) *gocql.Session {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4

	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Failed to connect to Cassandra: %v", err)
	}
	return session
}
