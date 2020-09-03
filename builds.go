package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func listBuilds(org string, branch string, workflowId int) error {
	fork := org

	cacheKey := "runs-" + org + "-"

	apiUrl := "https://api.github.com/repos/" + fork + "/hadoop-ozone/actions/"
	if workflowId > 0 {
		cacheKey += "-" + strconv.Itoa(workflowId)
		apiUrl += "workflows/" + strconv.Itoa(workflowId) + "/"
	}
	apiUrl += "runs?per_page=50"
	if branch != "" {
		cacheKey += "-" + branch
		apiUrl = apiUrl + "&branch=" + branch
	}
	apiGetter := func() ([]byte, error) {
		return readGithubApiV3(apiUrl)
	}
	runs, err := asJson(cachedGet3min(apiGetter, cacheKey))
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#run", "id", "created", "workflow", "branch", "commit", "Checks"})
	table.SetAutoWrapText(false)
	println()

	for _, run := range l(m(runs, "workflow_runs")) {

		jobsUrl := ms(run, "jobs_url")
		jobs, err := asJson(cachedGet(func() ([]byte, error) {
			return readGithubApiV3(jobsUrl)
		},
			org+"-"+"hadoop-ozone"+"-"+"-actions-runs-"+part(jobsUrl, -2)+"-jobs",
			buildResultCache))
		if err != nil {
			return err
		}

		workflowUrl := ms(run, "workflow_url")
		workflow, err := asJson(cachedGet3min(func() ([]byte, error) {
			return readGithubApiV3(workflowUrl)
		}, org+"-"+"hadoop-ozone-"+"workflow-"+part(workflowUrl, -1)))
		if err != nil {
			return err
		}
		table.Append([]string{
			mns(run, "run_number"),
			"#" + mns(run, "id"),
			ms(run, "created_at"),
			ms(workflow, "name"),
			ms(run, "head_branch"),
			limit(strings.Split(ms(run, "head_commit", "message"), "\n")[0], 50),
			stepsAsString(l(m(jobs, "jobs"))),
		})

	}
	table.Render()

	return err
}

func part(s string, i int) string {
	urlParts := strings.Split(s, "/")
	return urlParts[len(urlParts)+i]

}


func stepsAsString(jobs []interface{}) string {
	groups := make([]string, 4)

	for _, job := range jobs {
		name := m(job, "name").(string)
		conclusion := nilsafe(m(job, "conclusion"))
		status := m(job, "status").(string)

		statusChr := "."
		if strings.ToLower(status) != "completed" {
			statusChr = "%"
		} else {
			if strings.ToLower(conclusion.(string)) == "success" {
				statusChr = "_"
			} else {
				statusChr = string(name[0])
			}
		}

		groupIndex := 0
		if strings.Contains(name, "integration") {
			groupIndex = 1
		} else if strings.Contains(name, "acceptance") {
			groupIndex = 2
		} else if strings.Contains(name, "kubernetes") || strings.Contains(name, "coverage") {
			groupIndex = 3
		}
		groups[groupIndex] += statusChr
	}
	return strings.Join(groups, " ")
}

func buildStatus(pr interface{}) string {
	jobs := make([]interface{}, 0)

	for _, commitEdge := range l(m(pr, "commits", "edges")) {
		commit := m(commitEdge, "node", "commit")
		for _, suite := range l(m(commit, "checkSuites", "edges")) {
			for _, runs := range l(m(suite, "node", "checkRuns", "edges")) {
				node := m(runs, "node")
				jobs = append(jobs, node)
			}

		}
	}
	return stepsAsString(jobs)
}
