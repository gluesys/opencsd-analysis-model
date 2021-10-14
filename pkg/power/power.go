package power

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func NewFormula() *FormulaProvider {
	fp := &FormulaProvider{
		Formula:      formula{},
		HasFormula:   false,
		FormulaSlice: make([][]string, 0),
		PowerChan:    make([]float64, 0),
	}

	return fp
}

func (fp *FormulaProvider) Regression(start [][]string) (a float64, b float64, intercept float64) {
	records := make([][]float64, 0, len(start))
	powerslice := make([]float64, 0, len(start))
	cpuslice := make([]float64, 0, len(start))
	memslice := make([]float64, 0, len(start))
}

func (f *formula) getCoefficient(formula string) (err error) {
	temp := strings.Split(formula, " = ")
	spstring := strings.Split(temp[1], " + ")
	f.Intercept, err = strconv.ParseFloat(spstring[0], 64)
	if err != nil {
		return err
	}
	spstring = spstring[1:]
	for _, atom := range spstring {
		temp := strings.Split(atom, "*")
		metric := temp[0]
		strValue := temp[1]
		switch metric {
		case "Cpu":
			f.Alpha, err = strconv.ParseFloat(strValue, 64)
			if err != nil {
				return err
			}
			log.Println(f.Alpha)

		case "Memory":
			f.Beta, err = strconv.ParseFloat(strValue, 64)
			if err != nil {
				return err
			}
			log.Println(f.Beta)
		}

	}

	return nil

}

func (fp *FormulaProvider) GetPower(powerchan chan float64) {

	cmd := exec.Command("turbostat", "--Summary", "-i", "1", "-n", "1", "-s", "PkgWatt")

	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}

	slice := strings.Split(string(out), "PkgWatt\n")

	var a string
	for _, str := range slice {
		a = str
	}

	var b string
	b = a[0:5]
	// secondslice := strings.Split(a, " ")

	s, err := strconv.ParseFloat(b, 64)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(s)

	powerchan <- s

}
