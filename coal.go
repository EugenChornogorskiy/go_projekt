package main

import (
    "fmt"
    "time"
    "context"
)

type CoalPlant struct {
    commandChan   <-chan CoalCommand
    statusChan    chan chan CoalStatus
    stopChan      chan struct{}
    warmUpSteps   int
    warmDownSteps int
}

func NewCoalPlant(commandChan <-chan CoalCommand, statusChan chan chan CoalStatus) *CoalPlant {
    return &CoalPlant{
        commandChan:   commandChan,
        statusChan:    statusChan,
        stopChan:      make(chan struct{}),
        warmUpSteps:   0,
        warmDownSteps: 0,
    }
}

func (c *CoalPlant) Run(ctx context.Context) {
    go func() {
        running := false
        targetMW := 0.0
        currentMW := 0.0
        ticker := time.NewTicker(GridStep)
        defer ticker.Stop()

        for {
            select {
            case cmd := <-c.commandChan:
                if cmd.Start && !running {
                    running = true
                    targetMW = 10.0
                    c.warmUpSteps = 3 
                    c.warmDownSteps = 0
                    fmt.Println("[Coal] Warming up... (3h to full power)")
                } else if !cmd.Start && running {
                    running = false
                    targetMW = 0.0
                    c.warmDownSteps = 2 
                    c.warmUpSteps = 0
                    fmt.Println("[Coal] Cooling down... (2h to stop)")
                }
                if cmd.Response != nil {
                    cmd.Response <- true
                }

            case respCh := <-c.statusChan:
                respCh <- CoalStatus{Running: running || c.warmUpSteps > 0, Mw: currentMW}

            case <-ticker.C: 
                if c.warmUpSteps > 0 {
                    c.warmUpSteps--
                    currentMW = targetMW * (1.0 - float64(c.warmUpSteps)/3.0)
                    if c.warmUpSteps == 0 {
                        fmt.Printf("[Coal] Full power: %.1f MW\n", currentMW)
                    }
                } else if c.warmDownSteps > 0 {
                    c.warmDownSteps--
                    currentMW = targetMW * (float64(c.warmDownSteps) / 2.0)
                    if c.warmDownSteps == 0 {
                        currentMW = 0
                        fmt.Println("[Coal] Stopped")
                    }
                } else if running {
                    currentMW = targetMW
                }

			case <-ctx.Done(): 
				return
            }
        }
    }()
}

func (c *CoalPlant) Stop() {
    close(c.stopChan)
}