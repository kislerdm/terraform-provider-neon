// The application which generates the graph representing the distribution of the provider downloads in time.
package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"iter"
	"log"
	"net/http"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	var outPathRaw string
	flag.StringVar(&outPathRaw, "o", "/tmp", "dir to store the data.")
	flag.Parse()

	c, err := newCookies()
	if err != nil {
		log.Fatalf("could not init cookies:%v\n", err)
	}

	stats, err := fetchStatsWeb(c)
	if err != nil {
		log.Printf("could not fetch the data from terraform registry: %v\n", err)
		return
	}

	pOut := path.Join(outPathRaw, "data.txt")
	if err = os.WriteFile(pOut, []byte(stats), 0644); err != nil {
		log.Printf("could not save the raw data to %s: %v\n", pOut, err)
	}

	data, err := readData(stats)
	if err != nil {
		log.Printf("could not process the data: %v\n", err)
		return
	}
	sortData(data)

	dataJson, err := json.Marshal(toHtmlData(data))
	if err != nil {
		log.Printf("could not serialise processed data: %v\n", err)
		return
	}

	pOut = path.Join(outPathRaw, "data.json")
	if err = os.WriteFile(pOut, dataJson, 0644); err != nil {
		log.Printf("could not save projected data to %s: %v\n", pOut, err)
	}
}

func newCookies() (*cookies, error) {
	o := &cookies{
		Key: os.Getenv("TF_COOKIE_KEY"),
	}
	var err error
	if o.Key == "" {
		err = errors.Join(err, fmt.Errorf("env variable TF_COOKIE_KEY must be set"))
	}
	if err != nil {
		o = nil
	}
	return o, err
}

type cookies struct {
	// Key has to be updated on a monthly basis.
	Key string `cookie:"terraform-registry"`
}

func (c cookies) Next() iter.Seq[*http.Cookie] {
	val := reflect.ValueOf(c)
	return func(yield func(v *http.Cookie) bool) {
		for i := 0; i < val.NumField(); i++ {
			fType := val.Type().Field(i)
			v := &http.Cookie{
				Name: fType.Tag.Get("cookie"),
			}

			switch fType.Type.Kind() {
			case reflect.String:
				v.Value = val.Field(i).String()
			default:
				v.Value = fmt.Sprintf("%v", val.Field(i).Interface())
			}

			if !yield(v) {
				return
			}
		}
	}
}

func fetchStatsWeb(c *cookies) (o string, err error) {
	const url = "https://registry.terraform.io/v2/providers/3734/downloads"
	r, er := http.NewRequest(http.MethodGet, url, nil)
	if er != nil {
		err = fmt.Errorf("could not make request to %s: %v\n", url, er)
	}

	var resp *http.Response
	if err == nil {
		if c != nil {
			for cookie := range c.Next() {
				r.AddCookie(cookie)
			}
		}
		resp, er = http.DefaultClient.Do(r)
		if er != nil {
			err = fmt.Errorf("could not make request to %s: %v\n", url, er)
		}
	}

	if err == nil {
		defer func() { _ = resp.Body.Close() }()
		b, er := io.ReadAll(resp.Body)
		if er != nil {
			err = fmt.Errorf("could not read response from %s: %v\n", url, er)
		} else {
			o = string(b)
		}
	}

	return o, err
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

type htmlData struct {
	Dates    []string `json:"dates"`
	Versions []string `json:"versions"`
	// Downloads defines the number of downloads broken down by month and version
	// {2025-01: {0.7.1: 10, 0.7.0: 5}}
	Downloads map[string]map[string]int `json:"downloads"`
}

func toHtmlData(v []record) htmlData {
	o := htmlData{
		Dates: v[0].Date,
	}
	o.Downloads = make(map[string]map[string]int, len(o.Dates))
	for _, date := range o.Dates {
		o.Downloads[date] = make(map[string]int)
	}
	for _, el := range v {
		o.Versions = append(o.Versions, el.Version)
		for i, date := range el.Date {
			o.Downloads[date][el.Version] = el.Count[i]
		}
	}
	return o
}
