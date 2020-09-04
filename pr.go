package main

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"os"
	"strconv"
	"strings"
	"time"
)

//list pull requests (all/ready)
func run(all bool, authorFilter string, reference Reference) error {
	var key string
	key = reference.Org + "-" + reference.Repo + "-"
	if all {
		key += "pr"
	} else {
		key += "review"
	}
	apiCall := func() ([]byte, error) {
		return readPrWithGraphql(reference)
	}
	body, err := cachedGet3min(apiCall, key)
	if err != nil {
		return err
	}

	result := make(map[string]interface{})
	json.Unmarshal(body, &result)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Upd", "Author", "Summary", "Participants", "Check"})
	table.SetAutoWrapText(false)
	prs := m(result, "data", "repository", "pullRequests", "edges")

	for _, prNode := range l(prs) {

		pr := m(prNode, "node")

		if !all && !ready(pr) {
			continue
		}

		author := prAuthor(pr)
		participants := getParticipants(pr, author)
		feedback := feedbackCount(participants)
		statusMark := ""
		destMark := ""
		if ms(pr, "baseRefName") != "master" {
			destMark = "(->" + ms(pr, "baseRefName") + ")"
		}
		if ms(pr, "mergeable") == "CONFLICTING" {
			statusMark = "[C] "
		}
		if m(pr, "isDraft") == true {
			statusMark += "[D]"
		}

		updated, err := time.Parse(time.RFC3339, ms(pr, "updatedAt"))
		if err != nil {
			return errors.Wrap(err, "Can't parse updateAt field of PR:  "+ms(pr, "number"))
		}

		inactiveTime := time.Now().Sub(updated)

		if authorFilter == "" || authorFilter == author {
			prTitle := limit(statusMark+destMark+ms(pr, "title"), 50)
			if feedback == 0 {
				prTitle = color.YellowString(prTitle)
			}
			table.Append([]string{
				fmt.Sprintf("%d", int(m(pr, "number").(float64))),
				shortDuration(inactiveTime),
				">" + limit(author, 12),
				prTitle,
				strings.Join(participants, ","),
				buildStatus(pr),
			})
		}
	}
	table.Render() // Send output

	return nil
}

func feedbackCount(participants []string) int {
	i := 0
	for _, name := range participants {
		if !strings.Contains(name, "?") {
			i++
		}
	}
	return i
}

func shortDuration(duration time.Duration) string {
	hours := int(duration.Hours())
	var res string
	if hours > 24*30 {
		res = strconv.Itoa(hours/24/30) + "m"
	} else if hours > 24 {
		res = strconv.Itoa(hours/24) + "d"
	} else {
		res = strconv.Itoa(hours) + "h"
	}
	if hours > 168 {
		res = color.RedString(res)
	}
	return res
}

func ready(pr interface{}) bool {
	if mb(pr, "isDraft") {
		return false
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

	participants := make(map[string]string)

	for _, login := range reviewRequests(pr) {
		participants[limit(login, 5)] = "?"
	}

	for _, participant := range l(m(pr, "participants", "edges")) {
		login := ms(participant, "node", "login")
		participants[limit(login, 5)] = ""
	}

	for _, name := range filterReviews(reviews, "CHANGES_REQUESTED") {
		participants[name] = "✕"
	}

	for _, name := range filterReviews(reviews, "APPROVED") {
		participants[name] = "✓"
	}

	for _, name := range filterReviews(reviews, "COMMENTED") {
		participants[name] = "R"
	}

	last := ""
	comments := l(m(pr, "comments", "nodes"))
	if len(comments) > 0 {
		last = limit(ms(comments[0], "author", "login"), 5)
	}

	result := make([]string, 0)
	ix := 0
	for _, name := range sortedParticipants(participants) {

		if name == limit(author, 5) || name == "codec" {
			continue
		}
		status := participants[name]
		if ix > 5 {
			result = append(result, "...")
			break
		}
		ix++

		symbolAndName := name
		if symbolAndName == last {
			symbolAndName = strings.ToUpper(symbolAndName)
		}
		if status == "✓" {
			symbolAndName = status + color.GreenString(name)
		} else if status == "✕" {
			symbolAndName = status + color.RedString(name)
		} else if status == "?" {
			symbolAndName = status + color.YellowString(name)
		} else if status == "R" {
			symbolAndName = name
		}
		result = append(result, symbolAndName)
	}
	return result
}

func sortedParticipants(participants map[string]string) []string {
	result := make([]string, 0)
	for _, symbol := range []string{"✕", "✓", "R", "", "?"} {
		for k, v := range participants {
			if v == symbol {
				result = append(result, k)
			}
		}
	}
	return result
}

func reviewRequests(pr interface{}) []string {
	requests := make([]string, 0)
	for _, request := range l(m(pr, "reviewRequests", "edges")) {
		requests = append(requests, ms(request, "node", "requestedReviewer", "login"))
	}
	return requests
}

func lastReviewsPerUser(pr interface{}) map[string]interface{} {
	prAuthor := prAuthor(pr)
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

		} else if author != prAuthor {
			reviewers[author] = review
		}
	}
	return reviewers
}

func filterReviews(reviews map[string]interface{}, status string) []string {
	result := make([]string, 0)
	for _, review := range reviews {
		state := ms(review, "state")
		if state == status {
			result = append(result, limit(ms(review, "author", "login"), 5))
		}
	}
	return result
}

func prAuthor(pr interface{}) string {
	return ms(pr, "author", "login")
}

type statusTransform struct {
	position int
	abbrev   byte
}
