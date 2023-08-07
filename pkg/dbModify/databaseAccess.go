package dbModify

import (
	"database/sql"
	"fmt"
	"github.com/mailru/easyjson"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

type MatrixEntry struct {
	Id                                                                       int
	ContractStart, State, Util, Zone, RateCode, ProductOption, BillingMethod string
	Term                                                                     int
	UsageLower, UsageMiddle, UsageUpper                                      float32
}

type QueryParameters struct {
	StartDate   string `json:"startDate"`
	Util        string `json:"util"`
	DualBilling bool   `json:"dualBilling"`
	Terms       []int  `json:"terms"`
}

var UserParameters = QueryParameters{}

const parametersFilePath = "./parameters.json"
const dataSourcePath = "./data.db"
const dbInitSQL string = `

CREATE TABLE IF NOT EXISTS matrix (
	id 					INTEGER PRIMARY KEY,
	contract_start 		DATE NOT NULL,
	state_code 			TEXT NOT NULL,
	util_code			TEXT NOT NULL,
	util_zone			TEXT NOT NULL,
	util_rate_code		TEXT,
	product_option		TEXT NOT NULL,
	billing_method 		TEXT NOT NULL,
	contract_term		INTEGER NOT NULL,
	usage_lower			FLOAT NOT NULL,
	usage_middle 		FLOAT NOT NULL,
	usage_upper 		FLOAT NOT NULL
);

CREATE TABLE IF NOT EXISTS state_codes (
    id				INTEGER PRIMARY KEY,
    state_code		TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS util_codes (
    id				INTEGER NOT NULL PRIMARY KEY,
    util_code		TEXT NOT NULL,
	util_name		TEXT NOT NULL
);

INSERT INTO state_codes (state_code) VALUES 
	("DE"),
	("IL"),
	("MA"),
	("MD"),
	("ME"),
	("NH"),
	("NJ"),
	("OH"),
	("PA"),
	("RI");

INSERT INTO util_codes (util_code,util_name) VALUES
	("AECO",	"Atlantic City Electric Company"),			
	("APS",		"Potomac Edison - Allegheny Power"),			
	("BED",		"Eversource - Boston Edison"), 			
	("BGE",		"Baltimore Gas and Electric"),			
	("CAMB",	"Eversource - Cambridge Electric"), 			
	("CEI",		"Cleveland Electric"),			
	("CGE",		"Duke Energy"),			
	("CHGE",	"Central Hudson"),			
	("CILCO",	"Ameren Rate Zone II - CILCO"),			
	("CIPS",	"Ameren Rate Zone I - CIPS"),			
	("CMP",		"Central Maine Power"),			
	("COME",	"Eversource - Commonwealth Electric"),			
	("COMED",	"Commonwealth Edison"),			
	("CONE",	"Consolidated Edison"),			
	("CS",		"AEP - CS"),			
	("DELM",	"Conectiv Delmarva"),			
	("DELMDE",	"Delmarva"),			
	("DLCO",	"Duquesne Light Company"),
	("DPL",		"Dayton Power and Light"),
	("FGE",		"Unitil - Fitchburg Gas and Electric"),			
	("GSECO",	"Granite State Electric Co (Liberty Utilities)"),			
	("ILPWR",	"Ameren Rate Zone III - IP"),			
	("JCPL",	"Jersey Central Power & Light Company"),			
	("METED",	"Metropolitan Edison Company"),			
	("MSEL",	"National Grid - Massachusetts Electric Company"),			
	("MWST",	"Eversource - Western Massachusetts Electric"),			
	("NECO",	"National Grid - Narragansett Electric"),			
	("NHEC",	"New Hampshire Electric Co"),			
	("NIMO",	"National Grid - Niagara Mohawk"),			
	("NYOR",	"Orange and Rockland"),			
	("NYSEG",	"New York State Electric & Gas"),			
	("OE",		"Ohio Edison"),		
	("OPCO",	"AEP - OP"),			
	("PECO",	"PECO Energy"),			
	("PENELEC",	"Pennsylvania Electric Company"),			
	("PEPCO",	"Potomac Electric Power Company"),			
	("PP",		"Pennsylvania Power Company"),			
	("PPL",		"Pennsylvania Power and Light, Inc."),			
	("PSEG",	"Public Service Electric and Gas Company"),			
	("PSNH",	"Public Service Of New Hampshire"),			
	("RECO",	"Rockland Electric Company"),			
	("RGE",		"Rochester Gas & Electric"),			
	("TE",		"Toledo Edison"),			
	("UNITIL",	"Unitil Energy Systems"),			
	("WPP",		"Allegheny Power WPP");
`

func ProcessRow(row []string) bool {

	var contractStart, state, utility, zone, rateCodes, productOptions, billingMethod, term, usageLower, usageMiddle, usageUpper string

	contractStart = row[0]
	state = row[1]
	utility = row[2]
	zone = row[3]
	rateCodes = row[4]
	productOptions = row[5]
	billingMethod = row[6]
	term = row[7]
	usageLower = row[8]
	usageMiddle = row[9]
	usageUpper = row[10]

	db, openErr := sql.Open("sqlite", dataSourcePath)
	cobra.CheckErr(openErr)

	defer func(db *sql.DB) {
		err := db.Close()
		cobra.CheckErr(err)
	}(db)

	insertSQL := `INSERT INTO matrix (contract_start, state_code, util_code, util_zone, util_rate_code, product_option, billing_method, contract_term, usage_lower, usage_middle, usage_upper) 
	VALUES (?,?,?,?,?,?,?,?,?,?,?);`
	stmt, err := db.Prepare(insertSQL)
	cobra.CheckErr(err)
	res, err := stmt.Exec(contractStart, state, utility, zone, rateCodes, productOptions, billingMethod, term, usageLower, usageMiddle, usageUpper)
	cobra.CheckErr(err)

	id, err := res.LastInsertId()
	cobra.CheckErr(err)

	fmt.Println("Id inserted: " + strconv.FormatInt(id, 10))
	return true
}

func GetFilteredEntries() []MatrixEntry {
	fmt.Println(`Getting filtered entries...`)

	userParameters := ReadJson()
	LoadParameters(userParameters)

	db, openErr := sql.Open("sqlite", dataSourcePath)
	cobra.CheckErr(openErr)

	defer func(db *sql.DB) {
		err := db.Close()
		cobra.CheckErr(err)
	}(db)

	numParams := 0
	var paramsToInsert []string
	querySQL := `SELECT * FROM matrix WHERE `

	var rows *sql.Rows
	var err error

	fmt.Println("\nStart Date: " + UserParameters.StartDate)
	if UserParameters.StartDate != "" {
		querySQL += `contract_start = ? AND `
		//startParam = UserParameters.StartDate
		paramsToInsert = append(paramsToInsert, UserParameters.StartDate)
		numParams++
	}
	fmt.Println("Util: " + UserParameters.Util)
	if UserParameters.Util != "" {
		querySQL += `util_code = ? AND `
		paramsToInsert = append(paramsToInsert, UserParameters.Util)
		numParams++
	}
	fmt.Println("Dual Billing: " + strconv.FormatBool(UserParameters.DualBilling))
	if !(UserParameters.DualBilling) {
		querySQL += `billing_method != 'Dual' AND `
	}
	termsString, err := ffjson.Marshal(UserParameters.Terms)
	fmt.Println("Terms: " + string(termsString))
	if len(UserParameters.Terms) == 4 {
		querySQL += `contract_term IN (?,?,?,?);`
		switch numParams {
		case 0:
			rows, err = db.Query(querySQL, (UserParameters.Terms)[0], (UserParameters.Terms)[1], (UserParameters.Terms)[2], (UserParameters.Terms)[3])
			cobra.CheckErr(err)
		case 1:
			rows, err = db.Query(querySQL, paramsToInsert[0], (UserParameters.Terms)[0], (UserParameters.Terms)[1], (UserParameters.Terms)[2], (UserParameters.Terms)[3])
			cobra.CheckErr(err)
		case 2:
			rows, err = db.Query(querySQL, paramsToInsert[0], paramsToInsert[1], (UserParameters.Terms)[0], (UserParameters.Terms)[1], (UserParameters.Terms)[2], (UserParameters.Terms)[3])
			cobra.CheckErr(err)
		default:
			fmt.Println("error")
		}
	} else {
		querySQL += `contract_term IN (contract_term);`
		switch numParams {
		case 0:
			rows, err = db.Query(querySQL)
			cobra.CheckErr(err)
		case 1:
			rows, err = db.Query(querySQL, paramsToInsert[0])
			cobra.CheckErr(err)
		case 2:
			rows, err = db.Query(querySQL, paramsToInsert[1])
			cobra.CheckErr(err)
		default:
			fmt.Println("error")
		}
	}

	defer func(row *sql.Rows) {
		err := row.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(rows)

	var entries []MatrixEntry

	for rows.Next() {
		entry := MatrixEntry{}
		err := rows.Scan(&entry.Id, &entry.ContractStart, &entry.State, &entry.Util, &entry.Zone, &entry.RateCode, &entry.ProductOption, &entry.BillingMethod, &entry.Term, &entry.UsageLower, &entry.UsageMiddle, &entry.UsageUpper)
		if err != nil {
			fmt.Println(err)
		}
		entries = append(entries, entry)
	}

	return entries

}

func GetUtilCodes() [][]string {
	db, openErr := sql.Open("sqlite", dataSourcePath)
	cobra.CheckErr(openErr)

	defer func(db *sql.DB) {
		err := db.Close()
		cobra.CheckErr(err)
	}(db)

	querySQL := `SELECT util_code,util_name FROM util_codes; `

	rows, err := db.Query(querySQL)
	cobra.CheckErr(err)

	var utilCodes [][]string

	for rows.Next() {
		var utilPair []string
		var utilCode string
		var utilName string
		err := rows.Scan(&utilCode, &utilName)
		if err != nil {
			return nil
		}
		utilPair = append(utilPair, utilCode, utilName)
		utilCodes = append(utilCodes, utilPair)
	}
	return utilCodes
}

func InitializeDatabase() {
	db, openErr := sql.Open("sqlite", dataSourcePath)
	cobra.CheckErr(openErr)

	defer func(db *sql.DB) {
		err := db.Close()
		cobra.CheckErr(err)
	}(db)
	_, err := db.Exec(dbInitSQL)
	cobra.CheckErr(err)
}

func ReInitializeDatabase() {
	db, openErr := sql.Open("sqlite", dataSourcePath)
	cobra.CheckErr(openErr)

	defer func(db *sql.DB) {
		err := db.Close()
		cobra.CheckErr(err)
	}(db)
	dropInitSQL := `DROP TABLE IF EXISTS matrix;` + dbInitSQL
	_, err := db.Exec(dropInitSQL)
	cobra.CheckErr(err)
}

func LoadParameters(newParameters QueryParameters) {
	UserParameters = newParameters
	PrintParameters()
}

func PrintParameters() {
	fmt.Println("User Parameters: ")
	fmt.Println(UserParameters)
}

func ReadJson() QueryParameters {
	data, err := os.ReadFile(parametersFilePath)
	if err != nil {
		fmt.Println(err)
		return QueryParameters{}
	}
	defaultParameters := &(QueryParameters{})
	err = easyjson.Unmarshal(data, defaultParameters)
	if err != nil {
		fmt.Println(err)
		return QueryParameters{}
	}
	return *defaultParameters
}

func writeJson() {
	data, err := easyjson.Marshal(UserParameters)
	cobra.CheckErr(err)

	err = os.WriteFile("parameters.json", data, os.ModePerm)
	cobra.CheckErr(err)

}

func SetStartDate(startDate string) {
	parameters := ReadJson()
	LoadParameters(parameters)
	(UserParameters).StartDate = startDate
	writeJson()
	PrintParameters()
}

func SetUtil(util string) {
	parameters := ReadJson()
	LoadParameters(parameters)
	(UserParameters).Util = util
	writeJson()
	PrintParameters()
}

func SetDualBilling(dualBilling bool) {
	parameters := ReadJson()
	LoadParameters(parameters)
	(UserParameters).DualBilling = dualBilling
	writeJson()
	PrintParameters()
}

func SetTerms(terms []int) {
	parameters := ReadJson()
	LoadParameters(parameters)
	(UserParameters).Terms = terms
	writeJson()
	PrintParameters()
}
