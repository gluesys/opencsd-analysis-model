package simulator

import (
	"encoding/json"
	"errors"
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

func getTableSchema(tableName string) TableSchema {
	// TODO 스키마 데이터 로드하는 형식으로 바꿔야함 (queryEngine or ddl)
	schema := make(map[string]TableSchema)
	schema["employees"] = TableSchema{
		ColumnNames: []string{"emp_no", "birth_date", "first_name", "last_name", "gender", "hire_date"},
		ColumnTypes: []string{"int", "date", "char", "char", "char", "date"}, // int, char, varchar, TEXT, DATETIME,  ...
		ColumnSizes: []int{8, -1, 30, 30, 1, -1},                             // Data Size}
	}
	schema["nation"] = TableSchema{
		ColumnNames: []string{"N_NATIONKEY", "N_NAME", "N_REGIONKEY", "N_COMMENT"},
		ColumnTypes: []string{"int", "char", "int", "char"}, // int, char, varchar, TEXT, DATETIME,  ...
		ColumnSizes: []int{8, 25, 8, 152},                   // Data Size}
	}
	schema["region"] = TableSchema{
		ColumnNames: []string{"R_REGIONKEY", "R_NAME", "R_COMMENT"},
		ColumnTypes: []string{"int", "char", "varchar"}, // int, char, varchar, TEXT, DATETIME,  ...
		ColumnSizes: []int{8, 8, 152},                   // Data Size}
	}
	schema["part"] = TableSchema{
		ColumnNames: []string{"P_PARTKEY", "P_NAME", "P_MFGR", "P_BRAND", "P_TYPE", "P_SIZE", "P_CONTAINER", "P_RETAILPRICE", "P_COMMENT"},
		ColumnTypes: []string{"int", "varchar", "char", "char", "varchar", "int", "char", "decimal(15,2)", "varchar"}, // int, char, varchar, TEXT, DATETIME,  ...
		ColumnSizes: []int{8, 55, 25, 10, 25, 8, 10, 15, 101},                                                         // Data Size}
	}
	schema["supplier"] = TableSchema{
		ColumnNames: []string{"S_SUPPKEY", "S_NAME", "S_ADDRESS", "S_NATIONKEY", "S_PHONE", "S_ACCTBAL", "S_COMMENT"},
		ColumnTypes: []string{"int", "char", "varchar", "int", "char", "decimal(15,2)", "varchar"}, // int, char, varchar, TEXT, DATETIME,  ...
		ColumnSizes: []int{8, 25, 40, 8, 15, 15, 101},                                              // Data Size}
	}
	schema["partsupp"] = TableSchema{
		ColumnNames: []string{"PS_PARTKEY", "PS_SUPPKEY", "PS_AVAILQTY", "PS_SUPPLYCOST", "PS_COMMENT"},
		ColumnTypes: []string{"int", "varchar", "varchar", "int", "char", "decimal(15,2)", "char", "varchar"}, // int, char, varchar, TEXT, DATETIME,  ...
		//ColumnSizes: []int{110325, 110325, 110325, 110325},  // Data Size}
		ColumnSizes: []int{8, 25, 40, 8, 15, 15, 10, 117}, // Data Size}
	}
	schema["customer"] = TableSchema{
		ColumnNames: []string{"C_CUSTKEY", "C_NAME", "C_ADDRESS", "C_NATIONKEY", "C_PHONE", "C_ACCTBAL", "C_MKTSEGMENT", "C_COMMENT"},
		ColumnTypes: []string{"int", "varchar", "varchar", "int", "char", "decimal(15,2)", "char", "varchar"}, // int, char, varchar, TEXT, DATETIME,  ...
		ColumnSizes: []int{8, 25, 40, 8, 15, 15, 10, 117},                                                     // Data Size}
	}
	schema["orders"] = TableSchema{
		ColumnNames: []string{"O_ORDERKEY", "O_CUSTKEY", "O_ORDERSTATUS", "O_TOTALPRICE", "O_ORDERDATE", "O_ORDERPRIORITY", "O_CLERK", "O_SHIPPRIORITY", "O_COMMENT"},
		ColumnTypes: []string{"int", "int", "char", "decimal(15,2)", "date", "char", "char", "int", "varchar"}, // int, char, varchar, TEXT, DATETIME,  ...
		ColumnSizes: []int{8, 8, 1, 15, -1, 15, 15, 8, 79},                                                     // Data Size}
	}
	schema["lineitem"] = TableSchema{
		ColumnNames: []string{"L_ORDERKEY", "L_PARTKEY", "L_SUPPKEY", "L_LINENUMBER", "L_QUANTITY", "L_EXTENDEDPRICE", "L_DISCOUNT", "L_TAX", "L_RETURNFLAG", "L_LINESTATUS", "L_SHIPDATE", "L_COMMITDATE", "L_RECEIPTDATE", "L_SHIPINSTRUCT", "L_SHIPMODE", "L_COMMENT"},
		ColumnTypes: []string{"int", "int", "int", "int", "decimal(15,2)", "decimal(15,2)", "decimal(15,2)", "decimal(15,2)", "char", "char", "date", "date", "date", "char", "char", "varchar"}, // int, char, varchar, TEXT, DATETIME,  ...
		ColumnSizes: []int{8, 8, 8, 8, 15, 15, 15, 15, 1, 1, -1, -1, 25, 10, 44},                                                                                                                 // Data Size}
	}

	return schema[tableName]
}

func Parse(query string) (ParsedQuery, error) {
	// whereSlice := strings.Split(query, "WHERE")
	// whereSlice = strings.Split(query, "")
	querySlice := strings.Split(query, " ")
	parsedQuery := ParsedQuery{
		TableName:    "",
		Columns:      make([]Select, 0),
		WhereClauses: make([]Where, 0),
	}
	index := 0
	whereSlice := make([]string, 3)
	operatorFlag := false
	selectAllFlag := false

	flag := 0
	for _, atom := range querySlice {
		if strings.ToLower(atom) == "select" {
			//klog.Infoln("First Element select")
			//klog.Infoln("Current Index", index)
			continue
		} else if strings.ToLower(atom) == "from" {
			//klog.Infoln("Second Element from")
			index++
			//klog.Infoln("Current Index", index)
			continue
		} else if strings.ToLower(atom) == "where" {
			//klog.Infoln("Third Element from")
			index++
			flag = 1
			//oln("Current Index", index)
			continue
		} else if strings.ToLower(atom) == "and" && flag == 1 {
			continue
		}
		// log.Println(index)
		switch index {
		case 0: // select뒤에 나오는 인자를 파싱
			if atom == "*" {
				// nothing.
				// 모든 데이터를 의미함
				selectAllFlag = true
			} else if ok, aggregateName := isAggregateFunc(atom); ok {
				// 집계함수인 경우
				temp := strings.TrimPrefix(atom, aggregateName+"(")
				aggregateValue := strings.TrimSuffix(temp, ")")
				col := Select{
					ColumnType:     2,
					ColumnName:     "",
					AggregateName:  aggregateName,
					AggregateValue: aggregateValue,
				}
				parsedQuery.Columns = append(parsedQuery.Columns, col)
			} else {
				// 컬럼명인 경우
				columnName := strings.TrimSuffix(atom, ",")

				col := Select{
					ColumnType:     1,
					ColumnName:     columnName,
					AggregateName:  "",
					AggregateValue: "",
				}
				parsedQuery.Columns = append(parsedQuery.Columns, col)
			}
		case 1:
			parsedQuery.TableName = atom
		case 2:
			if operatorFlag {
				if ok, operator := isOperator(atom); ok {
					parsedQuery.WhereClauses[len(parsedQuery.WhereClauses)-1].Operator = operator
					operatorFlag = false
				} else {
					return ParsedQuery{}, errors.New("Invaild Query")
				}
			} else {
				if ok, exp := isEXP(atom); ok {
					whereSlice = strings.Split(atom, exp)
					w := Where{
						LeftValue:  whereSlice[0],
						Exp:        exp,
						RightValue: whereSlice[1],
						Operator:   "NULL",
					}
					parsedQuery.WhereClauses = append(parsedQuery.WhereClauses, w)
					// operatorFlag = true

				} else {
					return ParsedQuery{}, errors.New("Invaild Query")
				}
			}

		}
	}

	if selectAllFlag {
		schema := getTableSchema(parsedQuery.TableName)
		for _, columnName := range schema.ColumnNames {
			col := Select{
				ColumnType:     1,
				ColumnName:     columnName,
				AggregateName:  "",
				AggregateValue: "",
			}
			parsedQuery.Columns = append(parsedQuery.Columns, col)
		}

	}
	//klog.Infoln(*request)
	return parsedQuery, nil
}

func main() {
	log.SetFlags(log.Lshortfile)
}
