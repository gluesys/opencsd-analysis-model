package rest

import (
	"encoding/json"
	"log"
	"net/http"

	"analysis-model/pkg/analysis"
	"analysis-model/pkg/power"

	"github.com/julienschmidt/httprouter"
)

var flag = 1
var onCSD = 0

type Metrics struct {
	CPU    []string
	Memory []string
	Power  []string
}

func StartMeasure(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Println("Measure Start Request")
	cpuChan := make(chan float64)
	memChan := make(chan float64)
	powerChan := make(chan float64)
	var cpuList []float64
	var memList []float64
	var powerList []float64
	var powerAvg float64

	if onCSD != 0 {
		fp := power.NewFormula()
		for {
			if flag == 0 {
				break
			}
			go analysis.GetCPU(cpuChan)
			go analysis.GetMem(memChan)
			go fp.GetPower(powerChan)
			cpuList = append(cpuList, <-cpuChan)
			memList = append(memList, <-memChan)
			powerList = append(powerList, <-powerChan)
		}
	} else {
		for {
			if flag == 0 {
				break
			}
			go analysis.GetCPU(cpuChan)
			go analysis.GetMem(memChan)
			cpuList = append(cpuList, <-cpuChan)
			memList = append(memList, <-memChan)

		}
	}

	cpuTotal := 0.0
	for _, cpu := range cpuList {
		cpuTotal = cpuTotal + cpu
	}
	cpuAvg := cpuTotal / float64(len(cpuList))
	memTotal := 0.0
	for _, mem := range memList {
		memTotal = memTotal + mem
	}
	memAvg := memTotal / float64(len(memList))

	if onCSD == 0 {
		powerTotal := 0.0
		for _, power := range powerList {
			powerTotal = powerTotal + power
		}
		powerAvg = powerTotal / float64(len(powerList))
	}

	log.Println("CPU Usage", cpuAvg)
	log.Println("MEM Usage", memAvg)
	if onCSD == 0 {
		log.Println("POWER Usage", powerAvg)
	}

	predict := 96.2107 + (cpuAvg * -(0.4059)) + (memAvg * (-17.2624))
	log.Println("POWER Usage", predict)

	measure := analysis.Analysis{
		Cpu:    cpuAvg,
		Memory: memAvg,
		Energy: predict,
	}
	log.Println(measure)
	jsonString, err := json.Marshal(measure)
	if err != nil {
		log.Println(err)
	}

	w.Write(jsonString)
}

func EndMeasure(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Println("Measure End Request")

	flag = 0

}

func Run() {
	router := httprouter.New()
	router.GET("/start/measure", StartMeasure)
	router.GET("/end/measure", EndMeasure)

	log.Fatal(http.ListenAndServe(":50500", router))
}
