package main

import (
	"encoding/json"
	"github.com/elek/go-utils/github"
	"github.com/elek/go-utils/jira"
	jsonhelper "github.com/elek/go-utils/json"
	"github.com/pkg/errors"
	"os"
	"os/user"
	"regexp"
	"strings"
)

func CloseJira(jiraId string) error {
	jiraApi := jira.Jira{
		Url: "https://issues.apache.org/jira",
	}

	updated := map[string]interface{}{
		"fixVersions": []interface{}{
			map[string]interface{}{
				"add":
				map[string]interface{}{
					"name": "1.1.0",
				},
			},
		},
	}

	_, err := jiraApi.DoTransition(jiraId, "5", updated)
	return err
}

func JiraUser() (string, error) {
	if jiraUser := os.Getenv("JIRA_USER"); jiraUser != "" {
		return jiraUser, nil
	}
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	return user.Username, nil
}
func OpenJira(pullRequestId string, githubProject string) error {
	jiraProject := JiraNameFromGithubProject(githubProject)

	jiraApi := jira.Jira{
		Url: "https://issues.apache.org/jira",
	}

	pr, err := jsonhelper.AsJson(github.ReadGithubApiV3("https://api.github.com/repos/apache/" + githubProject + "/pulls/" + pullRequestId))
	if err != nil {
		return err
	}

	title := jsonhelper.MS(pr, "title")
	body := jsonhelper.MS(pr, "body")
	pullUrl := "https://github.com/apache/" + githubProject + "/pull/" + pullRequestId
	issuePattern, err := regexp.Compile(jiraProject + "-[0-9]+")
	if err != nil {
		return err
	}
	jiraId := issuePattern.FindString(title)
	jiraUser, err := JiraUser()
	if err != nil {
		return errors.Wrap(err, "Jira user couldn't be identified")
	}

	if jiraId == "" {
		issue := map[string]interface{}{
			"project": map[string]string{
				"key": jiraProject,
			},
			"summary": title,
			"assignee": map[string]string{
				"name": jiraUser,
			},
			"description": "Please see: " + pullUrl,
			"issuetype": map[string]string{
				"name": "Improvement",
			},
		}
		resp, err := jiraApi.CreateJira(issue)
		respJson, err := jsonhelper.AsJson([]byte(resp), err)
		if err != nil {
			return err
		}
		//{"id":"13348103","key":"HDDS-4627","self":"https://issues.apache.org/jira/rest/api/2/issue/13348103"}
		jiraId = jsonhelper.MS(respJson, "key")

	}
	if jiraId == "" {
		return errors.New("Couldn't get or create jira Id")
	}
	patch := make(map[string]string)
	if !strings.Contains(title, jiraId) {
		patch["title"] = jiraId + ". " + title
	}
	if !strings.Contains(body, jiraId) {
		patch["body"] = "JIRA: https://issues.apache.org/jira/browse/" + jiraId + "\n\n" + body
	}
	if len(patch)>0 {
		patchJson, err := json.Marshal(patch)
		if err != nil {
			return err
		}
		resp, err := github.CallGithubApiV3WithBody("PATCH", "https://api.github.com/repos/apache/"+githubProject+"/pulls/"+pullRequestId, patchJson)
		if err != nil {
			return err
		}
		println(resp.StatusCode)
	}

	return nil
}
