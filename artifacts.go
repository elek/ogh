package main

import (
	"archive/zip"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func downloadArtifacts(org string, buildIdExpression string, destinationDir string, all bool) error {

	if strings.HasPrefix(buildIdExpression, "pr/") {
		pr, err := GetPr(org, "hadoop-ozone", buildIdExpression[3:])
		if err != nil {
			return err
		}
		branch := ms(pr, "head", "ref")

		workflowRuns, err := GetWorkflowRunsOfBranch(org, "hadoop-ozone", "4453", branch)
		if err != nil {
			return err
		}
		id := mns(l(m(workflowRuns, "workflow_runs"))[0], "id")
		return downloadArtifactsOfRun(org, id, destinationDir+"/"+buildIdExpression, false)
	} else if strings.HasPrefix(buildIdExpression, "#") {
		return downloadArtifactsOfRun(org, buildIdExpression[1:], destinationDir+"/"+buildIdExpression[1:], all)
	} else {
		workflowRuns, err := GetAllWorkflowRuns(org, "hadoop-ozone")

		if err == nil {
			for _, run := range l(m(workflowRuns, "workflow_runs")) {
				runId := mns(run, "id")
				if mns(run, "run_number") == buildIdExpression {
					return downloadArtifactsOfRun(org, runId, destinationDir+"/"+runId, all)
				}

				if buildIdExpression == runId {
					return downloadArtifactsOfRun(org, runId, destinationDir+"/"+runId, all)
				}
			}
		}

	}
	return errors.New("Build is not found: " + buildIdExpression)
}

func downloadArtifactsOfRun(org string, runId string, destinationDir string, all bool) error {

	apiGetter := func() ([]byte, error) {
		return readGithubApiV3("https://api.github.com/repos/" + org + "/hadoop-ozone/actions/runs/" + runId + "/artifacts")
	}
	artifacts, err := asJson(cachedGet3min(apiGetter, org+"-actions-runs-"+runId+"-artifacts"))
	if err != nil {
		return err
	}

	results := make(map[string]interface{})
	jobs, err := GetWorkflowRunJobs(org, "hadoop-ozone", runId)
	if err != nil {
		return err
	}
	for _, job := range l(m(jobs, "jobs")) {
		results[JobToArtifactName(ms(job, "name"))] = ms(job, "conclusion")
	}

	err = os.MkdirAll(destinationDir, 0755)
	if err != nil {
		return errors.Wrap(err, "Can't created destination directory: "+destinationDir)
	}
	niceJobJson, err := json.MarshalIndent(jobs, "", "   ")
	if err != nil {
		return errors.Wrap(err, "Can't parse job API, runId="+runId)
	}
	jsonJobFile := path.Join(destinationDir, "job.json")

	err = ioutil.WriteFile(jsonJobFile, niceJobJson, 0755)
	if err != nil {
		return errors.Wrap(err, "Can't write out job file to "+jsonJobFile)
	}

	for _, artifact := range l(m(artifacts, "artifacts")) {
		name := ms(artifact, "name")
		result, found := results[name]
		if !found {
			return errors.New("Job result for the artifact " + name + " is unknown")
		}
		if all || result == "failure" {

			println("Downloading results of " + name + " to " + destinationDir)
			err = downloadAndExtract(name, ms(artifact, "archive_download_url"), destinationDir)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func downloadAndExtract(name string, url string, destinationDir string) error {

	zipPath := path.Join(destinationDir, name+".zip")

	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		resp, err := callGithubApiV3("GET", url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		zipFile, err := os.Create(zipPath)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipFile, resp.Body)
		if err != nil {
			return err
		}
	}
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	defer os.Remove(zipPath)

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		destFile := filepath.Join(destinationDir, name, f.Name)
		if f.FileInfo().IsDir() {
			err = os.MkdirAll(destFile, 0755)
			if err != nil {
				return err
			}
		} else {
			println("Extracting " + f.FileInfo().Name() + " to " + destFile)
			err = os.MkdirAll(path.Dir(destFile), 0755)
			if err != nil {
				return err
			}
			f, err := os.OpenFile(
				destFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
