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
	writeToFile("Year,Transaction Date,Posting Date,Description,Amount")

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

		for _, r := range rows {
			row := ""
			for _, word := range r.Content {
				row += fmt.Sprintf("%s ", word.S)
			}
			_startYear, _endYear := periodCoveredCheck(row)
			if _startYear != "" && _endYear != "" {
				startYear = _startYear
				endYear = _endYear
			}
		}

		for _, r := range rows {

			row := ""
			for _, word := range r.Content {
				row += fmt.Sprintf("%s ", word.S)
			}

			currentMonth := monthCheck(row)

			if currentMonth != "" {
				priceSplit := strings.Split(row, "$")

				if len(priceSplit) == 2 {

					transactionPrice := priceSplit[1]

					dateSplit := strings.Split(priceSplit[0], " ")

					transactionDate := fmt.Sprintf("%s %s", dateSplit[0], dateSplit[1])
					postingDate := fmt.Sprintf("%s %s", dateSplit[2], dateSplit[3])

					transactionTitle := ""

					for i := 4; i < len(dateSplit); i++ {
						transactionTitle += fmt.Sprintf("%s ", dateSplit[i])
					}

					currentYear := startYear

					if startYear != endYear && currentMonth == "DEC" {
						currentYear = startYear
					} else if startYear != endYear && currentMonth == "JAN" {
						currentYear = endYear
					}

					writeToFile(fmt.Sprintf("%s,%s,%s,%s,$%s", currentYear, transactionDate, postingDate, strings.TrimSpace(transactionTitle), transactionPrice))
				}
			}

		}
	}

	return "", nil
}

func monthCheck(row string) (currentMonth string) {
	months := [12]string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	_currentMonth := ""

	for _, month := range months {
		if strings.Contains(row, month) && !strings.Contains(row, "PAYMENT - THANK YOU -") {
			_currentMonth = month
		}
	}

	return _currentMonth
}

func periodCoveredCheck(row string) (startYear string, endYear string) {
	if strings.Contains(row, "Period Covered:") {
		dates := strings.Split(row, "Period Covered:")
		datesSplit := strings.Split(dates[1], "-")

		startdate := strings.TrimSpace(datesSplit[0])
		endDate := strings.TrimSpace(datesSplit[1])

		startYearSplit := strings.Split(startdate, ",")
		endYearSplit := strings.Split(endDate, ",")

		startYear := strings.TrimSpace(startYearSplit[1])
		endYear := strings.TrimSpace(endYearSplit[1])

		return startYear, endYear
	}
	return "", ""
}

func writeToFile(row string) {
	f, err := os.OpenFile("statements.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
