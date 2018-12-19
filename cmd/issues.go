package cmd

import (
	"bufio"
	"fmt"
	"github.com/rayburgemeestre/jirahours/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Entry struct {
	t       time.Time
	jirakey string
	message string
	repo    string
}

type EntrySet struct {
	entries []Entry
	repo    string
}

var (
	wg      sync.WaitGroup
	channel chan EntrySet
	dates   string
	output  string
)

func init() {
	issuesCmd.Flags().StringVarP(&dates, "file", "f", "dates.txt", "file to read dates from (e.g. dates.txt)")
	issuesCmd.Flags().StringVarP(&output, "out", "o", "issues.txt", "file to write commit entries to (e.g. issues.txt)")
	rootCmd.AddCommand(issuesCmd)
}

var issuesCmd = &cobra.Command{
	Use:   "issues",
	Short: "Read in a dates file and gather all relevant git commit messages for the min and max date found in this file.",
	Long:  `This will write the result to an output file so it can be manually edited before proceding to the generate step.`,
	Run: func(cmd *cobra.Command, args []string) {
		file, err := os.Open(dates)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		var datesAll []time.Time
		scanner := bufio.NewScanner(file)
		min := time.Date(3000, 0, 0, 0, 0, 0, 0, time.Local)
		max := time.Date(1000, 0, 0, 0, 0, 0, 0, time.Local)
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
			util.CheckIfError(err)
			month, err := strconv.Atoi(fields[1])
			util.CheckIfError(err)
			day, err := strconv.Atoi(fields[2])
			util.CheckIfError(err)

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

		repos := viper.GetStringSlice("repositories")
		fmt.Println("Reading", len(repos), "repositories.")

		var issuesAll []Entry
		channel = make(chan EntrySet)
		wg = sync.WaitGroup{}

		seen := map[string]bool{}

		wg.Add(1)
		go func() {
			for range repos {
				msg1 := <-channel
				fmt.Println("Received", len(msg1.entries), "from repo", msg1.repo)
				if len(msg1.entries) > 0 {
					for _, m := range msg1.entries {
						if _, seen := seen[m.message]; seen {
							continue
						}
						seen[m.message] = true
						issuesAll = append(issuesAll, m)
					}
				}
			}
			wg.Done()
		}()

		for _, path := range repos {
			wg.Add(1)
			go process(path, min, max)
		}
		wg.Wait()

		sort.Slice(issuesAll, func(i, j int) bool {
			return issuesAll[i].t.Before(issuesAll[j].t.Local())
		})

		{
			f, err := os.Create(output)
			util.CheckIfError(err)
			defer f.Close()

			for _, issue := range issuesAll {
				_, err := f.WriteString(fmt.Sprintln(issue.t.String(), "***", issue.jirakey, "***", issue.message, "***", issue.repo))
				util.CheckIfError(err)
			}
			f.Sync()
		}
		fmt.Println("Generated", output)
	},
}

func process(path string, from time.Time, until time.Time) {
	r, err := git.PlainOpen(path)
	if err != nil {
		fmt.Println("Error with repository:", path)
	}
	util.CheckIfError(err)

	ref, err := r.Head()
	util.CheckIfError(err)

	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	util.CheckIfError(err)

	var issues []Entry
	chunks := strings.Split(path, "/")
	reponame := chunks[len(chunks)-1]

	u := viper.GetString("regexes.user")
	u = strings.TrimRight(u, "\n\r \t")
	reUser := regexp.MustCompile(u)

	c := viper.GetString("regexes.commits")
	c = strings.TrimRight(c, "\n\r \t")
	reCommit := regexp.MustCompile(c)

	err = cIter.ForEach(func(c *object.Commit) error {
		if reUser.MatchString(c.Author.Name) {
			if c.Author.When.After(from) && c.Author.When.Before(until) {
				commitMessageLines := strings.Split(c.Message, "\n")
				for _, line := range commitMessageLines {
					result := make(map[string]string)
					match := reCommit.FindStringSubmatch(line)
					if len(match) <= 2 {
						continue
					}
					for i, name := range reCommit.SubexpNames() {
						if i != 0 && name != "" {
							result[name] = match[i]
						}
					}
					e := Entry{
						t:       c.Author.When.Local(),
						jirakey: result["key"],
						message: result["message"],
						repo:    reponame,
					}
					issues = append(issues, e)
				}
			}
		}
		return nil
	})
	util.CheckIfError(err)
	channel <- EntrySet{issues, reponame}
	wg.Done()
}
