package main

import (
	"math/rand"
	"time"
	"context"
)

type BaseConsumer struct {
	id         string
	priority   int
	demandChan chan<- DemandReport
	stopChan   chan struct{}
}

func NewConsumer(id string, priority int, demandChan chan<- DemandReport) *BaseConsumer {
	return &BaseConsumer{
		id:         id,
		priority:   priority,
		demandChan: demandChan,
		stopChan:   make(chan struct{}),
	}
}

func (c *BaseConsumer) Run(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(GridStep)
		hour := 0
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				var demand float64
				switch c.priority {
				case 3:
					if (hour%24 >= 7 && hour%24 <= 9) || (hour%24 >= 18 && hour%24 <= 22) {
						demand = 2.0 + rand.Float64()*1.0
					} else {
						demand = 0.8 + rand.Float64()*0.5
					}
				case 2:
					if hour%24 >= 6 && hour%24 <= 18 {
						demand = 8.0 + rand.Float64()*4.0
					} else {
						demand = 1.5 + rand.Float64()*0.5
					}
				default:
					demand = 3.0 + rand.Float64()*0.5
				}

				responseCh := make(chan SupplyStatus, 1)
				req := DemandReport{
					ID:         c.id,
					DemandMW:   demand,
					Priority:   c.priority,
					ResponseCh: responseCh,
				}

				select {
				case c.demandChan <- req:
				case <-c.stopChan:
					return
				}

				select {
				case <-responseCh:
				case <-time.After(GridStep / 2):
				case <-c.stopChan:
					return
				}
				hour++
			case <-ctx.Done(): 
				return
			}
		}
	}()
}

func (c *BaseConsumer) Stop() {
	close(c.stopChan)
}

func (c *BaseConsumer) GetID() string {
	return c.id
}