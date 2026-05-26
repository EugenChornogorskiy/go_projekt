package main

import (
	"fmt"
	"sort"
	"time"
	"context"
)

type GridHub struct {
	demandChan    chan DemandReport
	renewableChan <-chan RenewableUpdate
	forecastChan  <-chan ForecastReport
	coalCmdChan   chan<- CoalCommand
	coalStatusReq chan chan CoalStatus
	ESScommandChan  chan ESSCommand
	logger       chan interface{}
	consumers    map[string]DemandReport
	weatherChan <-chan WeatherData
	stopChan     chan struct{} 
}

func NewGridHub(
	demandChan chan DemandReport,
	renewableChan <-chan RenewableUpdate,
	forecastChan <-chan ForecastReport,
	coalCmdChan chan<- CoalCommand,
	coalStatusReq chan chan CoalStatus,
	ess  chan ESSCommand,
	logger chan interface{},
	weatherChan <-chan WeatherData,
) *GridHub {
	return &GridHub{
		demandChan:    demandChan,
		renewableChan: renewableChan,
		forecastChan:  forecastChan,
		coalCmdChan:   coalCmdChan,
		coalStatusReq: coalStatusReq,
		ESScommandChan: ess,
		logger:        logger,
		consumers:     make(map[string]DemandReport),
		weatherChan:   weatherChan,
		stopChan:      make(chan struct{}),
	}
}

func (g *GridHub) Run(ctx context.Context) {
	ticker := time.NewTicker(GridStep)
	renewableMW := 0.0
	lastWeatherWind := 0.0
	lastWeatherSolar := 0.0

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				g.balanceGrid(renewableMW, lastWeatherWind, lastWeatherSolar)
			case weather:= <-g.weatherChan:
				lastWeatherWind = weather.WindSpeed
				lastWeatherSolar = weather.SolarRadiation
			case update := <-g.renewableChan:
				renewableMW = update.MW
			case forecast := <-g.forecastChan:
				g.handleForecast(forecast)
			case req := <-g.demandChan:
				g.consumers[req.ID] = req 
			case <-g.stopChan:
				return
			}
		}
	}()
}

func (g *GridHub) balanceGrid(renewableMW float64, windSpeed float64, solarRad float64) {
	totalDemand := 0.0
	var consumerList []DemandReport
	for _, c := range g.consumers {
		totalDemand += c.DemandMW
		consumerList = append(consumerList, c)
	}

	available := renewableMW
 
	statusCh := make(chan CoalStatus, 1)
	g.coalStatusReq <- statusCh
	coalStatus := <-statusCh
	coalMW := 0.0
	if coalStatus.Running {
		coalMW = coalStatus.Mw
		available += coalStatus.Mw
	}

	if available < totalDemand {
		resp := make(chan float64, 1)
		g.ESScommandChan <- ESSCommand{Type: "discharge", Amount: totalDemand - available, Response: resp}
		discharge := <-resp
		available += discharge
	} else if available > totalDemand {
		excess := available - totalDemand
		resp := make(chan float64, 1)
		g.ESScommandChan <- ESSCommand{Type: "charge", Amount: excess, Response: resp}
		charged := <-resp
		available -= charged
	}
	finalBalance := available - totalDemand
	resp := make(chan float64, 1)
	g.ESScommandChan <- ESSCommand{Type: "status", Amount: 0, Response: resp}
	essSoc := <-resp 

	log := (map[string]interface{}{
		"Event":          "grid_state", 
		"Timestamp":      time.Now(),
		"RenewableMW":    renewableMW,
		"CoalMW":         coalMW,
		"TotalProduction": available,
		"TotalDemand":    totalDemand,
		"InitialBalance": available - totalDemand,
		"FinalBalance":   finalBalance,
		"Baterie":   	  essSoc*100,  
		"CoalRunning":    coalStatus.Running,
	})
	g.logger <- log
	if available < totalDemand {
		sort.Slice(consumerList, func(i, j int) bool {
			return consumerList[i].Priority < consumerList[j].Priority
		})
		remaining := available
		for _, c := range consumerList {
			if remaining >= c.DemandMW {
				c.ResponseCh <- SupplyStatus{AllocatedMW: c.DemandMW, Reason: "OK"}
				remaining -= c.DemandMW
			} else {
				c.ResponseCh <- SupplyStatus{AllocatedMW: 0, Reason: "LoadShed"}
				log := (map[string]interface{}{
					"event":    "load_shed",
					"consumer": c.ID,
					"time":     time.Now(),
				})
				g.logger <- log
			}
		}
	} else {
		for _, c := range consumerList {
			c.ResponseCh <- SupplyStatus{AllocatedMW: c.DemandMW, Reason: "OK"}
		}
	}
	status := "STABLE"
	if available < totalDemand*0.9 {
		status = "CRITICAL"
	}
	fmt.Println("==============================")
	fmt.Printf("[Pogoda] Wiatr: %.1f km/h | Słońce: %.0f%%\n", windSpeed, solarRad*100)
	fmt.Printf("[Produkcja] OZE: %.1f MW | Konwencjonalna: %.1f MW | Baterie: %.0f%% (SoC)\n",
		renewableMW, coalMW, essSoc*100)
	fmt.Printf("[Sieć] Popyt: %.1f MW | Bilans: %+.1f MW | Stan: [%s]\n",
		totalDemand, finalBalance, status)
	fmt.Println("==============================")
}

func (g *GridHub) handleForecast(forecast ForecastReport) {
	resp := make(chan bool, 1)
	if forecast.Trend < -0.1 {   
		select {
			case g.coalCmdChan <- CoalCommand{Start: true, Response: resp}:
				<-resp
			default:
				
	 	}
	} else if forecast.Trend > 0.1 {    
		select {
			case g.coalCmdChan <- CoalCommand{Start: false, Response: resp}:
				<-resp
			default:
				
	 	}
	}
}

func (g *GridHub) Stop() {
	close(g.stopChan)
}