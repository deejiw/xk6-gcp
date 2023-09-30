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

func (g *Gcp) SpreadsheetAppend(spreadsheetId string, sheetName string, valueRange []interface{}) (string, error) {
	ctx := context.Background()
	g.sheetClient()

	row := &sheets.ValueRange{
		Values: [][]interface{}{valueRange},
	}

	res, err := g.sheet.Spreadsheets.Values.Append(spreadsheetId, sheetName, row).ValueInputOption("RAW").Context(ctx).Do()
	if err != nil || res.HTTPStatusCode != 200 {
		log.Fatalf("unable to append data into sheet %s <%v>.", sheetName, err)
	}

	return "", nil
}

func (g *Gcp) SpreadsheetUpdate(spreadsheetId string, sheetName string, cellRange string, valueRange []interface{}) (string, error) {
	ctx := context.Background()
	g.sheetClient()

	row := &sheets.ValueRange{
		Values: [][]interface{}{valueRange},
	}

	res, err := g.sheet.Spreadsheets.Values.Update(spreadsheetId, fmt.Sprintf("%s!%s", sheetName, cellRange), row).ValueInputOption("RAW").Context(ctx).Do()
	if err != nil || res.HTTPStatusCode != 200 {
		log.Fatalf("unable to update data into sheet %s range %s <%v>.", sheetName, cellRange, err)
	}

	return "", nil
}

// Enhancement to operate Google Sheet with SQL-like parameters.
func (g *Gcp) SpreadsheetGetRowByFirstColumn(spreadsheetId string, sheetName string, firstColumnValue string) (map[string]interface{}, error) {
	cellRange := g.findCellRange(spreadsheetId, sheetName)
	rows, _ := g.SpreadsheetGet(spreadsheetId, sheetName, cellRange)
	headers := rows[0]

	for _, row := range rows {
		if strings.TrimSpace(row[0].(string)) == firstColumnValue {
			return mergeKV(headers, row), nil
		}
	}

	fmt.Printf("No value in column %s matches %v", headers[0], firstColumnValue)
	return nil, nil
}

// Similar to https://pkg.go.dev/google.golang.org/api/sheets/v4#SpreadsheetsService.GetByDataFilter
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

func (g *Gcp) SpreadsheetAppendWithUniqueId(spreadsheetId string, sheetName string, values []interface{}) (int64, error) {
	ctx := context.Background()
	g.sheetClient()

	rows, _ := g.SpreadsheetGet(spreadsheetId, sheetName, "A:A")
	uniqueId := getUniqueId(rows)

	row := &sheets.ValueRange{
		Values: [][]interface{}{append([]interface{}{uniqueId}, values...)},
	}

	res, err := g.sheet.Spreadsheets.Values.Append(spreadsheetId, sheetName, row).ValueInputOption("RAW").Context(ctx).Do()
	if err != nil || res.HTTPStatusCode != 200 {
		log.Fatalf("unable to append data into sheet %s <%v>.", sheetName, err)
	}

	return uniqueId, nil
}

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

func (g *Gcp) findCellRange(spreadsheetId string, sheetName string) string {
	rows, _ := g.SpreadsheetGet(spreadsheetId, sheetName, "1:1")
	return fmt.Sprintf("A:%s", columnIndexToLetter(len(rows[0])-1))
}
