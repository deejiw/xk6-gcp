package gcp

import (
	"context"
	"fmt"
	"log"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Direct interface from https://pkg.go.dev/google.golang.org/api/sheets/v4#SpreadsheetsValuesService.

// This function retrieves data from a Google Sheet.
// Parameters:
// - spreadsheetId: the ID of the Google Sheet.
// - sheetName: the name of the sheet to retrieve data from.
// - cellRange: the range of cells to retrieve data from.
// Returns:
// - [][]interface{}: a 2D slice of interface{} values representing the retrieved data.
// - error: an error if one occurred, otherwise nil.
func (g *Gcp) SpreadsheetGet(spreadsheetId string, sheetName string, cellRange string) ([][]interface{}, error) {
	g.sheetClient()

	res, err := g.sheet.Spreadsheets.Values.Get(spreadsheetId, fmt.Sprintf("%s!%s", sheetName, cellRange)).Do()
	if err != nil || res.HTTPStatusCode != 200 {
		log.Fatalf("unable to get data from range %s in sheet %s  <%v>.", cellRange, sheetName, err)
	}

	if len(res.Values) == 0 {
		fmt.Printf("No data found in range %s on sheet %s.", cellRange, sheetName)
		return nil, nil
	}

	return res.Values, nil
}

// Appends a row of data to a Google Sheet.
// Parameters:
// - spreadsheetId: the ID of the Google Sheet.
// - sheetName: the name of the sheet to append data to.
// - valueRange: a slice of interface{} values representing the data to append.
// Returns:
// - string: an empty string.
// - error: an error if one occurred, otherwise nil.
func (g *Gcp) SpreadsheetAppend(spreadsheetId string, sheetName string, valueRange []interface{}) (string, error) {
	ctx := context.Background()
	g.sheetClient()

	row := &sheets.ValueRange{
		Values: [][]interface{}{valueRange},
	}

	res, err := g.sheet.Spreadsheets.Values.Append(spreadsheetId, sheetName, row).ValueInputOption("RAW").InsertDataOption("INSERT_ROWS").Context(ctx).Do()
	if err != nil || res.HTTPStatusCode != 200 {
		log.Fatalf("unable to append data into sheet %s <%v>.", sheetName, err)
	}

	return "", nil
}

// Similar to https://pkg.go.dev/google.golang.org/api/sheets/v4#SpreadsheetsService.GetByDataFilter
// Get a row from a Google Sheet based on filters.
// Parameters:
// - spreadsheetId: the ID of the Google Sheet.
// - sheetName: the name of the sheet to search data in.
// - filters: a map of column names to values to search for in the specified column.
// Returns:
// - map[string]interface{}: a map of the row data if a match is found.
// - error: an error if one occurred, otherwise nil.
func (g *Gcp) SpreadsheetGetRowByFilters(spreadsheetId string, sheetName string, filters map[string]string) (map[string]interface{}, error) {
	cellRange := g.findCellRange(spreadsheetId, sheetName)
	rows, _ := g.SpreadsheetGet(spreadsheetId, sheetName, cellRange)
	headers := rows[0]

	// Find matching rows based on the filters
	for _, row := range rows {
		match := true
		for key, value := range filters {
			headerIndex := findHeaderIndex(headers, key)
			if headerIndex == -1 || headerIndex >= len(row) || strings.TrimSpace(row[headerIndex].(string)) != value {
				match = false
				break
			}
		}
		if match {
			return mergeKV(headers, row), nil
		}
	}

	fmt.Printf("No row matches filters %v", filters)
	return nil, nil
}

// This function appends a row of data to a Google Sheet with a unique ID.
// Parameters:
// - spreadsheetId: the ID of the Google Sheet.
// - sheetName: the name of the sheet to append data to.
// - values: a slice of interface{} values representing the data to append.
// Returns:
// - int64: the unique ID of the appended row.
// - error: an error if one occurred, otherwise nil.
func (g *Gcp) SpreadsheetAppendWithUniqueId(spreadsheetId string, sheetName string, values map[string]interface{}) (int64, error) {
	ctx := context.Background()
	g.sheetClient()

	rows, _ := g.SpreadsheetGet(spreadsheetId, sheetName, "A:A")
	uniqueId := getUniqueId(rows)

	sorted := sortValuesByHeaders(rows[0], values)

	row := &sheets.ValueRange{
		Values: [][]interface{}{append([]interface{}{uniqueId}, sorted...)},
	}

	res, err := g.sheet.Spreadsheets.Values.Append(spreadsheetId, sheetName, row).ValueInputOption("RAW").Context(ctx).Do()
	if err != nil || res.HTTPStatusCode != 200 {
		log.Fatalf("unable to append data into sheet %s <%v>.", sheetName, err)
	}

	return uniqueId, nil
}

func (g *Gcp) SpreadsheetGetRowByFiltersAndAppendIfNotExist(spreadsheetId string, sheetName string, filters map[string]string, values map[string]interface{}) (int64, error) {
	var id int64
	ctx := context.Background()
	g.sheetClient()

	rowByFilters, _ := g.SpreadsheetGetRowByFilters(spreadsheetId, sheetName, filters)
	rows, _ := g.SpreadsheetGet(spreadsheetId, sheetName, "A:A")

	if rowByFilters == nil {

		id = getUniqueId(rows)
	} else {
		return rowByFilters["id"].(int64), nil
	}

	sorted := sortValuesByHeaders(rows[0], values)

	row := &sheets.ValueRange{
		Values: [][]interface{}{append([]interface{}{id}, sorted...)},
	}

	res, err := g.sheet.Spreadsheets.Values.Append(spreadsheetId, sheetName, row).ValueInputOption("RAW").Context(ctx).Do()
	if err != nil || res.HTTPStatusCode != 200 {
		log.Fatalf("unable to append data into sheet %s <%v>.", sheetName, err)
	}

	return id, nil
}

// This function initializes the Google Sheets client.
func (g *Gcp) sheetClient() {
	if g.sheet == nil {
		ctx := context.Background()
		jwt, err := getJwtConfig(g.keyByte, g.scope)
		if err != nil {
			log.Fatalf("could not get JWT config with scope %s <%v>.", g.scope, err)
		}

		c, err := sheets.NewService(ctx, option.WithTokenSource(jwt.TokenSource(ctx)))

		if err != nil {
			log.Fatalf("could not initialize Sheets client <%v>.", err)
		}

		g.sheet = c
	}
}

// This function returns the cell range of the first row of a Google Sheet.
// Parameters:
// - spreadsheetId: the ID of the Google Sheet.
// - sheetName: the name of the sheet to retrieve data from.
// Returns:
// - string: the cell range of the first row.
func (g *Gcp) findCellRange(spreadsheetId string, sheetName string) string {
	rows, _ := g.SpreadsheetGet(spreadsheetId, sheetName, "1:1")
	return fmt.Sprintf("A:%s", columnIndexToLetter(len(rows[0])-1))
}