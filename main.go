package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/dcu/pdf"
)

//LineItem Information
type LineItem struct {
	transactionMonth string
	transactionDay   string
	postingMonth     string
	postingDay       string
	description      string
	amount           string
}

//StatementPeriods dates
type StatementPeriods struct {
	startYear string
	endYear   string
}

func main() {
	writeToFile("Transaction Date,Posting Date,Description,Amount")

	rootDir := "./statements"
	files, err := ioutil.ReadDir(rootDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		_, err := readPdf(fmt.Sprintf("%s/%s", rootDir, f.Name()))
		if err != nil {
			panic(err)
		}
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

			periods := getPeriods(row)
			lineItem := getLineItem(row)

			if (StatementPeriods{} != periods) {
				startYear = periods.startYear
				endYear = periods.endYear
			}

			if (LineItem{} == lineItem) {
				continue
			}
			currentTransactionYear := getCurrentYear(lineItem.transactionMonth, startYear, endYear)
			currentPostingYear := getCurrentYear(lineItem.postingMonth, startYear, endYear)

			//Removes statement payments from transactions
			if strings.Contains(lineItem.description, "PAYMENT - THANK YOU -") {
				continue
			}

			//Accounting for CREDIT charges on statments
			if strings.Contains(lineItem.description, "CREDIT") {
				lineItem.amount = fmt.Sprintf("-%s", lineItem.amount)
				lineItem.description = strings.Replace(lineItem.description, "-", "", -1)
			}

			writeToFile(fmt.Sprintf("%s %s %s,%s %s %s,%s,%s",
				lineItem.transactionMonth,
				lineItem.transactionDay,
				currentTransactionYear,
				lineItem.postingMonth,
				lineItem.postingDay,
				currentPostingYear,
				strings.TrimSpace(lineItem.description),
				lineItem.amount,
			))
		}
	}

	return "", nil
}

var preriodsRegex = regexp.MustCompile(`Period Covered: [a-zA-Z]* [0-9]{2}, (?P<startYear>\d{4}?) \- [a-zA-Z]* [0-9]{2}, (?P<endYear>\d{4}?)`)

//Gets the statement period range
func getPeriods(row string) (periods StatementPeriods) {
	groupItems := getGroupNames(row, *preriodsRegex)
	periods = StatementPeriods{
		startYear: groupItems["startYear"],
		endYear:   groupItems["endYear"],
	}

	return
}

var lineItemRegex = regexp.MustCompile(`(?P<transactionMonth>[a-zA-Z]{3}?)\s(?P<transactionDay>[0-9]{2}?)\s(?P<postingMonth>[A-Z]{3})\s(?P<postingDay>[0-9]{2}?)\s(?P<description>.*\s)(?P<amount>\$[0-9]*.[0-9]{2})`)

//Gets the line items by running regex on the line item
func getLineItem(row string) (lineItem LineItem) {
	groupItems := getGroupNames(row, *lineItemRegex)
	lineItem = LineItem{
		transactionMonth: groupItems["transactionMonth"],
		transactionDay:   groupItems["transactionDay"],
		postingMonth:     groupItems["postingMonth"],
		postingDay:       groupItems["postingDay"],
		description:      groupItems["description"],
		amount:           groupItems["amount"],
	}

	return
}

//Get the current year of the line item (Transaction Date and Posting Date)
func getCurrentYear(transactionMonth string, startYear string, endYear string) (currentYear string) {
	currentYear = startYear

	if startYear != endYear && transactionMonth == "DEC" {
		currentYear = startYear
	} else if startYear != endYear && transactionMonth == "JAN" {
		currentYear = endYear
	}
	return
}

//Get the group names from the regex statements
func getGroupNames(row string, regex regexp.Regexp) (groupItems map[string]string) {
	groupItems = make(map[string]string)
	groupNames := regex.SubexpNames()

	for _, match := range regex.FindAllStringSubmatch(row, -1) {
		for groupIdx, group := range match {
			name := groupNames[groupIdx]
			if name != "" {
				groupItems[name] = group
			}
		}
	}
	return
}

//Writes line items to CSV document
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
}
