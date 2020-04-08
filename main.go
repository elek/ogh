package main

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/skratchdot/open-golang/open"
	"github.com/urfave/cli"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"
)

var version string
var commit string
var date string

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	if len(os.Args) == 2 {
		_, err := strconv.Atoi(os.Args[1])
		if err == nil {
			err := open.Start("http://github.com/apache/hadoop-ozone/pull/" + os.Args[1])
			if err != nil {
				panic(err)
			}
			return
		}
	}

	app := cli.NewApp()
	app.Name = "ogh"
	app.Description = "Ozone Github Development helper"
	app.Version = fmt.Sprintf("%s (%s, %s)", version, commit, date)
	app.Commands = []cli.Command{
		{
			Name:    "review",
			Aliases: []string{"r"},
			Usage:   "Show the review queue (all READY pull requests)",
			Action: func(c *cli.Context) error {
				return run(false, "")
			},
		},
		{
			Name:    "pull-requests",
			Aliases: []string{"pr"},
			Usage:   "Show all the available pull requests",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "user",
					Usage: "Github user or organization name",
					Value: "",
				},
			},
			Action: func(c *cli.Context) error {
				return run(true, c.String("user"))
			},
		},
		{
			Name:    "mine",
			Aliases: []string{"pr"},
			Usage:   "Show results of the pr from the current user",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "user",
					Usage: "Github user name. (Default: current user)",
					Value: "",
				},
			},
			Action: func(c *cli.Context) error {
				userName := c.String("user")
				if userName == "" {
					user, err := user.Current()
					if err == nil {
						userName = user.Username
					}
				}
				return run(true, userName)
			},
		},
		{
			Name:  "artifacts",
			Usage: "Download build arfitacts",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "user",
					Usage: "Github user or organization name",
					Value: "apache",
				},
				cli.StringFlag{
					Name:  "dir",
					Usage: "Destination dir to save the downloaded artifacts",
					Value: "/tmp",
				},
				cli.BoolFlag{
					Name:  "all",
					Usage: "If not used, only the failed artifacts will be downloaded.",
				},
			},
			Action: func(c *cli.Context) error {
				return downloadArtifacts(c.String("user"), c.Args().Get(0), c.String("dir"), c.Bool("all"))
			},
		},
		{
			Name:    "builds",
			Aliases: []string{"b"},
			Usage:   "Print results of branch builds.",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "user",
					Usage: "Github user or organization name",
					Value: "apache",
				},
			},
			Action: func(c *cli.Context) error {
				return listBuilds(c.String("user"), "master", 8247)
			},
		},
		{
			Name:      "archive",
			Usage:     "Save artifacts and build results of master builds to a specific dir.",
			ArgsUsage: "destination directory to save the artifacts",
			Action: func(c *cli.Context) error {
				dir := "/tmp"
				if c.NArg() > 0 {
					dir = c.Args().Get(0)
				}
				return archiveBuilds(dir)
			},
		},
		{
			Name:    "rerun",
			Aliases: []string{"rr"},
			Usage:   "Rerun a build of the specific PR.",
			Action: func(c *cli.Context) error {
				return rerun("apache", c.Args().Get(0))
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}

}



func run(all bool, authorFilter string) error {
	var key string
	if all {
		key = "pr"
	} else {
		key = "review"
	}
	body, err := cachedGet3min(readGithubApiV4, key)
	if err != nil {
		return err
	}

	result := make(map[string]interface{})
	json.Unmarshal(body, &result)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Author", "Summary", "Participants", "Check"})
	table.SetAutoWrapText(false)
	prs := m(result, "data", "repository", "pullRequests", "edges")

	for _, prNode := range l(prs) {

		pr := m(prNode, "node")

		if !all && !ready(pr) {
			continue
		}

		author := ms(pr, "author", "login")
		participants := getParticipants(pr, author)
		mergeableMark := ""
		destMark := ""
		if ms(pr, "baseRefName") != "master" {
			destMark = "(->" + ms(pr, "baseRefName") + ")"
		}
		if ms(pr, "mergeable") == "CONFLICTING" {
			mergeableMark = "[C] "
		}
		if authorFilter == "" || authorFilter == author {
			table.Append([]string{
				fmt.Sprintf("%d", int(m(pr, "number").(float64))),
				">" + limit(author, 12),
				limit(mergeableMark+destMark+ms(pr, "title"), 50),
				limit(strings.Join(participants, ","), 35),
				buildStatus(pr),
			})
		}
	}
	table.Render() // Send output

	return nil
}

func ready(pr interface{}) bool {
	if ms(pr, "mergeable") == "CONFLICTING" {
		return false
	}
	for _, commitEdge := range l(m(pr, "commits", "edges")) {
		commit := m(commitEdge, "node", "commit")
		for _, suite := range l(m(commit, "checkSuites", "edges")) {
			for _, runs := range l(m(suite, "node", "checkRuns", "edges")) {
				conclusion := ms(runs, "node", "conclusion")
				if conclusion == "FAILURE" || conclusion == "CANCELLED" {
					return false
				}
			}
		}
		break
	}

	for _, review := range lastReviewsPerUser(pr) {
		state := ms(review, "state")
		if state == "CHANGES_REQUESTED" {
			return false
		}
	}

	return true
}

func getParticipants(pr interface{}, author string) []string {
	reviews := lastReviewsPerUser(pr)

	participants := make([]string, 0)

	participants = append(participants, filterReviews(reviews, "CHANGES_REQUESTED", "✕")...)
	participants = append(participants, filterReviews(reviews, "APPROVED", "✓")...)
	participants = append(participants, filterReviews(reviews, "COMMENTED", "")...)

	for _, participant := range l(m(pr, "participants", "edges")) {
		login := ms(participant, "node", "login")
		if _, ok := reviews[login]; !ok && login != author {
			participants = append(participants, limit(login, 5))
		}
	}
	return participants
}

func lastReviewsPerUser(pr interface{}) map[string]interface{} {
	reviewers := make(map[string]interface{})
	for _, review := range l(m(pr, "reviews", "nodes")) {
		author := ms(review, "author", "login")
		if last_review, found := reviewers[author]; found {

			oldRecord, err := time.Parse(time.RFC3339, ms(last_review, "updatedAt"))
			if err != nil {
				panic(err)
			}

			newRecord, err := time.Parse(time.RFC3339, ms(review, "updatedAt"))
			if err != nil {
				panic(err)
			}

			if oldRecord.Before(newRecord) {
				reviewers[author] = review
			}

		} else {
			reviewers[author] = review
		}
	}
	return reviewers
}

func filterReviews(reviews map[string]interface{}, status string, symbol string) []string {
	result := make([]string, 0)
	for _, review := range reviews {
		state := ms(review, "state")
		if state == status {
			result = append(result, symbol+limit(strings.ToUpper(ms(review, "author", "login")), 5))
		}
	}
	return result
}

type statusTransform struct {
	position int
	abbrev   byte
}
