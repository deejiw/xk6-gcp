package gcp

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// This function generates a unique ID for a new row in a Google Sheet.
// Parameters:
// - rows: a slice of slices of interface{} representing the rows of a Google Sheet.
// Returns:
// - int64: a unique ID for a new row.
func getUniqueId(rows [][]interface{}) int64 {
	var id int64

	if len(rows) > 1 {
		lastID, err := strconv.ParseInt(rows[len(rows)-1][0].(string), 10, 64)
		if err != nil {
			log.Fatalf("unable to parse the last ID from last row as %s <%v>.", rows[len(rows)-1], err)
		}
		id = lastID + 1
	} else {
		id = 1
	}

	return id
}

// This function merges two slices of interface{} into a map[string]interface{}.
// Parameters:
// - keys: a slice of interface{} representing the keys of the map.
// - values: a slice of interface{} representing the values of the map.
// Returns:
// - map[string]interface{}: a map with keys and values from the input slices.
func mergeKV(keys []interface{}, values []interface{}) map[string]interface{} {
	mergedMap := make(map[string]interface{})
	// Check if the lengths of keys and values arrays are the same
	if len(keys) == len(values) {
		// Iterate over the keys and values arrays
		for i := 0; i < len(keys); i++ {
			key := keys[i]
			value := values[i]
			mergedMap[key.(string)] = value
		}
	} else {
		fmt.Println("error: length of keys and values arrays must be the same.")
	}

	return mergedMap
}

// This function converts a column index to its corresponding letter in a Google Sheet.
// Parameters:
// - index: an integer representing the column index.
// Returns:
// - string: a string representing the corresponding letter of the column index.
func columnIndexToLetter(index int) string {
	var result string

	for index >= 0 {
		result = string(index%26+65) + result
		index = index/26 - 1
	}

	return result
}

// This function finds the index of a header in a slice of headers.
// Parameters:
// - headers: a slice of interface{} representing the headers.
// - header: a string representing the header to find.
// Returns:
// - int: the index of the header in the slice, or -1 if not found.
func findHeaderIndex(headers []interface{}, header string) int {
	for i, h := range headers {
		if strings.TrimSpace(h.(string)) == header {
			return i
		}
	}
	return -1
}

// This function sorts values by headers.
// Parameters:
// - headers: a slice of interface{} representing the headers.
// - values: a map[string]interface{} representing the values to sort.
// Returns:
// - []interface{}: a slice of interface{} values sorted by the headers.
func sortValuesByHeaders(headers []interface{}, values map[string]interface{}) []interface{} {
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[header.(string)] = i
	}

	// Create the new row data in the correct order
	var sorted []interface{}
	for columnName, value := range values {
		index, found := headerMap[columnName]
		if !found {
			log.Fatalf("column '%s' not found in the sheet", columnName)
		}
		sorted = append(sorted, value)
		index++
		for len(sorted) < index {
			sorted = append(sorted, nil)
		}
	}

	return sorted
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
