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

func ConnectToAll(nodes []string, keyspace string) []*gocql.Session {
	var sessions []*gocql.Session

	if len(nodes) == 0 {
		log.Fatalf("No Cassandra nodes specified in configuration")
	}

	for _, node := range nodes {
		cluster := gocql.NewCluster(node)
		cluster.Keyspace = keyspace
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
		log.Fatalf("Failed to connect to any Cassandra nodes. Please check your configuration and network connectivity.")
	}

	log.Printf("Successfully connected to %d/%d Cassandra nodes", len(sessions), len(nodes))
	return sessions
}
