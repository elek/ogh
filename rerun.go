package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func rerun(org string, prId string) error {

	commits, err := GetPrCommits(org, "hadoop-ozone", prId)
	if err != nil {
		return err
	}
	lastCommit := ms(commits[len(commits)-1], "sha")

	workflowRuns, err := GetWorkflowRuns(org, "hadoop-ozone", "4453")
	for _, workflowRun := range l(m(workflowRuns, "workflow_runs")) {
		if ms(workflowRun, "head_sha") == lastCommit {
			resp, err := callGithubApiV3("POST", ms(workflowRun, "rerun_url"))
			if err != nil {
				return err
			}
			if resp.StatusCode != 200 {
				fmt.Println(resp.Status)
				io.Copy(os.Stderr, resp.Body)
			}
			return nil
		}
	}
	return errors.New("Couldn't find recent workflow run with the SHA of the last commit in the PR " + lastCommit)

}
