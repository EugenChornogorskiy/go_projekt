package main

import ( 
	"time"
	"context"
) 

type ESSStatusRequest struct {
	Response chan ESSStatus
}

type ESS struct {
	capacityMW   float64
	totalEnergy  float64
	maxEnergy    float64
	commandChan  chan ESSCommand
	statusChan   chan chan ESSStatus
	stopChan     chan struct{}
}

func NewESS(capacityMW, maxEnergyMWh float64,commandChan  chan ESSCommand) *ESS {
	return &ESS{
		capacityMW:  capacityMW,
		totalEnergy: maxEnergyMWh * 0.5,
		maxEnergy:   maxEnergyMWh,
		commandChan: commandChan,
		statusChan:  make(chan chan ESSStatus, 10),
		stopChan:    make(chan struct{}),
	}
}

func (e *ESS) Run(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(GridStep)
		defer ticker.Stop()

		for {
			select {
			case cmd := <-e.commandChan:
				switch cmd.Type {
				case "charge":
					charged := e.charge(cmd.Amount)
					if cmd.Response != nil {
						cmd.Response <- charged
					}
				case "discharge":
					discharged := e.discharge(cmd.Amount)
					if cmd.Response != nil {
						cmd.Response <- discharged
					}
				case "status":
					if cmd.Response != nil {
						cmd.Response <- e.totalEnergy / e.maxEnergy
					}
				}

			case respCh := <-e.statusChan:
				respCh <- ESSStatus{
					SoC:         e.totalEnergy / e.maxEnergy,
					ChargeMW:    e.capacityMW,
					DischargeMW: e.capacityMW,
				} 
			case <-e.stopChan:
				return
			}
		}
	}()
}

func (e *ESS) charge(amount float64) float64 {
	maxPossible := e.capacityMW
	if amount > maxPossible {
		amount = maxPossible
	}
	energyToAdd := amount
	if e.totalEnergy+energyToAdd > e.maxEnergy {
		energyToAdd = e.maxEnergy - e.totalEnergy
	}
	e.totalEnergy += energyToAdd
	return energyToAdd
}

func (e *ESS) discharge(amount float64) float64 {
	maxPossible := e.capacityMW
	if amount > maxPossible {
		amount = maxPossible
	}
	energyToRemove := amount
	if e.totalEnergy-energyToRemove < 0 {
		energyToRemove = e.totalEnergy
	}
	e.totalEnergy -= energyToRemove
	return energyToRemove
} 
 

func (e *ESS) Stop() {
	close(e.stopChan)
}