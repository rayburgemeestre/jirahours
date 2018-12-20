package cmd

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
)

var (
	outputWorklogs string
)

func init() {
	syncbackCmd.Flags().StringVarP(&dates, "file", "f", "dates.txt", "file to read dates from (e.g. dates.txt)")
	syncbackCmd.Flags().StringVarP(&outputWorklogs, "out", "o", "existing_tempo_hours.txt", "file to write existing worklog hours to (e.g. existing_tempo_hours.txt)")
	rootCmd.AddCommand(syncbackCmd)
}

var syncbackCmd = &cobra.Command{
	Use:   "syncback",
	Short: "Sync back already existing worklogs for given date range from Jira Tempo hours",
	Long:  `This data can be used to submit "around" these existing log entries (like Meetings you logged manually)`,
	Run: func(cmd *cobra.Command, args []string) {
		min, _ := util.GetMinMaxDatefile(dates)
		username := viper.GetString("jira_credentials.username")
		password := viper.GetString("jira_credentials.password")
		url := fmt.Sprintf(viper.GetString("jira_worklog_api.list"), min.UnixNano()/1000000)

		f, err := os.Create(outputWorklogs)
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
				when := item["started"].(string)
				what := item["comment"].(string)
				timeSpentSeconds := item["timeSpentSeconds"].(float64)
				if author != username {
					continue
				}
				_, err := f.WriteString(fmt.Sprintln(when, "***", author, "***", timeSpentSeconds, "***", what))
				util.CheckIfError(err)
			}
		}
	},
}
