package main

import "time"

const (
	WeatherStep        = 5 * time.Millisecond
	GridStep           = 60 * time.Millisecond
	WeatherPerGrid     = int(GridStep / WeatherStep)
	ForecastHorizon    = 5  
	PredictorBufferSize = WeatherPerGrid 
)

type DemandReport struct {
	ID         string
	DemandMW   float64
	Priority   int 
	ResponseCh chan SupplyStatus
}
type CoalCommand struct {
	Start    bool 
	Response chan bool 
}
type CoalStatus struct {
	Running bool
	Mw      float64
}
type CoalStatusRequest struct {
	ResponseChan chan CoalStatus
}
type SupplyStatus struct {
	AllocatedMW  float64
	Reason       string 
}

type ForecastReport struct {
	Timestamp        time.Time
	Trend            float64
	PredictedMWChange float64 
}

type WeatherData struct {
	WindSpeed     float64
	SolarRadiation float64
	Timestamp     time.Time
}

type ESSStatus struct {
	SoC        float64
	ChargeMW   float64
	DischargeMW float64
}
type RenewableUpdate struct {
	MW float64
}
type ESSCommand struct {
	Type     string 
	Amount   float64
	Response chan float64 
}
