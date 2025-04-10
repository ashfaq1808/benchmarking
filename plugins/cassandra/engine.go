package cassandra

import (
	"benchmark_tool/core"
	"fmt"
	"math/rand"
	"time"

	"github.com/gocql/gocql"
)

type CassandraPlugin struct {
	Config   *core.Config
	Sessions []*gocql.Session
}

func (p *CassandraPlugin) Connect() error {
	for _, host := range p.Config.Hosts {
		cluster := gocql.NewCluster(host)
		cluster.Keyspace = p.Config.Keyspace
		cluster.Consistency = gocql.Quorum
        cluster.Timeout = 30 * time.Second
		session, err := cluster.CreateSession()
		if err != nil {
			return err
		}
		p.Sessions = append(p.Sessions, session)
	}
	return nil
}

func (p *CassandraPlugin) GetRandomSession() *gocql.Session {
	return p.Sessions[rand.Intn(len(p.Sessions))]
}

func (p *CassandraPlugin) Close() {
	for _, session := range p.Sessions {
		session.Close()
	}
}

func (p *CassandraPlugin) Write() error {
	id := fmt.Sprintf("p_%d", time.Now().UnixNano())
	session := p.GetRandomSession()
	return session.Query(fmt.Sprintf("INSERT INTO %s (id, data) VALUES (?, ?)", p.Config.Table), id, "payload").Exec()
}

func (p *CassandraPlugin) Read(id string) (string, error) {
	var data string
	session := p.GetRandomSession()
	err := session.Query(fmt.Sprintf("SELECT data FROM %s WHERE id = ?", p.Config.Table), id).Scan(&data)
	return data, err
}
