package main

func GetWorkflowRunJobs(org string, repo string, runId string) (map[string]interface{}, error) {
	apiGetter := func() ([]byte, error) {
		return readGithubApiV3("https://api.github.com/repos/" + org + "/" + repo + "/actions/runs/" + runId + "/jobs")
	}
	return asJson(cachedGet(apiGetter, org+"-"+repo+"-"+"-actions-runs-"+runId+"-jobs", buildResultCache))
}

func GetArtifacts(org string, repo string, runId string) (map[string]interface{}, error) {
	apiGetter := func() ([]byte, error) {
		return readGithubApiV3("https://api.github.com/repos/" + org + "/" + repo + "/actions/runs/" + runId + "/artifacts")
	}
	return asJson(cachedGet3min(apiGetter, org+"-"+repo+"-"+"-actions-runs-"+runId+"-artifacts"))
}



func GetWorkflowRuns(org string, repo string, workflowId string) (map[string]interface{}, error) {
	apiGetter := func() ([]byte, error) {
		return readGithubApiV3("https://api.github.com/repos/" + org + "/" + repo + "/actions/workflows/" + workflowId + "/runs")
	}
	return asJson(cachedGet3min(apiGetter, org+"-"+repo+"-actions-workflows-"+workflowId+"-runs"))
}

func GetPrCommits(org string, repo string, pullId string) ([]interface{}, error) {
	apiGetter := func() ([]byte, error) {
		return readGithubApiV3("https://api.github.com/repos/" + org + "/" + repo + "/pulls/" + pullId + "/commits")
	}
	return asJsonList(cachedGet3min(apiGetter, org+"-"+repo+"-pulls-"+pullId+"-commits"))
}

func GetChecksForCommits(org string, repo string, commitId string) (map[string]interface{}, error) {
	apiGetter := func() ([]byte, error) {
		return readGithubApiV3("https://api.github.com/repos/" + org + "/" + repo + "/commits/" + commitId + "/check-runs")
	}
	return asJson(cachedGet3min(apiGetter, org+"-"+repo+"-commits-"+commitId+"-check-runs"))
}
