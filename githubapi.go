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

func GetWorkflowRunsOfBranch(org string, repo string, workflowId string, branch string) (map[string]interface{}, error) {
	cacheKey := org + "-" + repo + "-actions-workflows-" + workflowId + "-runs"
	url := "https://api.github.com/repos/" + org + "/" + repo + "/actions/workflows/" + workflowId + "/runs?per_page=100"
	if branch != "" {
		cacheKey += "-" + branch
		url += url + "&branch=" + branch
	}
	apiGetter := func() ([]byte, error) {
		return readGithubApiV3(url)
	}

	return asJson(cachedGet3min(apiGetter, cacheKey))
}

func GetWorkflowRuns(org string, repo string, workflowId string) (map[string]interface{}, error) {
	return GetWorkflowRunsOfBranch(org, repo, workflowId, "")
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
