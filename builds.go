package main

import (
	"github.com/rs/zerolog/log"
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

var transmap = map[string]statusTransform{
	"compile":                           statusTransform{0, 'b'},
	"rat":                               statusTransform{1, 'r'},
	"author":                            statusTransform{2, 'a'},
	"checkstyle":                        statusTransform{3, 'c'},
	"findbugs":                          statusTransform{4, 'f'},
	"unit":                              statusTransform{5, 'u'},
	"acceptance":                        statusTransform{6, 'a'},
	"it-freon":                          statusTransform{8, 'f'},
	"it-filesystem":                     statusTransform{9, 's'},
	"it-filesystem-contract":            statusTransform{10, 'c'},
	"it-client-and-hdds":                statusTransform{11, 'h'},
	"it-client":                         statusTransform{11, 'c'},
	"it-hdds-om":                        statusTransform{12, 'm'},
	"it-om":                             statusTransform{12, 'm'},
	"it-ozone":                          statusTransform{13, 'o'},
	"integration (freon)":               statusTransform{8, 'f'},
	"integration (filesystem)":          statusTransform{9, 's'},
	"integration (filesystem-hdds)":     statusTransform{9, 's'},
	"integration (filesystem-contract)": statusTransform{10, 't'},
	"integration (client)":              statusTransform{11, 'c'},
	"integration (hdds-om)":             statusTransform{12, 'm'},
	"integration (ozone)":               statusTransform{13, 'o'},
	"coverage":                          statusTransform{19, 'c'},
	"acceptance (secure)":               statusTransform{15, 'm'},
	"acceptance (unsecure)":             statusTransform{16, 'm'},
	"acceptance (misc)":                 statusTransform{17, 'm'},
}

func stepsAsString(jobs []interface{}) string {
	result := []byte("....... ...... ... .")
	for _, job := range jobs {
		name := m(job, "name").(string)
		conclusion := nilsafe(m(job, "conclusion"))
		status := m(job, "status").(string)
		if statusTrafo, ok := transmap[name]; ok {
			statusChr := byte('.')
			if status != "completed" {
				statusChr = byte('%')
			} else {
				if conclusion == "success" {
					statusChr = byte('_')
				} else {
					statusChr = byte(statusTrafo.abbrev)
				}
			}
			result[statusTrafo.position] = statusChr
		} else {
			log.Warn().Msg("transformation item is missing for job " + name)
		}
	}
	return string(result)
}

func buildStatus(pr interface{}) string {
	result := []byte("....... ...... ... .")

	for _, commitEdge := range l(m(pr, "commits", "edges")) {
		commit := m(commitEdge, "node", "commit")
		for _, context := range l(m(commit, "status", "contexts")) {
			cx := ms(context, "context")
			if statusTrafo, ok := transmap[cx]; ok {
				statusChr := byte('.')
				switch ms(context, "state") {
				case "SUCCESS":
					statusChr = byte('_')
				case "PENDING":
					statusChr = byte('%')
				case "FAILURE":
					statusChr = statusTrafo.abbrev
				}
				result[statusTrafo.position] = statusChr
			}
		}

		for _, suite := range l(m(commit, "checkSuites", "edges")) {
			for _, runs := range l(m(suite, "node", "checkRuns", "edges")) {
				name := m(runs, "node", "name").(string)
				conclusion := nilsafe(m(runs, "node", "conclusion"))
				status := m(runs, "node", "status").(string)
				if statusTrafo, ok := transmap[name]; ok {
					statusChr := byte('.')
					if status != "COMPLETED" {
						statusChr = byte('%')
					} else {
						if conclusion == "SUCCESS" {
							statusChr = byte('_')
						} else {
							statusChr = byte(statusTrafo.abbrev)
						}
					}
					result[statusTrafo.position] = statusChr
				}
			}

		}
	}
	return string(result)
}
