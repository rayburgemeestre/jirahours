package cmd

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rayburgemeestre/jirahours/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var (
	undoWorklogs string
)

func init() {
	undoCmd.Flags().StringVarP(&dates, "file", "f", "dates.txt", "file to read dates from (e.g. dates.txt)")
	undoCmd.Flags().StringVarP(&undoWorklogs, "delete", "o", "undo_worklogs.sh", "file to write undo worklog hours script to (e.g. undo_worklogs.sh)")
	rootCmd.AddCommand(undoCmd)
}

// HACK: this undo cmd is a copy & paste from syncback, and should be refactored to get rid of the huge code duplication.
var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Generate undo script for all worklog entries in given date range from Jira Tempo hours",
	Long:  `This script can be manually checked before executing and is generally useful if you somehow screwed up automated submission of worklog hours`,
	Run: func(cmd *cobra.Command, args []string) {
		min, max := util.GetMinMaxDatefile(dates)
		username := viper.GetString("jira_credentials.username")
		password := viper.GetString("jira_credentials.password")
		url := fmt.Sprintf(viper.GetString("jira_worklog_api.list"), min.UnixNano()/1000000)
		f, err := os.Create(undoWorklogs)
		util.CheckIfError(err)
		for url != "" {
			fmt.Println("Gathering data from URL:", url)

			req, err := http.NewRequest("GET", url, nil)
			req.SetBasicAuth(username, password)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			var dat map[string]interface{}
			if err := json.Unmarshal(b, &dat); err != nil {
				panic(err)
			}

			if dat["nextPage"] == nil {
				url = ""
			} else {
				url = dat["nextPage"].(string)
			}

			worklogIds := dat["values"].([]interface{})
			var ids []int
			for _, v := range worklogIds {
				m := v.(map[string]interface{})
				id := int(m["worklogId"].(float64))
				util.CheckIfError(err)
				ids = append(ids, id)
			}

			type Payload struct {
				Ids []int `json:"ids"`
			}

			pl := &Payload{ids}

			jsonStr, _ := json.Marshal(pl)

			req, err = http.NewRequest("POST", viper.GetString("jira_worklog_api.detail"), bytes.NewBuffer(jsonStr))
			req.SetBasicAuth(username, password)
			req.Header.Set("Content-Type", "application/json")

			client = &http.Client{}
			resp, err = client.Do(req)
			if err != nil {
				panic(err)
			}

			b, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			defer func() {
				err := resp.Body.Close()
				util.CheckIfError(err)
			}()

			var dat2 []map[string]interface{}
			if err := json.Unmarshal(b, &dat2); err != nil {
				panic(err)
			}

			for _, item := range dat2 {
				authordat := item["author"].(map[string]interface{})
				author := authordat["name"].(string)
				what := item["comment"].(string)
				issueId := item["issueId"].(string)
				worklogId := item["id"].(string)
				when := item["started"].(string)
				whentime, err := time.Parse("2006-01-02", when[0:10])
				if min.Before(whentime) && max.After(whentime) {
					// date within range
				} else {
					continue // skip
				}
				if author != username {
					continue
				}
				fmt.Println(item)
				_, err = f.WriteString(fmt.Sprintf("jirahours delete -i %s -w %s # %s\n", issueId, worklogId, what))
				util.CheckIfError(err)
			}
		}
	},
}
