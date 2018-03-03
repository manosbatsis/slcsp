//
// A solution to the homework at https://github.com/adhocteam/homework/tree/master/slcsp written in Go.
//
// Please note this is a Java guy's very first Go program, an attempt to something useful for yours truly personally.
// The code was written for an interview test with no real use for the recipient anyway, so I saw it as a nice
// opportunity to give the language a try.
//
// Thus, you will probably not find a professional sample of Go code here, especially since I tend to see things from
// an OO perspective and have little knowledge of the language properties, idioms and conventions. This actually took
// hours to write while googling for the absolute basics. Even worse, anything is possible without a test case!
//
// The code makes the following assumptions:
//
//	- The objective is to find, for each zip code, the SLCSP of the rate area it corresponds to
//	- The county code is irrelevant to the problem, as rate areas is what actually connects zip codes plan rates
//	- The relationship between zip codes and rate areas is considered many-to-one in relational terms
// 	- Zip codes mapped to multiple rate areas are ignored, i.e. considered ambiguous and left blank per the instructions
// 	- Zip codes for which less than two silver plans are discovered are to be given No SLCSP value in the resulting file
// 	- An in-memory index is the reasonable performance trade-off, i.e. rather small input files are expected
//	- The `state` and `rate_area` columns in zips.scv and plans.csv compose the business key of a rate area (the tuple)
//	- All records in zips.csv and plans.csv are complete, i.e. without missing any column values
//	- No other input data validation or correction is necessary
//
// Tp run this program you need to have Go installed in your path. Simply navigate to the slcsp folder in your command
// line interface and execute the following:
//
// go run slcsp.go
//
// The program will produce the folowing output:
//
//	INFO:    2017/11/20 03:46:29 slcsp.go:201: Parsed 51541 records from zips.csv to 38804 zip codes and 477 rate areas
//	WARNING: 2017/11/20 03:46:29 slcsp.go:205: Note: zips.csv contained 3723 ambiguous zip codes (see trace or Appendix C in COMMENTS)
//	INFO:    2017/11/20 03:46:29 slcsp.go:377: Parsed 22239 records from plans.csv to 22239 plans
//	WARNING: 2017/11/20 03:46:29 slcsp.go:380: Note: plans.csv contained 7 unmapped areas (see trace or Appendix A in COMMENTS)
//	INFO:    2017/11/20 03:46:29 slcsp.go:352: Wrote 51 records to slcsp-modified.csv
//	WARNING: 2017/11/20 03:46:29 slcsp.go:355: Note: 10 zip codes had insufficient plan info, i.e. less than two plans (see trace or Appendix B in COMMENTS)
//
// Appendix A, B, C: see COMMENTS file

package main

