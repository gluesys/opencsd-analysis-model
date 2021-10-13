package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"analysis-model/pkg/analysis"

	"github.com/julienschmidt/httprouter"
)

var flag = 1
var ans analysis.Analysis

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func StartMeasure(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Println("Measure Start")
	// flagChan := make(chan int)
	// avgChan := make(chan float64)
	cpuChan := make(chan float64)
	memChan := make(chan float64)
	var cpuList []float64
	var memList []float64

	// analysis.GetCPU(flagChan, avgChan)
	// go analysis.GetCPU(cpuChan)
	// go analysis.GetMem()
	// log.Println(ans)

	for {
		if flag == 0 {
			break
		}
		go analysis.GetCPU(cpuChan)
		go analysis.GetMem(memChan)
		cpuList = append(cpuList, <-cpuChan)
		memList = append(memList, <-memChan)
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
	// analysis.GetMemory()
	log.Println("CPU Usage", cpuAvg)
	log.Println("MEM Usage", memAvg)

	predict := 96.2107 + (cpuAvg * -(0.4059)) + (memAvg * (-17.2624))
	log.Println("POWER Usage", predict)

	measure := analysis.Analysis{
		Cpu:    cpuAvg,
		Memory: memAvg,
		Energy: predict,
	}
	log.Println(measure)
	ans = measure
	jsonString, err := json.Marshal(measure)
	if err != nil {
		log.Println(err)
	}

	w.Write(jsonString)
}

func EndMeasure(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Println("Measure End")
	// flagChan := make(chan int)
	// analysis.SetFlag(flagChan)
	// flagChan <- 1
	// analysis.SetFlag(flagChan)
	flag = 0
	// log.Println(ans)
	// jsonString, err := json.Marshal(ans)
	// if err != nil {
	// 	log.Println(err)
	// }

	// w.Write(jsonString)
}

func Run() {
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/start/measure", StartMeasure)
	router.GET("/end/measure", EndMeasure)

	log.Fatal(http.ListenAndServe(":50500", router))
}
