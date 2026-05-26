package main

import (
    "bufio"
    "encoding/json"
    "os"
    "sync"
	"context"
)

type FileLogger struct {
    file     *os.File
    writer   *bufio.Writer
    logChan  chan interface{}
    stopChan chan struct{}
    wg       sync.WaitGroup
}

func NewFileLogger(path string,logChan chan interface{}) (*FileLogger) {
    f, _ := os.Create(path) 
    l := &FileLogger{
        file:     f,
        writer:   bufio.NewWriter(f),
        logChan:  logChan,
        stopChan: make(chan struct{}),
    }
    l.wg.Add(1)  
    return l
}

func (l *FileLogger) Run(ctx context.Context) {
    defer l.wg.Done()
    for {
        select {
        case entry := <-l.logChan:
            data, _ := json.Marshal(entry)
            l.writer.Write(data)
            l.writer.WriteByte('\n')
            l.writer.Flush()
        case <-ctx.Done(): 
            for {
                select {
                case entry := <-l.logChan:
                    data, _ := json.Marshal(entry)
                    l.writer.Write(data)
                    l.writer.WriteByte('\n')
                default:
                    l.writer.Flush()
                    l.file.Close()
                    return
                }
            }
        }
    }
} 

func (l *FileLogger) Flush() {
    l.writer.Flush()
} 