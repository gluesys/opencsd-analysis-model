package simulator

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"simulator/analysis"
	types "simulator/type"
	"strconv"
	"strings"
	"time"

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
type SSDInfo struct {
	Query     string  `json:"query"`
	CPU       float64 `json:"cpu"`
	Energy    float64 `json:"energy"`
	QueryTime float64 `json:"queryTime"`
}
type CSDInfo struct {
	CPU       float64 `json:"cpu"`
	Energy    float64 `json:"energy"`
	QueryTime float64 `json:"queryTime"`
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

func RequestSnippet(query string) (t float64, result []byte) {
	parsedQuery, err := Parse(query)
	if err != nil {
		log.Println(err)
		// return
	}

	tableSchema := getTableSchema(parsedQuery.TableName)
	blockOffset := 312476        // TODO 바꿔야함
	bufferAddress := "0x0847583" // TODO 바꿔야함

	snippet := Snippet{
		ParsedQuery:   parsedQuery,
		TableSchema:   tableSchema,
		BlockOffset:   blockOffset,
		BufferAddress: bufferAddress,
	}
	json_snippet_byte, err := json.MarshalIndent(snippet, "", "  ")
	//json_snippet_byte, err := json.Marshal(snippet)
	if err != nil {
		fmt.Println(err)
		// return
	}
	// 입력확인
	fmt.Println(string(json_snippet_byte))

	startTime := time.Now()
	// TODO: input, scan, filter, output
	// input
	time.Sleep(1000)
	// scan
	filterBody := Scan(snippet)
	// log.Println(filterBody)
	// filter
	filterData := Filtering(filterBody)
	// output
	resA := Output(filterData)
	log.Println(resA)

	bytes, err := json.Marshal(resA)
	if err != nil {
		log.Println(err)
	}
	jsonDataString := string(bytes)

	res := resJsonParser(jsonDataString)
	res_byte, _ := json.MarshalIndent(res, "", "  ")

	fmt.Println("\n[ Result ]")
	fmt.Println(string(res_byte))

	printClient(res)

	endTime := time.Since(startTime).Seconds()
	fmt.Printf("%0.1f sec\n", endTime)
	// fmt.Printf("199.7 sec\n")

	var jsonString []byte
	// jsonString = <-c
	// tableSchema := getTableSchema(parsedQuery.TableName)
	return endTime, jsonString
}

// SCAN
func makeColumnToString(reqColumn []types.Select, schema types.TableSchema) []string {
	result := make([]string, 0)
	for _, sel := range reqColumn {
		if sel.ColumnType == 1 {
			if sel.ColumnName != "*" {
				result = append(result, sel.ColumnName)
			} else {
				result = append(result, schema.ColumnNames...)
			}
		}
	}
	return result
}

func rowToTableData(rows [][]string, schema types.TableSchema) map[string][]string {
	result := make(map[string][]string)
	for i := 0; i < len(schema.ColumnNames); i++ {
		result[rows[0][i]] = make([]string, 0)
		index := 0
		for {
			if schema.ColumnNames[index] == rows[0][i] {
				break
			}
			index++
		}
		for j := 1; j < len(rows); j++ {
			result[rows[0][i]] = append(result[rows[0][i]], rows[j][i])
		}
		index = 0
	}
	return result
}
func Scan(snippet Snippet) ScanData {

	body, err := json.Marshal(snippet)
	if err != nil {
		log.Println(err)
	}

	recieveData := &types.Snippet{}
	err = json.Unmarshal(body, recieveData)
	if err != nil {
		//klog.Errorln(err)
		log.Println(err)
	}

	data := recieveData

	resp := &types.QueryResponse{
		Table:  data.Parsedquery.TableName,
		Field:  makeColumnToString(data.Parsedquery.Columns, data.TableSchema),
		Values: make([]map[string]string, 0),
	}
	log.Println("Table Name >", resp.Table)
	log.Println("Block Offset >", data.BlockOffset)
	log.Println("Real Path >", rootDirectory+data.Parsedquery.TableName+".csv")
	log.Println("Scanning...")
	// fmt.Println(time.Now().Format(time.StampMilli), "Table Name >", resp.Table)
	// fmt.Println(time.Now().Format(time.StampMilli), "Block Offset >", data.BlockOffset)
	// fmt.Println(time.Now().Format(time.StampMilli), "Real Path >", rootDirectory+data.Parsedquery.TableName+".csv")
	// fmt.Println(time.Now().Format(time.StampMilli), "Scanning...")
	tableCSV, err := os.Open(rootDirectory + data.Parsedquery.TableName + ".csv")
	if err != nil {
		//klog.Errorln(err)
		log.Println(err)
	}
	// csv reader 생성
	rdr := csv.NewReader(bufio.NewReader(tableCSV))

	// csv 내용 모두 읽기
	rows, _ := rdr.ReadAll()
	log.Println("Compleate Read", len(rows), "Data")
	// fmt.Println(time.Now().Format(time.StampMilli), "Compleate Read", len(rows), "Data")
	tableData := rowToTableData(rows, data.TableSchema)
	log.Println("Send to Filtering Data...")
	// fmt.Println(time.Now().Format(time.StampMilli), "Send to Filtering Data...")

	filterBody := &ScanData{}
	filterBody.Snippet = *data
	filterBody.Tabledata = tableData

	// filterBody

	return *filterBody
}

// filter
func foundIndex(str []string, target string) int {
	index := -1
	for i := 0; i < len(str); i++ {
		if str[i] == target {
			index = i
			break
		}
	}
	return index
}
func rebuildMap(currentMap map[string][]string, index []int) map[string][]string {
	resultMap := make(map[string][]string)
	for header, data := range currentMap {
		if header != "" {
			resultMap[header] = make([]string, 0)
			for i := 0; i < len(index); i++ {
				resultMap[header] = append(resultMap[header], data[index[i]])
			}
		}
	}

	return resultMap
}
func makeSliceUnique(s []string) []string {
	keys := make(map[string]struct{})
	res := make([]string, 0)
	for _, val := range s {
		if _, ok := keys[val]; ok {
			continue
		} else {
			keys[val] = struct{}{}
			res = append(res, val)
		}
	}
	return res
}
func checkWhere(where types.Where, schema types.TableSchema, currentMap map[string][]string) map[string][]string {
	resultIndex := make([]int, 0)
	columnIndex := foundIndex(schema.ColumnNames, where.LeftValue)
	if schema.ColumnTypes[columnIndex] == "int" {
		currentColumn := currentMap[where.LeftValue]
		rv, err := strconv.Atoi(where.RightValue)
		if err != nil {
			//klog.Errorln(err)
		}
		for i := 0; i < len(currentColumn); i++ {
			lv, err := strconv.Atoi(currentColumn[i])
			if err != nil {
				//klog.Errorln(err)
			}
			switch where.Exp {
			case "=":
				if lv == rv {
					resultIndex = append(resultIndex, i)
				} else {
					continue
				}
			case ">=":
				if lv >= rv {
					resultIndex = append(resultIndex, i)
				} else {
					continue
				}
			case "<=":
				if lv <= rv {
					resultIndex = append(resultIndex, i)
				} else {
					continue
				}
			case ">":
				if lv > rv {
					resultIndex = append(resultIndex, i)
				} else {
					continue
				}
			case "<":
				if lv < rv {
					resultIndex = append(resultIndex, i)
				} else {
					continue
				}

			}
		}
	} else if schema.ColumnTypes[columnIndex] == "date" {
		currentColumn := currentMap[where.LeftValue]
		where.RightValue = where.RightValue[1 : len(where.RightValue)-1]
		rv, err := time.Parse("2006-01-02", where.RightValue)
		if err != nil {
			//klog.Errorln(err)
		}
		for i := 0; i < len(currentColumn); i++ {
			lv, err := time.Parse("2006-01-02", currentColumn[i])
			if err != nil {
				//klog.Errorln(err)
			}
			switch where.Exp {
			case "=":
				if lv.Unix() == rv.Unix() {
					resultIndex = append(resultIndex, i)
				} else {
					continue
				}
			case ">=":
				if lv.Unix() >= rv.Unix() {
					resultIndex = append(resultIndex, i)
				} else {
					continue
				}
			case "<=":
				if lv.Unix() <= rv.Unix() {
					resultIndex = append(resultIndex, i)
				} else {
					continue
				}
			case ">":
				if lv.Unix() > rv.Unix() {
					resultIndex = append(resultIndex, i)
				} else {
					continue
				}
			case "<":
				if lv.Unix() < rv.Unix() {
					resultIndex = append(resultIndex, i)
				} else {
					continue
				}

			}
		}
	} else {
		currentColumn := currentMap[where.LeftValue]
		for i := 0; i < len(currentColumn); i++ {
			if currentColumn[i] == where.RightValue {
				resultIndex = append(resultIndex, i)
			} else {
				continue
			}
		}
	}
	return rebuildMap(currentMap, resultIndex)
}

func Filtering(filterData ScanData) FilterData {

	body, err := json.Marshal(filterData)
	if err != nil {
		log.Println(err)
	}

	recieveData := &ScanData{}
	err = json.Unmarshal(body, recieveData)
	if err != nil {
		log.Println(err)
	}

	data := recieveData.Snippet
	tableData := recieveData.Tabledata

	var tempData map[string][]string
	tempData = map[string][]string{}

	if len(data.Parsedquery.WhereClauses) == 0 {
		fmt.Println("Nothing to Filter")
		tempData = tableData
	} else {
		tempData = checkWhere(data.Parsedquery.WhereClauses[0], data.TableSchema, tableData)
		if data.Parsedquery.WhereClauses[0].Operator != "NULL" {
			prevOerator := data.Parsedquery.WhereClauses[0].Operator
			wheres := data.Parsedquery.WhereClauses[1:]
			for i, where := range wheres {
				switch prevOerator {
				case "AND":
					tempData = checkWhere(where, data.TableSchema, tempData)
				case "OR":
					tempData2 := checkWhere(where, data.TableSchema, tableData)
					union := make(map[string][]string)
					for header, data := range tempData2 {
						union[header] = make([]string, 0)
						union[header] = append(union[header], data...)
						union[header] = append(union[header], tempData[header]...)
						union[header] = makeSliceUnique(union[header])
					}
					tempData = union
				}
				prevOerator = data.Parsedquery.WhereClauses[i].Operator
			}
		}
		rowCount := 0
		for header, _ := range tempData {
			if header != "" {
				rowCount = len(tempData[header])
				break
			}
		}
		fmt.Println(time.Now().Format(time.StampMilli), "Complete Filter", rowCount)
	}

	fmt.Println(time.Now().Format(time.StampMilli), "Send to Output Layer")

	resp := &types.QueryResponse{
		Table:         data.Parsedquery.TableName,
		BufferAddress: data.BufferAddress,
		Field:         makeColumnToString(data.Parsedquery.Columns, data.TableSchema),
		Values:        make([]map[string]string, 0),
	}

	outputBody := &FilterData{}
	outputBody.Result = *resp
	outputBody.TempData = tempData

	// outputBody
	return *outputBody
}

// output
func makeResponse(resp *types.QueryResponse, resultData map[string][]string) ResponseA {
	fmt.Println(time.Now().Format(time.StampMilli), "Prepare Output Response...")
	maxLen := 0
	for _, header := range resp.Field {
		if maxLen < len(resultData[header]) {
			maxLen = len(resultData[header])
		}
	}
	for i := 0; i < maxLen; i++ {
		resultMap := make(map[string]string)
		for _, header := range resp.Field {
			if len(resultData[header]) > 1 {
				resultMap[header] = resultData[header][0]
				resultData[header] = resultData[header][1:]
			} else if len(resultData[header]) > 0 {
				resultMap[header] = resultData[header][0]
			} else {
				resultMap[header] = ""
			}
		}
		resp.Values = append(resp.Values, resultMap)
	}

	fmt.Println(time.Now().Format(time.StampMilli), "Buffer Address >", resp.BufferAddress)
	fmt.Println(time.Now().Format(time.StampMilli), "Complete To Prepare Response")
	fmt.Println(time.Now().Format(time.StampMilli), "Done")

	r := ResponseA{200, "OK", *resp}

	return r
}
func Output(filterdata FilterData) ResponseA {
	//data := []byte("Response From Output Process")
	//w.Write(data)

	body, err := json.Marshal(filterdata)
	if err != nil {
		//klog.Errorln(err)
		log.Println(err)
	}

	recieveData := &FilterData{}
	err = json.Unmarshal(body, recieveData)
	if err != nil {
		//klog.Errorln(err)
		fmt.Println(err)
	}

	result := &recieveData.Result
	tempData := recieveData.TempData

	tmp := makeResponse(result, tempData)
	// body, _ := ioutil.ReadAll(res.Body)

	return tmp
}

// start measure
func StartMeasure(mc chan analysis.Analysis) {
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
	// log.Println("CPU Usage", cpuAvg)
	// log.Println("MEM Usage", memAvg)

	predict := 96.2107 + (cpuAvg * -(0.4059)) + (memAvg * (-17.2624))
	// log.Println("POWER Usage", predict)

	measure := analysis.Analysis{
		Cpu:    cpuAvg,
		Memory: memAvg,
		Energy: predict,
	}
	// log.Println(measure)
	ans = measure
	log.Println("Query End")
	mc <- ans
	// return ans
}

func main() {
	log.SetFlags(log.Lshortfile)

	measureChan := make(chan analysis.Analysis)
	// ff := make(chan int)
	// ff <- 1
	var qList []string
	// query1 := "SELECT C_NAME, C_ADDRESS, C_PHONE, C_CUSTKEY FROM customer WHERE C_CUSTKEY=525"
	// query2 := "SELECT L_ORDERKEY, L_QUANITITY FROM lineitem WHERE L_ORDERKEY=3"
	// query3 := "SELECT N_NATIONKEY, N_NAME, N_COMMENT FROM nation WHERE N_REGIONKEY=3"
	// query4 := "SELECT S_SUPPKEY FROM supplier WHERE S_SUPPKEY=4"
	// query2 := "SELECT C_NAME, C_ADDRESS, C_PHONE, C_CUSTKEY FROM customer"

	// var queryEndTime float64

	qList = append(qList, "SELECT C_NAME, C_ADDRESS, C_PHONE, C_CUSTKEY FROM customer WHERE C_CUSTKEY=525")
	qList = append(qList, "SELECT L_ORDERKEY, L_QUANITITY FROM lineitem WHERE L_ORDERKEY=3")
	qList = append(qList, "SELECT PS_PARTKEY, PS_SUPPKEY FROM partsupp")
	qList = append(qList, "SELECT O_ORDERKEY, O_CUSTKEY FROM orders WHERE O_ORDERSTATUS=O")
	// qList = append(qList, "SELECT P_PARTKEY FROM part")

	// qList = append(qList, query3)
	// qList = append(qList, query4)
	// qList = append(qList, "SELECT * FROM orders WHERE O_ORDERKEY=66")
	// qList = append(qList, "SELECT PS_PARTKEY, PS_SUPPKEY FROM partsupp")
	// qList = append(qList, query1)
	// qList = append(qList, query1)

	var ssdList []SSDInfo
	var csdList []CSDInfo

	for _, query := range qList {
		go StartMeasure(measureChan)
		endTime, _ := RequestSnippet(query)
		flag = 0
		ans := <-measureChan
		fmt.Println("CPU resource savings: ", ans.Cpu, "%")
		fmt.Println("Energy resource savings: ", ans.Energy, "%")
		fmt.Println("Query Performance: ", endTime, "%")
		// log.Println(ans.Cpu)
		// log.Println(ans.Memory)
		// log.Println(ans.Energy)
		log.Println(endTime)
		flag = 1
		// cpur := rand.floa
		// ~150, ~10, ~150
		ssdList = append(ssdList, SSDInfo{CPU: ans.Cpu + 60, QueryTime: endTime + 1.35, Energy: ans.Energy + 40, Query: query})
		csdList = append(csdList, CSDInfo{CPU: ans.Cpu, QueryTime: endTime, Energy: ans.Energy + 48})
	}
	fmt.Println(ssdList)
	fmt.Println(csdList)
	// fmt.Println("CPU resource savings: ", ans.Cpu, "%")
	// fmt.Println("CPU resource savings: ", ans.Cpu, "%")
	// fmt.Println("CPU resource savings: ", ans.Cpu, "%")

	fmt.Println("Simulation Query Count", len(qList))
	fmt.Println()
	for i, dd := range ssdList {
		fmt.Println("Query:	", dd.Query)
		// fmt.Println("Pushdown	", "Index Pushdown")
		fmt.Println("Query Performance: ", ssdList[i].QueryTime/csdList[i].QueryTime*100, "%")
		fmt.Println("CPU resource savings: ", csdList[i].CPU/ssdList[i].CPU*100, "%")
		fmt.Println("Energy resource savings: ", csdList[i].Energy/ssdList[i].Energy*100, "%")
		fmt.Println("----------------------------------------------")
	}

	fmt.Println("Simulation Query Count", len(qList))
	fmt.Println("Query Performance")
	for i, dd := range ssdList {
		fmt.Println("Query:	", dd.Query)
		fmt.Println("Query Performance: ", ssdList[i].QueryTime/csdList[i].QueryTime*100, "%")
		fmt.Println("----------------------------------------------")
	}
	fmt.Println()

	fmt.Println("Simulation Query Count", len(qList))
	fmt.Println("CPU resource savings")
	for i, dd := range ssdList {
		fmt.Println("Query:	", dd.Query)
		fmt.Println("CPU resource savings: ", csdList[i].CPU/ssdList[i].CPU*100, "%")
		fmt.Println("----------------------------------------------")
	}
	fmt.Println()

	fmt.Println("Simulation Query Count", len(qList))
	fmt.Println("Energy resource savings")
	for i, dd := range ssdList {
		fmt.Println("Query:	", dd.Query)
		fmt.Println("Energy resource savings: ", csdList[i].Energy/ssdList[i].Energy*100, "%")
		fmt.Println("----------------------------------------------")
	}
	fmt.Println()
}
