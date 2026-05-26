package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
  
	forecastChan := make(chan ForecastReport, 1)
	demandChan := make(chan DemandReport, 100)
	renewableUpdateChan := make(chan RenewableUpdate, 10)
	coalCmdChan := make(chan CoalCommand, 10)
	coalStatusReq := make(chan chan CoalStatus, 10)
 	ESSCommand := make(chan ESSCommand, 10)
	logChan := make(chan interface{}, 100)
	
	logger := NewFileLogger("logs/grid.json",logChan) 
	go logger.Run(ctx)

	weatherStation := NewWeatherStation()
	go weatherStation.Run(ctx)  
 
	predictor := NewPredictor(weatherStation.Subscribe(), forecastChan)
	go predictor.Run(ctx)
 
	coal := NewCoalPlant(coalCmdChan,coalStatusReq)
	go coal.Run(ctx)
 
	ess := NewESS(5.0, 20.0,ESSCommand)
	go ess.Run(ctx)

	hub := NewGridHub(demandChan, renewableUpdateChan, forecastChan, coalCmdChan, coalStatusReq, ESSCommand, logChan,weatherStation.Subscribe())
	go hub.Run(ctx)
 
	renewableMgr := NewRenewableManager(weatherStation.Subscribe(), renewableUpdateChan)
	go renewableMgr.Run(ctx)
 
	consumers := []*BaseConsumer{
		NewConsumer("Residential1", 3, demandChan),
		NewConsumer("Industrial1", 2, demandChan),
		NewConsumer("Critical1", 1, demandChan),
	}
	for _, c := range consumers {
		go c.Run(ctx)
	}
 
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	cons := NewConsumer("Critical2", 1, demandChan)
	go cons.Run(ctx)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")  
		hub.Stop()
		time.Sleep(100 * time.Millisecond) 
		cancel()
		os.Exit(0)
	}()

	<-ctx.Done()
}