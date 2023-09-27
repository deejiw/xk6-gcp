package gcp

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func (g *Gcp) SpreadsheetGet(spreadsheetId string, sheetName string, cellRange string) ([][]interface{}, error) {
	// sheetName := getSheetName(c, spreadsheetId, sheetId)
	if g.sheet == nil {
		g.sheetClient()
	}

	res, err := g.sheet.Spreadsheets.Values.Get(spreadsheetId, fmt.Sprintf("%s!%s", sheetName, cellRange)).Do()
	if err != nil || res.HTTPStatusCode != 200 {
		log.Fatalf("unable to get data from sheet %s range %s <%v>.", sheetName, cellRange, err)
	}

	return res.Values, nil
}

func (g *Gcp) SpreadsheetAppend(spreadsheetId string, sheetName string, cellRange string, valueRange []interface{}) (string, error) {
	if g.sheet == nil {
		g.sheetClient()
	}

	// sheetName := getSheetName(c, spreadsheetId, sheetId)
	row := &sheets.ValueRange{
		Values: [][]interface{}{valueRange},
	}

	res, err := g.sheet.Spreadsheets.Values.Append(spreadsheetId, sheetName, row).Do()
	if err != nil || res.HTTPStatusCode != 200 {
		log.Fatalf("unable to append data into sheet %s range %s <%v>.", sheetName, cellRange, err)
	}

	return "", nil
}

func (g *Gcp) SpreadsheetUpdate(spreadsheetId string, sheetName string, cellRange string, valueRange []interface{}) (string, error) {
	ctx := context.Background()

	if g.sheet == nil {
		g.sheetClient()
	}
	// sheetName := getSheetName(c, spreadsheetId, sheetId)
	row := &sheets.ValueRange{
		Values: [][]interface{}{valueRange},
	}

	res, err := g.sheet.Spreadsheets.Values.Update(spreadsheetId, fmt.Sprintf("%s!%s", sheetName, cellRange), row).ValueInputOption("USER_ENTERED").Context(ctx).Do()
	if err != nil || res.HTTPStatusCode != 200 {
		log.Fatalf("unable to update data into sheet %s range %s <%v>.", sheetName, cellRange, err)
	}

	return "", nil
}

func (g *Gcp) sheetClient() {
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

// func getSheetName(c *sheets.Service, spreadsheetId string, sheetId int) string {
// 	res, err := c.Spreadsheets.Get(spreadsheetId).Fields("sheets(properties(sheetId,title))").Do()
// 	if err != nil || res.HTTPStatusCode != 200 {
// 		log.Println(err)
// 		return ""
// 	}

// 	sheetName := ""
// 	for _, v := range res.Sheets {
// 		prop := v.Properties
// 		if prop.SheetId == int64(sheetId) {
// 			sheetName = prop.Title
// 			break
// 		}
// 	}

// 	return sheetName
// }
