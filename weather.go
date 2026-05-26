package main

import (
	"math/rand"
	"time"
	"context"
)

type WeatherStation struct {
	subscribers []chan WeatherData
	stopChan    chan struct{}
}

func NewWeatherStation() *WeatherStation {
	return &WeatherStation{
		subscribers: make([]chan WeatherData, 0),
		stopChan:    make(chan struct{}),
	}
}

func (w *WeatherStation) Subscribe() <-chan WeatherData {
	ch := make(chan WeatherData, 2) 
	w.subscribers = append(w.subscribers, ch)
	return ch
}

func (w *WeatherStation) Run(ctx context.Context) {
	ticker := time.NewTicker(WeatherStep)
	go func() {
		wind := 10.0
		solar := 0.5
		for {
			select {
			case <-ticker.C: 
				wind += (rand.Float64() - 0.5) * 2.0
				if wind < 0 {
					wind = 0
				}
				if wind > 50 {
					wind = 50
				}
				solar += (rand.Float64() - 0.5) * 0.1
				if solar < 0 {
					solar = 0
				}
				if solar > 1 {
					solar = 1
				}
				data := WeatherData{
					WindSpeed:     wind,
					SolarRadiation: solar,
					Timestamp:     time.Now(),
				}

				for _, ch := range w.subscribers {
					select {
					case ch <- data:
					default:
					}
				}
			case <-ctx.Done():  
				ticker.Stop()
				return
			}
		}
	}()
} 