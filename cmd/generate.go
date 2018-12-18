package cmd

import (
	"../util"
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	issues       string
	outputScript string
)

type DateLogs struct {
	date       time.Time
	hours      string
	key        string
	msg        string
	hourminute string // hh:mm
	minutes    int
}

type Issue struct {
	key string
	msg string
}

func init() {
	generateCmd.Flags().StringVarP(&dates, "file", "f", "dates.txt", "file to read dates from (e.g. dates.txt)")
	generateCmd.Flags().StringVarP(&issues, "in", "i", "issues.txt", "file to read commit entries from (e.g. issues.txt)")
	generateCmd.Flags().StringVarP(&outputScript, "out", "o", "submit_hours.sh", "file to write bash script (e.g. submit_hours.sh)")
	rootCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a bash script to submit jira hours",
	Long:  `The reason we generate a bash script is so you can manually inspect before screwing up your Tempo hours`,
	Run: func(cmd *cobra.Command, args []string) {

		datesAll := readDates(dates)

		issuesAll := readIssues(issues)

		worklogs := generateWorklogs(datesAll, issuesAll)

		writeWorklogSubmitScript(worklogs)
	},
}

func writeWorklogSubmitScript(worklogs [][]DateLogs) {
	f, err := os.Create(outputScript)
	util.CheckIfError(err)
	defer func() {
		err := f.Close()
		util.CheckIfError(err)
	}()
	for _, worklog := range worklogs {
		n, err := f.WriteString("\n")
		util.CheckIfError(err)
		fmt.Println("Written", n, "bytes")

		for i := 0; i < len(worklog); i++ {
			log := worklog[i]
			log.key = strings.Trim(log.key, " ")
			log.msg = strings.Trim(log.msg, " ")
			n, err := f.WriteString(
				fmt.Sprintf(`jirahours submit -s "%04d-%02d-%02d" -w "%s" -j "%s" -m "%s"%c`,
					log.date.Year(),
					log.date.Month(),
					log.date.Day(),
					log.hourminute,
					log.key,
					log.msg,
					'\n',
				))
			fmt.Println("Written", n, "bytes")
			util.CheckIfError(err)
		}
	}
}

func generateWorklogs(datesAll []time.Time, issuesAll []Issue) (worklogs [][]DateLogs) {
	worklogs = [][]DateLogs{}

	totalMinutes := len(datesAll) * viper.GetInt("log_hours_per_day") * 60
	totalIssues := len(issuesAll)
	minutesPerIssue := totalMinutes / totalIssues
	currentIssueMinutes := minutesPerIssue
	issueIndex := 0

	for i := 0; i < len(datesAll); i++ {
		date := datesAll[i]
		logsForDate := []DateLogs{}
		logged := 0
		logPerDay := viper.GetInt("log_hours_per_day") * 60

		for toLog := logPerDay; toLog > 0; {
			if issueIndex >= len(issuesAll) {
				break
			}
			issue := issuesAll[issueIndex]
			// enough available from current issue to fulfill this dates log
			if currentIssueMinutes >= toLog {
				hours, key, msg := logTime(issue, toLog)
				logsForDate = append(logsForDate, DateLogs{
					date,
					hours,
					key,
					msg,
					"",
					0,
				})
				currentIssueMinutes -= toLog
				logged += toLog
				toLog = 0
			} else { // else: log what we have left
				hours, key, msg := logTime(issue, currentIssueMinutes)
				logsForDate = append(logsForDate, DateLogs{
					date,
					hours,
					key,
					msg,
					"",
					0,
				})
				toLog -= currentIssueMinutes
				logged += currentIssueMinutes
				currentIssueMinutes = 0
			}
			// nothing left for this issue, advance to next one
			if currentIssueMinutes <= 0 {
				issueIndex++
				currentIssueMinutes = minutesPerIssue
			}
		}

		// jira seems to round to 15 minute blocks, so most of the time,
		//  logs don't add up to exactly 8 hours.. so we do a second pass fixing that
		totalMinutes = 0
		newLogsForDate := []DateLogs{}
		for i := 0; i < len(logsForDate); i++ {
			log := logsForDate[i]
			time := strings.Split(log.hours, ":")
			h, err := strconv.Atoi(time[0])
			util.CheckIfError(err)
			m, err := strconv.Atoi(time[1])
			util.CheckIfError(err)
			minutesRounded := int(math.Round((float64(h*60.0)+float64(m))/15.0) * 15)
			totalMinutes += minutesRounded
			log.hourminute = minutesToHours(minutesRounded)
			log.minutes = minutesRounded
			if minutesRounded == 0 {
				continue
			}
			newLogsForDate = append(newLogsForDate, log)
		}
		logsForDate = newLogsForDate

		loggedLess := logPerDay - totalMinutes // minutes we logged less (can be negative)
		if loggedLess != 0 {
			// fix by modifying another log entry (that doesn't end up with a negative value)
			for i := 0; i < len(logsForDate); i++ {
				log := logsForDate[i]
				correctedValue := log.minutes + loggedLess
				if correctedValue < 0 {
					continue
				}
				log.minutes = correctedValue
				log.hourminute = minutesToHours(log.minutes)
				logsForDate[i] = log
				break
			}
		}
		worklogs = append(worklogs, logsForDate)
	}
	return
}

func readDates(filename string) (dates []time.Time) {
	file, err := os.Open(filename)
	util.CheckIfError(err)
	defer func() {
		err := file.Close()
		util.CheckIfError(err)
	}()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, ";") {
			continue
		}
		fields := strings.Split(line, "-")
		if len(fields) < 3 {
			panic("Not enough fields on line found")
		}
		year, err := strconv.Atoi(fields[0])
		util.CheckIfError(err)
		month, err := strconv.Atoi(fields[1])
		util.CheckIfError(err)
		day, err := strconv.Atoi(fields[2])
		util.CheckIfError(err)

		t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
		dates = append(dates, t)
	}
	return
}

func readIssues(filename string) (issues []Issue) {
	file, err := os.Open(filename)
	util.CheckIfError(err)
	defer func() {
		err := file.Close()
		util.CheckIfError(err)
	}()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, "***")
		if len(fields) != 4 {
			panic("There should be four fields in issues.txt per line.")
		}
		issues = append(issues, Issue{key: fields[1], msg: fields[2]})
	}
	return
}

func minutesToHours(minutes int) string {
	hours := int(minutes / 60.0)
	minutes = minutes - (hours * 60)
	return fmt.Sprintf("%02d:%02d", hours, minutes)
}

func logTime(issue Issue, minutes int) (hours string, key string, msg string) {
	hours = minutesToHours(minutes)
	msg = strings.Replace(issue.msg, "\"", "'", -1)
	key = issue.key
	return
}
