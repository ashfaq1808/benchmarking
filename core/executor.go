package core

import (
    "fmt"
    "math/rand"
    "sync"
    "time"
)

var writtenKeys []string
var keyMutex sync.Mutex

func RunBenchmark(e *CassandraEngine, cfg *Config) {
    var wg sync.WaitGroup
    totalOps := cfg.TotalReads + cfg.TotalWrites
    opsPerWorker := totalOps / cfg.Concurrency

    readTarget := cfg.TotalReads
    writeTarget := cfg.TotalWrites

    for i := 0; i < cfg.Concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            reads, writes := 0, 0
            for j := 0; j < opsPerWorker; j++ {
                if writes < writeTarget && (reads >= readTarget || rand.Intn(2) == 0) {
                    performWrite(e)
                    writes++
                } else if reads < readTarget {
                    performRead(e)
                    reads++
                }
            }
        }()
    }
    wg.Wait()
}

func performWrite(e *CassandraEngine) {
    id := fmt.Sprintf("w_%d", time.Now().UnixNano())
    payload := "payload"
    session := e.GetRandomSession()
    start := time.Now()
    err := session.Query(fmt.Sprintf(`INSERT INTO %s (id, data) VALUES (?, ?)`, e.Config.Table), id, payload).Exec()
    duration := time.Since(start).Milliseconds()

    if err == nil {
        keyMutex.Lock()
        writtenKeys = append(writtenKeys, id)
        keyMutex.Unlock()
        logResult("write", duration, true, "")
    } else {
        logResult("write", duration, false, err.Error())
    }
}

func performRead(e *CassandraEngine) {
    keyMutex.Lock()
    if len(writtenKeys) == 0 {
        keyMutex.Unlock()
        return
    }
    id := writtenKeys[rand.Intn(len(writtenKeys))]
    keyMutex.Unlock()

    var data string
    session := e.GetRandomSession()
    start := time.Now()
    err := session.Query(fmt.Sprintf(`SELECT data FROM %s WHERE id = ?`, e.Config.Table), id).Scan(&data)
    duration := time.Since(start).Milliseconds()

    if err == nil {
        logResult("read", duration, true, "")
    } else {
        logResult("read", duration, false, err.Error())
    }
}

