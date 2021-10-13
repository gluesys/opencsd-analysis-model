package analysis

import (
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mackerelio/go-osstat/memory"
)

type Analysis struct {
	Cpu    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Energy float64 `json:"energy"`
}

// var flag = 1

// func SetFlag(flagChan chan int) {
// 	flag = <-flagChan
// }

func cpuMeasure() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			return
		}
	}
	return
}

var ssdCpu, csdCpu = 0, 0

func GetCPU(cpuChan chan float64) {
	// var cpuList []float64
	// for {
	idle0, total0 := cpuMeasure()
	time.Sleep(1 * time.Second)
	idle1, total1 := cpuMeasure()
	idleTicks := float64(idle1 - idle0)
	totalTicks := float64(total1 - total0)
	cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks

	cpuBenefit := csdCpu / ssdCpu
	log.Printf("CPU usage is %f%% [busy: %f, total: %f]\n", cpuUsage, totalTicks-idleTicks, totalTicks)
	cpuChan <- cpuUsage

	fmt.Println(cpuBenefit)
	// 	if flag == 0 {
	// 		break
	// 	}
	// }
	// total := 0.0
	// for _, cpu := range cpuList {
	// 	total += cpu
	// }
	// avg := total / float64(len(cpuList))

	// avgChan <- avg
}

func GetMem(memChan chan float64) {
	mem, err := memory.Get()
	if err != nil {
		log.Println(err)
	}
	log.Println(mem.Total)
	log.Println(mem.Used)
	log.Println(mem.Cached)
	log.Println(mem.Free)
	memUsage := 100 * (float64(mem.Used) / float64(mem.Total))
	log.Println(memUsage)

	memChan <- memUsage
}

func GetMemory() {
	PrintMemUsage()

	var overall [][]int
	for i := 0; i < 4; i++ {

		a := make([]int, 0, 999999)
		overall = append(overall, a)

		PrintMemUsage()
		time.Sleep(time.Second)
	}

	// Clear our memory and print usage, unless the GC has run 'Alloc' will remain the same
	overall = nil
	PrintMemUsage()

	// Force GC to clear up, should see a memory drop
	runtime.GC()
	PrintMemUsage()
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
