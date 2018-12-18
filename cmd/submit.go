package cmd

import (
	"../util"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var (
	submitDate string
	worklog    string
	jiraKey    string
	message    string
)

func init() {
	submitCmd.SetArgs([]string{"arg1"})
	submitCmd.Flags().StringVarP(&submitDate, "submit-date", "s", "", "submit date (e.g. 2018-01-01)")
	submitCmd.Flags().StringVarP(&worklog, "worklog", "w", "", "worklog (e.g. 2:30 (for 2 hours and 30 minutes))")
	submitCmd.Flags().StringVarP(&jiraKey, "jira-key", "j", "", "jira key (e.g. CM-12345)")
	submitCmd.Flags().StringVarP(&message, "message", "m", "", "worklog message (e.g. I did such and such..)")
	rootCmd.AddCommand(submitCmd)
}

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit hours to jira based on credentials specified in your config.",
	Long:  `Submit hours to jira based on credentials specified in your config.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: No input checking
		fmt.Println("submit date:", submitDate)
		fmt.Println("worklog:", worklog)
		fmt.Println("jira key:", jiraKey)
		fmt.Println("message:", message)

		// Transform for Tempo hours
		submitDate = fmt.Sprintf("%sT11:11:11.111+0200", submitDate)
		f := strings.Split(worklog, ":")
		if len(f) != 2 {
			panic("Worklog not specified correctly.")
		}
		hours, err := strconv.Atoi(f[0])
		util.CheckIfError(err)
		minutes, err := strconv.Atoi(f[0])
		util.CheckIfError(err)
		minutes -= (minutes % 15)

		tss := (hours * 60 * 60) + minutes*60

		type Payload struct {
			TimeSpentSeconds int    `json:"timeSpentSeconds"`
			Started          string `json:"started"`
			Comment          string `json:"comment"`
		}

		pl := &Payload{tss, submitDate, message}
		fmt.Println(pl)

		fmt.Println("started:", submitDate)
		fmt.Println("timeSpentSeconds:", tss)
		fmt.Println("issue:", jiraKey)
		fmt.Println("comment:", message)

		username := viper.GetString("jira_credentials.username")
		password := viper.GetString("jira_credentials.password")

		fmt.Println(username, password)

		jsonStr, _ := json.Marshal(pl)

		url := fmt.Sprintf("https://jira.brightcomputing.com:8443/rest/api/latest/issue/%s/worklog", jiraKey)

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		req.SetBasicAuth(username, password)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer func() {
			err := resp.Body.Close()
			util.CheckIfError(err)
		}()

		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("response Body:", string(body))
	},
}
