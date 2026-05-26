package main

import (
	"time"
	"context"
)

type RenewableManager struct {
	weatherChan <-chan WeatherData
	updateChan  chan<- RenewableUpdate
	stopChan    chan struct{}
}

func NewRenewableManager(weatherChan <-chan WeatherData, updateChan chan<- RenewableUpdate) *RenewableManager {
	return &RenewableManager{
		weatherChan: weatherChan,
		updateChan:  updateChan,
		stopChan:    make(chan struct{}),
	}
}

func (r *RenewableManager) Run(ctx context.Context) {
	ticker := time.NewTicker(WeatherStep)
	var lastWind, lastSolar float64

	go func() {
		defer ticker.Stop()
		for {
			select {
			case wd := <-r.weatherChan:
				lastWind = wd.WindSpeed
				lastSolar = wd.SolarRadiation
			case <-ticker.C:
				mw := (lastWind/50.0)*20 + lastSolar*10
				select {
				case r.updateChan <- RenewableUpdate{MW: mw}:
				default:
				}
			case <-ctx.Done(): 
				return 
			}
		}
	}()
} 