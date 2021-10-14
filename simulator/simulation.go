package simulator

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"simulator/analysis"
	types "simulator/type"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type Snippet struct {
	ParsedQuery   ParsedQuery `json:"parsedQuery"`
	TableSchema   TableSchema `json:"tableSchema"`
	BlockOffset   int         `json:"blockOffset"`
	BufferAddress string      `json:"bufferAddress"`
}
type ParsedQuery struct {
	TableName    string   `json:"tableName"`
	Columns      []Select `json:"columnName"`
	WhereClauses []Where  `json:"whereClause"`
}
type Select struct {
	ColumnType     int    `json:"columnType"` // 1: (columnName), 2: (aggregateName,aggregateValue)
	ColumnName     string `json:"columnName"`
	AggregateName  string `json:"aggregateName"`
	AggregateValue string `json:"aggregateValue"`
}

type Where struct {
	LeftValue  string `json:"leftValue"`
	Exp        string `json:"exp"`
	RightValue string `json:"rightValue"`
	Operator   string `json:"operator"` // "AND": 뒤에 나오는 Where은 And조건, "OR": 뒤에 나오는 Where은 OR 조건, "NULL": 뒤에 나오는 조건 없음
}
type TableSchema struct {
	ColumnNames []string `json:"columnNames"`
	ColumnTypes []string `json:"columnTypes"` // int, char, varchar, TEXT, DATETIME,  ...
	ColumnSizes []int    `json:"columnSizes"` // Data Size
}

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    Data   `json:"data"`
}
type Data struct {
	Table  string              `json:"table"`
	Field  []string            `json:"field"`
	Values []map[string]string `json:"values"`
}

type Analysis struct {
	Cpu    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Energy float64 `json:"energy"`
}

type ScanData struct {
	Snippet   types.Snippet       `json:"snippet"`
	Tabledata map[string][]string `json:"tabledata"`
}
type FilterData struct {
	Result   types.QueryResponse `json:"result"`
	TempData map[string][]string `json:"tempData"`
}
type ResponseA struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

var flag = 1
var ans analysis.Analysis

const rootDirectory = "/root/workspace/usr/coyg/module/tpch/"

func resJsonParser(jsonDataString string) Response {
	var res Response

	if err := json.Unmarshal([]byte(jsonDataString), &res); err != nil {
		log.Fatal(err)
	}

	return res
}

func isAggregateFunc(atom string) (bool, string) {
	aggregateList := []string{"count", "sum", "avg", "max", "min"}
	for _, aggregater := range aggregateList {
		if strings.Contains(atom, aggregater+"(") {
			return true, aggregater
		}
	}
	return false, ""
}
func isOperator(atom string) (bool, string) {
	opList := []string{"and", "or"}
	for _, op := range opList {
		atom = strings.ToLower(atom)
		if atom == op {
			return true, op
		}
	}
	return false, ""
}
func isEXP(atom string) (bool, string) {
	expList := []string{">=", "<=", ">", "<", "="}
	for _, exp := range expList {
		if strings.Contains(atom, exp) {
			return true, exp
		}
	}
	return false, ""
}
func printClient(res Response) {
	if res.Code == 200 {

		datas := [][]string{}

		for _, value := range res.Data.Values {
			data := []string{}
			for _, field := range res.Data.Field {
				data = append(data, string(value[field]))
			}

			datas = append(datas, data)

		}

		fmt.Println()

		table := tablewriter.NewWriter(os.Stdout)

		table.SetHeader(res.Data.Field)
		table.SetBorder(true)
		table.SetAutoFormatHeaders(false)
		table.SetCaption(true, "Total: "+strconv.Itoa(len(datas)))
		table.AppendBulk(datas)
		table.Render()

	} else {
		log.Fatal(res.Message)
	}
}

func main() {
	log.SetFlags(log.Lshortfile)
}
