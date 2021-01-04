package main

import (
	"encoding/json"
	"github.com/elek/go-utils/github"
	"github.com/elek/go-utils/jira"
	jsonhelper "github.com/elek/go-utils/json"
	"github.com/pkg/errors"
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

func OpenJira(pullRequestId string) error {
	jiraApi := jira.Jira{
		Url: "https://issues.apache.org/jira",
	}

	pr, err := jsonhelper.AsJson(github.ReadGithubApiV3("https://api.github.com/repos/apache/ozone/pulls/" + pullRequestId))
	if err != nil {
		return err
	}

	title := jsonhelper.MS(pr, "title")
	body := jsonhelper.MS(pr, "body")
	pullUrl := "https://github.com/apache/ozone/pull/" + pullRequestId
	issuePattern, err := regexp.Compile("HDDS-[0-9]+")
	if err != nil {
		return err
	}
	jiraId := issuePattern.FindString(title)
	if jiraId == "" {
		issue := map[string]interface{}{
			"project": map[string]string{
				"key": "HDDS",
			},
			"summary":     title,
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
		resp, err := github.CallGithubApiV3WithBody("PATCH", "https://api.github.com/repos/apache/ozone/pulls/"+pullRequestId, patchJson)
		if err != nil {
			return err
		}
		println(resp.StatusCode)
	}

	return nil
}
