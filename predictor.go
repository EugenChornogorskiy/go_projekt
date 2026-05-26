package main

import (
	"time"
	"context"
)

type PredictorImpl struct {
	weatherChan <-chan WeatherData
	forecastChan chan<- ForecastReport
	buffer      []WeatherData
	stopChan    chan struct{}
}

func NewPredictor(weatherChan <-chan WeatherData, forecastChan chan<- ForecastReport) *PredictorImpl {
	return &PredictorImpl{
		weatherChan:  weatherChan,
		forecastChan: forecastChan,
		buffer:       make([]WeatherData, 0, PredictorBufferSize),
		stopChan:     make(chan struct{}),
	}
}

func (p *PredictorImpl) Run(ctx context.Context) {
	weatherTicker := time.NewTicker(WeatherStep)
	gridTicker := time.NewTicker(GridStep)
	
	go func() {
		for {
			select {
			case <-weatherTicker.C:
				select {
				case wd := <-p.weatherChan:
					p.buffer = append(p.buffer, wd)
					if len(p.buffer) > PredictorBufferSize {
						p.buffer = p.buffer[1:]
					}
				default:
				}
				
			case <-gridTicker.C:
				if len(p.buffer) < 2 {
					continue
				}

				first := p.buffer[0]
				last := p.buffer[len(p.buffer)-1]
				deltaWind := last.WindSpeed - first.WindSpeed
				deltaSolar := last.SolarRadiation - first.SolarRadiation

				trend := (deltaWind/10.0 + deltaSolar) / 2.0
				predictedMWChange := trend * 15.0 * 2.0
				forecast := ForecastReport{
					Timestamp:         time.Now(),
					Trend:             trend,
					PredictedMWChange: predictedMWChange,
				}
				select {
				case p.forecastChan <- forecast:
				default:

				}
			case <-ctx.Done():  
				weatherTicker.Stop()
				gridTicker.Stop()
				return
			}
		}
	}()
} 