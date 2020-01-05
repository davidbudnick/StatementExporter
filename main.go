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
		fmt.Println(f.Name())
		content, err := readPdf(fmt.Sprintf("%s/%s", rootDir, f.Name()))
		if err != nil {
			panic(err)
		}
		fmt.Println(content)
	}

}

func readPdf(path string) (string, error) {
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
			if monthCheck(row) {

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

					writeToFile(fmt.Sprintf("%s,%s,%s,$%s", transactionDate, postingDate, strings.TrimSpace(transactionTitle), transactionPrice))
				}

			}
		}
	}
	return "", nil
}

func monthCheck(row string) bool {
	months := [12]string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	for _, month := range months {
		if strings.Contains(row, month) && !strings.Contains(row, "PAYMENT - THANK YOU -") {
			return true
		}
	}
	return false
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
