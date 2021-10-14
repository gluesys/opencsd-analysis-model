package power

import (
	"fmt"
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
