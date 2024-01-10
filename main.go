package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	links := []string{
		"https://en.wikipedia.org/wiki/Mobile_Network_Codes_in_ITU_region_2xx_(Europe)",
		"https://en.wikipedia.org/wiki/Mobile_Network_Codes_in_ITU_region_3xx_(North_America)",
		"https://en.wikipedia.org/wiki/Mobile_Network_Codes_in_ITU_region_4xx_(Asia)",
		"https://en.wikipedia.org/wiki/Mobile_Network_Codes_in_ITU_region_5xx_(Oceania)",
		"https://en.wikipedia.org/wiki/Mobile_Network_Codes_in_ITU_region_6xx_(Africa)",
		"https://en.wikipedia.org/wiki/Mobile_Network_Codes_in_ITU_region_7xx_(South_America)",
		"https://en.wikipedia.org/wiki/Mobile_country_code",
	}

	data := make(map[string]string)

	for _, link := range links {
		response, err := http.Get(link)
		if err != nil {
			fmt.Println("Failed to fetch page:", err)
			continue
		}
		defer response.Body.Close()

		doc, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			fmt.Println("Failed to parse HTML:", err)
			continue
		}

		doc.Find("table.wikitable").Each(func(i int, table *goquery.Selection) {
			columns := table.Find("tr").First().Find("th")
			if strings.TrimSpace(columns.Eq(0).Text()) != "MCC" {
				return // Skip table without MCC column
			}
			if strings.TrimSpace(columns.Eq(1).Text()) != "MNC" {
				return // Skip table without MNC column
			}
			if strings.TrimSpace(columns.Eq(3).Text()) != "Operator" {
				return // Skip table without Operator column
			}
			table.Find("tr").Each(func(j int, row *goquery.Selection) {
				if j == 0 {
					return // Skip header row
				}

				columns := row.Find("td")
				if columns.Length() >= 4 {
					mcc := strings.TrimSpace(columns.Eq(0).Text())
					if strings.Contains(mcc, "\n") {
						mcc = strings.Split(mcc, "\n")[1]
					}
					if _, err := strconv.ParseInt(strings.TrimSpace(mcc), 10, 64); err != nil {
						return // Skip row without MNC
					}
					mnc := strings.TrimSpace(columns.Eq(1).Text())
					brand := strings.TrimSpace(columns.Eq(2).Text())
					operator := strings.TrimSpace(columns.Eq(3).Text())
					if operator == "" {
						operator = brand
					}
					operator = strings.ReplaceAll(operator, "“", "\"")
					operator = strings.ReplaceAll(operator, "’", "'")
					operator = strings.ReplaceAll(operator, " ", " ")
					operator = strings.ReplaceAll(operator, "–", "-")

					if strings.Contains(mnc, "-") {
						mncRange := strings.Split(mnc, "-")
						mncStart, err := strconv.ParseInt(strings.TrimSpace(mncRange[0]), 10, 64)
						if err != nil {
							fmt.Println("Failed to parse MNC:", err)
							return
						}
						mncEnd, err := strconv.ParseInt(strings.TrimSpace(mncRange[1]), 10, 64)
						if err != nil {
							fmt.Println("Failed to parse MNC:", err)
							return
						}
						for mnc := mncStart; mnc <= mncEnd; mnc++ {
							mccmnc := mcc + strconv.FormatInt(mnc, 10)
							data[mccmnc] = operator
						}
					} else {
						if _, err := strconv.ParseInt(strings.TrimSpace(mnc), 10, 64); err != nil {
							return // Skip row without MNC
						}
						mccmnc := mcc + mnc
						data[mccmnc] = operator
					}
				}
			})
		})
	}

	// Convert data to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Failed to marshal data to JSON:", err)
		return
	}

	jsonData = bytes.ReplaceAll(jsonData, []byte("\\u0026"), []byte("&"))
	jsonData = bytes.ReplaceAll(jsonData, []byte("\\u003c"), []byte("<"))
	jsonData = bytes.ReplaceAll(jsonData, []byte("\\u003e"), []byte(">"))

	// Save JSON data to a file
	err = os.WriteFile("operators.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Failed to save JSON file:", err)
		return
	}

	fmt.Println("Data has been successfully saved to operators.json file")
}
