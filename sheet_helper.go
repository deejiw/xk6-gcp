package gcp

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

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

func columnIndexToLetter(index int) string {
	var result string

	for index >= 0 {
		result = string(index%26+65) + result
		index = index/26 - 1
	}

	return result
}

func findHeaderIndex(headers []interface{}, header string) int {
	for i, h := range headers {
		if strings.TrimSpace(h.(string)) == header {
			return i
		}
	}
	return -1
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
