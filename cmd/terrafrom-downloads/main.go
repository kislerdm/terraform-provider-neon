// The application which generates the graph representing the distribution of the provider downloads in time.
package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

func main() {
	var outPath string
	flag.StringVar(&outPath, "o", "public/index.html", "path to store generated HTML page.")
	flag.Parse()

	stats, err := fetchStats()
	if err != nil {
		log.Fatalf("could not fetch the data from terraform registry: %v\n", err)
	}

	data, err := readData(stats)
	if err != nil {
		log.Fatalf("could not process the data: %v\n", err)
	}
	sortData(data)

	fHTML, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("could not open file for saving HTML page %s: %v\n", outPath, err)
	}
	defer func() { _ = fHTML.Close() }()

	dataJson, err := json.Marshal(data)
	if err != nil {
		log.Printf("could not serialise processed data: %v\n", err)
		return
	}
	// html/template would require additional transformation of data,
	// hence rely on text/template because no html escape is required.
	if err = templateGen.Execute(fHTML, string(dataJson)); err != nil {
		log.Printf("could not generate HTML page: %v\n", err)
		return
	}
}

//go:embed stats.txt
var statsData string

func fetchStats() (string, error) {
	return statsData, nil
}

func sortData(v []record) {
	for i, el := range v {
		sort.Sort(el)
		v[i] = el
	}
}

type record struct {
	Date    []string `json:"date"`
	Count   []int    `json:"count"`
	Version string   `json:"version"`
}

func (r record) Len() int {
	return len(r.Date)
}

func (r record) Less(i, j int) bool {
	iT, _ := time.Parse("2006-01", r.Date[i])
	jT, _ := time.Parse("2006-01", r.Date[j])
	return jT.After(iT)
}

func (r record) Swap(i, j int) {
	r.Date[i], r.Date[j] = r.Date[j], r.Date[i]
	r.Count[i], r.Count[j] = r.Count[j], r.Count[i]
}

func readData(v string) (data []record, err error) {
	var version []string
	var countsByYearByMonthByVersionIndex = make(map[string]map[string][]int)
	var countsByMonthByVersionIndex = make(map[string][]int, 12)

	for iRow, row := range strings.Split(v, "\n") {
		els := strings.Split(row, ",")
		if len(els) > 1 {
			switch {
			case els[0] == "Date\\Version":
				for _, el := range els[1:] {
					version = append(version, strings.TrimSpace(el))
				}
			case isMonth(els[0]):
				for iCol, el := range els[1:] {
					if vv, er := strconv.Atoi(el); er != nil {
						err = errors.Join(err,
							fmt.Errorf("could count for row %d col %d': %w", iRow, iCol+1, er),
						)
					} else {
						countsByMonthByVersionIndex[els[0]] = append(countsByMonthByVersionIndex[els[0]], vv)
					}
				}
			case isYear(els[0]):
				countsByYearByMonthByVersionIndex[els[0]] = countsByMonthByVersionIndex
				countsByMonthByVersionIndex = make(map[string][]int, 12)
			}
		}
	}

	data = make([]record, len(version))
	for year, countsByMonthByVersionIndexVal := range countsByYearByMonthByVersionIndex {
		for monthName, countByVersionIndex := range countsByMonthByVersionIndexVal {
			for versionIndex, count := range countByVersionIndex {
				date, er := time.Parse("2006-January", fmt.Sprintf("%s-%s", year, monthName))
				if err != nil {
					err = errors.Join(err, fmt.Errorf("error parsing date monthName= %s, year= %s: %w",
						monthName, year, er))
				} else {
					data[versionIndex].Date = append(data[versionIndex].Date, date.Format("2006-01"))
					data[versionIndex].Count = append(data[versionIndex].Count, count)
					data[versionIndex].Version = version[versionIndex]
				}
			}
		}
	}

	return data, err
}

func isYear(s string) bool {
	return strings.HasPrefix(s, "20")
}

func isMonth(s string) bool {
	var months = map[string]struct{}{
		"January": {}, "February": {}, "March": {},
		"April": {}, "May": {}, "June": {},
		"July": {}, "August": {}, "September": {},
		"October": {}, "November": {}, "December": {},
	}
	var ok bool
	_, ok = months[s]
	return ok
}

//go:embed index.html.templ
var templateHTML string

var templateGen = template.Must(template.New("page").Parse(templateHTML))
