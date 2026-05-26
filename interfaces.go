package main


type EnergySource interface { 
	Run() error 
	stop()
}

type Predictor interface {
	Start()
	Stop()
	GetForecast() ForecastReport
}

type Consumer interface {
	GetDemand() float64
	GetPriority() int
	GetID() string
	UpdateStatus(status SupplyStatus)
}
type EnergyStorage interface {
	Charge(amount float64) float64
	Discharge(amount float64) float64
	GetSoC() float64
	GetCapacityMW() float64
}

type WeatherProvider interface {
	Start()
	Subscribe() <-chan WeatherData
}

type DataLogger interface {
	Log(entry interface{})
	Flush()
	Stop()
}