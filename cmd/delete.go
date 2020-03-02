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
	"net/http"
)

var (
	ticketId  string
	worklogId string
)

func init() {
	deleteCmd.Flags().StringVarP(&ticketId, "jira-key", "i", "", "ticket id")
	deleteCmd.Flags().StringVarP(&worklogId, "worklog-id", "w", "", "worklog id")
	rootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete worklog entry for jira issue.",
	Long:  `Delete worklog entry for jira issue.`,
	Run: func(cmd *cobra.Command, args []string) {
		type Payload struct {
		}
		pl := &Payload{}
		username := viper.GetString("jira_credentials.username")
		password := viper.GetString("jira_credentials.password")
		jsonStr, _ := json.Marshal(pl)
		url := fmt.Sprintf(viper.GetString("jira_worklog_api.delete_worklog"), ticketId, worklogId)
		req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonStr))
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
		fmt.Println("Delete worklog in Jira:", string(jsonStr), "Response:", resp.Status)
	},
}