import (
	"bufio"
	"encoding/csv"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

const LEVEL_NAME_SILVER = "Silver"

var (
	// Logging
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger

	// Input files
	zipsFilePath          string
	plansFilePath         string
	slcspFilePath         string
	slcspCompleteFilePath string

	// Used for labels validation, looks like there's no way to have array-based constants
	LABELS_PLANS = []string{"plan_id", "state", "metal_level", "rate", "rate_area"}
	LABELS_ZIPS  = []string{"zipcode", "state", "county_code", "name", "rate_area"}
)

// Initialization: logging etc.
// Accepts log writers for trace, info, warning and error levels.
func Init(traceHandle io.Writer, infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer) {

	// Input files, no need for cmd arguments at this point
	zipsFilePath = "zips.csv"
	plansFilePath = "plans.csv"
	slcspFilePath = "slcsp.csv"
	slcspCompleteFilePath = "slcsp-modified.csv"

	Trace = log.New(traceHandle,     "TRACE:   ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(infoHandle,       "INFO:    ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorHandle,     "ERROR:   ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {

	// Initialize logging. To trace, change the first argument to os.Stdout
	Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	// Get an index that maps zip codes to rate areas
	rateAreasIndex := NewRateAreasIndex(zipsFilePath)

	// Update index entries with plan information
	rateAreasIndex.ParsePlans(plansFilePath)

	// Parse the SLCSP CSV template and write a complete version
	// that includes the SLCSP rate per zip code in the same order
	rateAreasIndex.ToFile(slcspFilePath, slcspCompleteFilePath)

}

// Utility function to parse CSV files.Accepts a filename, a line-handling function
// and an optional labels line-handling function as arguments.
// Each line (record) is passed as a string to the line-handling function.
// If the headlineHandler argument is null, the first line is treated as any other.
// Returns the number of lines (records) processed.
func readCsvFileLines(filename string, lineHandler func(string), headlineHandler func(string)) int {

	// Get file for reading, set to cleanup
	file, _ := os.Open(filename)
	scanner := bufio.NewScanner(file)
	defer file.Close()

	// handle labels?
	if headlineHandler != nil {
		if scanner.Scan() {
			headlineHandler(scanner.Text())
		}
	}

	// lines counter
	recordsCount := 0

	// Parse CSV records. No need to use the encoding/csv package.
	// This parses the headings row as well but we don't seem to care
	for scanner.Scan() {

		// Pass each line to the given line handler function
		lineHandler(scanner.Text())
		// update the processed records counter
		recordsCount++
	}
	// error handling/report
	if err := scanner.Err(); err != nil {
		Error.Fatal(err)
	}
	return recordsCount
}

// A geographic region rates apply to. The area is a tuple of a state and a number, for example, NY 1, IL 14.
// Specifies the top 2 lowest rate for a silver plan in the rate area.
type RateArea struct {
	// the state two-letter postal abbreviation
	state string
	// the area number
	number string
	// the area's lowest rate for a silver plan
	top1SilverPlan float64
	// the area's second lowest rate for a silver plan
	top2SilverPlan float64
}

// Get the name of a rate area as a string. The form is "<state> <number>"
func (rateArea RateArea) GetName() string {
	return rateArea.state + " " + rateArea.number
}

// Get SLCSP for this rate area as a formatted string with up to two decimal points
// and no trailing zero digits (same as the original CSV input)
func (rateArea RateArea) GetSlcspString() string {
	return strconv.FormatFloat(rateArea.top2SilverPlan, 'f', -1, 64)
}

// Initializer for RateArea instances
func NewRateArea(state, number string) RateArea {
	area := RateArea{}
	area.state = state
	area.number = number
	return area
}

// Used to cather information regarding a specific zip code
type RateAreasIndex struct {
	rateAreaNamesByZip    map[string]string
	rateAreasByName       map[string]*RateArea
	ambiguousZipcodes     map[string]bool
	unmappedPlanRateAreas map[string]bool
	insufficientPlanZips  map[string]bool
}

// Create a zip - rate areas index from a zip codes file,
// assuming (zipcode,state,county_code,name,rate_area) columns
func NewRateAreasIndex(filename string) RateAreasIndex {

	// Initialize the index
	rateAreasIndex := RateAreasIndex{}
	rateAreasIndex.rateAreaNamesByZip = make(map[string]string)
	rateAreasIndex.rateAreasByName = make(map[string]*RateArea)
	rateAreasIndex.ambiguousZipcodes = make(map[string]bool)
	rateAreasIndex.unmappedPlanRateAreas = make(map[string]bool)
	rateAreasIndex.insufficientPlanZips = make(map[string]bool)

	// A CSV line handler that parses lines to zip-area mappings.
	// Lines are basically CSV records with the columns: zipcode,state,county_code,name,rate_area.
	linehandler := func(line string) {
		fields := strings.Split(line, ",")
		zipCode := fields[0]
		// Add or update mapping
		rateAreasIndex.AddMapping(zipCode, fields[1], fields[4])
	}

	// Parse file using the above function as the line handler
	recordsCount := readCsvFileLines(filename, linehandler, buildLabelsValidatingLinehandler(LABELS_ZIPS))

	// Report parsing work
	Info.Printf("Parsed %d records from %s to %d zip codes and %d rate areas\n",
		recordsCount, filename, len(rateAreasIndex.rateAreaNamesByZip), len(rateAreasIndex.rateAreasByName))
	// Report ambiguous zip codes
	if len(rateAreasIndex.ambiguousZipcodes) > 0 {
		Warning.Printf("Note: %s contained %d ambiguous zip codes (see trace or Appendix C in COMMENTS)", filename, len(rateAreasIndex.ambiguousZipcodes))
	}
	return rateAreasIndex
}

// Creates the corresponding entry if missing.
// Returns the rate area name in any case
func (rai RateAreasIndex) AddRateArea(state, number string) string {

	var rateAreaName string

	// Use the existing RateArea entry matching the state and number, if any
	if match, ok := rai.rateAreasByName[state+" "+number]; ok {
		rateAreaName = match.GetName()
	} else { // else add a new area and return it after adding to the index
		rateArea := NewRateArea(state, number)
		rateAreaName = rateArea.GetName()
		rai.rateAreasByName[rateAreaName] = &rateArea
	}
	return rateAreaName
}

// Add a mapping between the zip code and rate area it corresponds to.
// Return the RateArea associated with the mapping
func (rai RateAreasIndex) AddMapping(zipCode, state, number string) {

	areaName := rai.AddRateArea(state, number)

	// Check if a zip code mapping to an area already exists
	if mappedAreaName, ok := rai.rateAreaNamesByZip[zipCode]; ok {

		// If the existing rate area mapping for the given zip code
		// is different than the newly given, flag the zip code as ambiguous
		if mappedAreaName != areaName {
			rai.ambiguousZipcodes[zipCode] = true
			Trace.Printf("Ambiguous zip code %s: mapping to %s failed as a mapping to %s is already defined",
				zipCode, areaName, mappedAreaName)
		}
	} else { // add a new rate area mapping otherwise
		rai.rateAreaNamesByZip[zipCode] = areaName
	}
}

// Add the plan if it currently makes the Top 2 silver plans (i.e. rateArea.topSilverPlans)
// Return true if the plan made the Top 2, false otherwise
func (rai RateAreasIndex) AddPlan(state, number, planId, level, rate string) bool {

	madeIt := false

	// Is it a silver plan?
	if LEVEL_NAME_SILVER == level {

		// convert rate value
		newRate, err := strconv.ParseFloat(rate, 64)
		if err != nil {
			Error.Fatalf("Could not parse rate %s for plan %s", rate, planId)
		}

		// get the rate area entry
		mappingAreaName := state + " " + number
		if rateArea, ok := rai.rateAreasByName[mappingAreaName]; ok {

			// Is it our new top two plans?
			if newRate < rateArea.top1SilverPlan || rateArea.top1SilverPlan == 0 {
				rateArea.top2SilverPlan = rateArea.top1SilverPlan
				rateArea.top1SilverPlan = newRate
				madeIt = true
			} else if newRate < rateArea.top2SilverPlan || rateArea.top2SilverPlan == 0 { // ... or top 2?

				rateArea.top2SilverPlan = newRate
				madeIt = true
			}

			if madeIt {
				Trace.Printf("Ignored plan as rate: %s, was lower than the current top two plans: %s, %s", newRate, rateArea.top1SilverPlan, rateArea.top2SilverPlan)
			}

		} else { // the area does not exist!
			rai.unmappedPlanRateAreas[mappingAreaName] = true
			Trace.Printf("Plan applies to unmapped rate area: %s", mappingAreaName)
		}

	}

	return madeIt
}

// Get the SLCSP for the given zip code,
// if the entry exists and is not ambiguous
func (rai RateAreasIndex) GetSlcspForZipcode(zipCode string) string {
	slcsp := ""
	if _, ok := rai.ambiguousZipcodes[zipCode]; ok {
		// Ignore as zipcode is ambiguous
	} else {
		// Use formatted slcsp if area exists
		if existingAreaName, ok := rai.rateAreaNamesByZip[zipCode]; ok {
			slcsp = rai.rateAreasByName[existingAreaName].GetSlcspString()
		}

		// 0 means insufficient plans info, i.e. less than two plans for the zip code
		if(slcsp == "0"){
			// Reset to empty string
			slcsp = ""
			// Note zip code for reporting later
			rai.insufficientPlanZips[zipCode] = true
			Trace.Printf("Insufficient plans info, i.e. less than two plans for zip code: %s", zipCode)
		}
	}
	return slcsp
}

// Parse the SLCSP CSV template and write a complete version
// that includes the SLCSP rate per zip code in the same order
func (rai RateAreasIndex) ToFile(templateFileName string, outFileName string) {

	// setup the output and clean-up
	file, err := os.Create(outFileName)
	checkError("Cannot create file", err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// A CSV line handler that writes a complete version
	// of the original line, including the slcsp value.
	// Note: Ambiguous zip codes will have an empty slcsp value
	linehandler := func(line string) {

		// get the zipcode column value
		fields := strings.Split(line, ",")
		zipCode := fields[0]

		// get the slcsp column value if a non-ambiguous entry exists,
		// or an empty string otherwise
		slcsp := rai.GetSlcspForZipcode(zipCode)

		// write record
		record := []string{zipCode, slcsp}
		err := writer.Write(record)
		checkError("Cannot write to file", err)
	}

	// CSV line handler for cloning labels
	cloneLabels := func(line string) {
		err := writer.Write(strings.Split(line, ","))
		checkError("Cannot write to file", err)
	}

	// Parse file using the above function as the line handler
	recordsCount := readCsvFileLines(templateFileName, linehandler, cloneLabels)

	// Report
	Info.Printf("Wrote %d records to %s\n", recordsCount, outFileName)

	if len(rai.insufficientPlanZips) > 0 {
		Warning.Printf("Note: %d zip codes had insufficient plan info, i.e. less than two plans (see trace or Appendix B in COMMENTS)", len(rai.insufficientPlanZips))
	}
}

// Parse the given plans CSV file and update the index accordingly
func (rai RateAreasIndex) ParsePlans(filename string) {

	// A CSV line handler that adds each as a plan to the
	// rate area it applies to.
	// Expected columns: plan_id,state,metal_level,rate,rate_area
	linehandler := func(line string) {

		// get the individual columns
		fields := strings.Split(line, ",")
		// process the plan: state, number, planId, level, rate
		rai.AddPlan(fields[1], fields[4], fields[0], fields[2], fields[3])
	}

	// Parse file using the above function as the line handler
	recordsCount := readCsvFileLines(filename, linehandler, buildLabelsValidatingLinehandler(LABELS_PLANS))

	// Report
	Info.Printf("Parsed %d records from %s to %d plans\n",
		recordsCount, filename, recordsCount)
	if len(rai.unmappedPlanRateAreas) > 0 {
		Warning.Printf("Note: %s contained %d unmapped areas (see trace or Appendix A in COMMENTS)", filename, len(rai.unmappedPlanRateAreas))
	}
}

// Utility function to log an error and exit if the error exists
func checkError(message string, err error) {
	if err != nil {
		Error.Fatal(message, err)
	}
}

// No-op line handler, used to ignore CSV labels row
func linehandlerNoop(line string) {

}

// Factory for validating CSV  heading line handlers
func buildLabelsValidatingLinehandler(labels []string) func(string) {
	return func(line string) {
		fields := strings.Split(line, ",")
		if !areLabelsEqual(labels, fields) {
			Error.Fatalf("Invalid labels, expected: %s, actual: %s", labels, fields)
		}
	}
}

// Utility function to basically compare string slices,
// used here to compare labels for CSV file validation
func areLabelsEqual(a, b []string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
