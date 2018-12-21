// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package util

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetMinMaxDatefile(filename string) (min time.Time, max time.Time) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	var datesAll []time.Time
	scanner := bufio.NewScanner(file)
	min = time.Date(3000, 0, 0, 0, 0, 0, 0, time.Local)
	max = time.Date(1000, 0, 0, 0, 0, 0, 0, time.Local)
	for scanner.Scan() {
		line := scanner.Text()
		commented := false
		if strings.Contains(line, ";") {
			commented = true
			line = strings.TrimSpace(line[1:])
		}
		if commented {

		}
		fields := strings.Split(line, "-")
		if len(fields) < 3 {
			fmt.Println(fields)
			panic("Not enough fields on line found")
		}
		year, err := strconv.Atoi(fields[0])
		CheckIfError(err)
		month, err := strconv.Atoi(fields[1])
		CheckIfError(err)
		day, err := strconv.Atoi(fields[2])
		CheckIfError(err)

		var t time.Time
		if t = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local); t.Before(min) {
			min = t
		}
		if t := time.Date(year, time.Month(month), day+1, 0, 0, 0, 0, time.Local); t.After(max) {
			max = t
		}
		if !commented {
			datesAll = append(datesAll, t)
		}
	}
	return
}
