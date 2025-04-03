package core

import (
    "github.com/gocql/gocql"
    "math/rand"
)

type CassandraEngine struct {
    Config   *Config
    Sessions []*gocql.Session
}

func (e *CassandraEngine) Connect() error {
    for _, host := range e.Config.Hosts {
        cluster := gocql.NewCluster(host)
        cluster.Keyspace = e.Config.Keyspace
        cluster.Consistency = gocql.Quorum
        session, err := cluster.CreateSession()
        if err != nil {
            return err
        }
        e.Sessions = append(e.Sessions, session)
    }
    return nil
}

func (e *CassandraEngine) GetRandomSession() *gocql.Session {
    return e.Sessions[rand.Intn(len(e.Sessions))]
}

func (e *CassandraEngine) Close() {
    for _, session := range e.Sessions {
        session.Close()
    }
}
