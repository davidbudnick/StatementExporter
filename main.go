package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/dcu/pdf"
)

func main() {
	writeToFile("Transaction Date,Posting Date,Description,Amount")

	rootDir := "./statements"
	files, err := ioutil.ReadDir(rootDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		content, err := readPdf(fmt.Sprintf("%s/%s", rootDir, f.Name()))
		if err != nil {
			panic(err)
		}
		fmt.Println(content)
	}

}

func readPdf(path string) (string, error) {

	var startYear string
	var endYear string

	f, r, err := pdf.Open(path)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return "", err
	}
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		rows, _ := p.GetTextByRow()

		startYear, endYear = getYear(rows)

		for _, r := range rows {

			row := ""
			for _, word := range r.Content {
				row += fmt.Sprintf("%s ", word.S)
			}

			currentMonth := getMonth(row)

			if currentMonth != "" && startYear != "" && endYear != "" {

				//Split price out of the line item
				priceSplit := strings.Split(row, "$")
				if len(priceSplit) != 2 {
					return "", nil
				}
				transactionPrice := priceSplit[1]

				//Split on space so I can get the dates
				dateSplit := strings.Split(priceSplit[0], " ")
				transactionDate := fmt.Sprintf("%s %s", dateSplit[0], dateSplit[1])
				postingDate := fmt.Sprintf("%s %s", dateSplit[2], dateSplit[3])

				//The rest of the string inclues the line item description
				transactionTitle := ""
				for i := 4; i < len(dateSplit); i++ {
					transactionTitle += fmt.Sprintf("%s ", dateSplit[i])
				}

				//Logic for setting the current year of the statment
				//Taking the start of the stament date and the end of the statmenet to check if it is going into the next year
				currentYear := startYear
				if startYear != endYear && currentMonth == "DEC" {
					currentYear = startYear
				} else if startYear != endYear && currentMonth == "JAN" {
					currentYear = endYear
				}

				//Writes all the transactions
				writeToFile(fmt.Sprintf("%s %s,%s %s,%s,$%s", transactionDate, currentYear, postingDate, currentYear, strings.TrimSpace(transactionTitle), transactionPrice))
			}
		}
	}

	return "", nil
}

func getMonth(row string) string {
	months := [12]string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}

	for _, month := range months {
		if strings.Contains(row, month) && !strings.Contains(row, "PAYMENT - THANK YOU -") {
			return month
		}
	}

	return ""
}

func getYear(rows pdf.Rows) (startYear string, endYear string) {

	for _, r := range rows {
		row := ""
		for _, word := range r.Content {
			row += fmt.Sprintf("%s ", word.S)
		}
		if strings.Contains(row, "Period Covered:") {
			dates := strings.Split(row, "Period Covered:")
			datesSplit := strings.Split(dates[1], "-")

			startdate := strings.TrimSpace(datesSplit[0])
			endDate := strings.TrimSpace(datesSplit[1])

			startYearSplit := strings.Split(startdate, ",")
			endYearSplit := strings.Split(endDate, ",")

			startYear := strings.TrimSpace(startYearSplit[1])
			endYear := strings.TrimSpace(endYearSplit[1])
			if startYear != "" && endYear != "" {
				return startYear, endYear
			}
		}
	}

	return "", ""
}

func writeToFile(row string) {
	f, err := os.OpenFile("transactions.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(fmt.Sprintf("%s\n", row))); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	f.Close()
	if err != nil {
		panic(err)
	}
}
