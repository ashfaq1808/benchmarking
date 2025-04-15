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

	for _, node := range nodes {
		cluster := gocql.NewCluster(node)
		cluster.Keyspace = keyspace
		cluster.Consistency = gocql.Quorum
		cluster.ProtoVersion = 4

		session, err := cluster.CreateSession()
		if err != nil {
			log.Fatalf("Failed to connect to %s: %v", node, err)
		}
		sessions = append(sessions, session)
	}

	return sessions
}
