package power

import (
	"github.com/appleboy/easyssh-proxy"
	"github.com/sajari/regression"
)

type FormulaProvider struct {
	Formula      formula
	HasFormula   bool
	FormulaSlice [][]string
	SSHClient    *easyssh.MakeConfig
	PowerChan    []float64
}

type formula struct {
	Start      chan [][]string
	Alpha      float64
	Beta       float64
	gamma      float64
	delta      float64
	Intercept  float64
	Regression regression.Regression
}
