package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func archiveBuilds(outputDir string) error {

	runs, err := GetWorkflowRunsOfBranch("apache", "hadoop-ozone", "8247", "master")
	if err != nil {
		return err
	}

	limit := 50
	for _, run := range l(m(runs, "workflow_runs")) {

		if (ms(run,"event") == "pull_request") {
			continue
		}

		limit = limit - 1
		if limit == 0 {
			break
		}
		createdString := ms(run, "created_at")
		created, err := time.Parse(time.RFC3339, createdString)
		if err != nil {
			return errors.Wrap(err, "Can't parse creation time of the build "+createdString)
		}

		runId := mns(run, "run_number")
		buildDir := path.Join(outputDir, created.Format("2006/01/02"), runId)
		log.Info().Msgf("Download artifacts of build %s", runId)

		runJson := path.Join(buildDir, "run.json")
		niceJobJson, err := json.MarshalIndent(run, "", "   ")
		if err != nil {
			return errors.Wrap(err, "Can't parse job API, runId="+runId)
		}
		_ = os.MkdirAll(filepath.Dir(runJson), 0755)
		err = ioutil.WriteFile(runJson, niceJobJson, 0755)
		if err != nil {
			return errors.Wrap(err, "Can't write out run json file"+runJson)
		}

		jobJson := path.Join(buildDir, "job.json")
		//we can skip the download if the job is already downloaded and all the
		//jobs were finished
		if _, err := os.Stat(jobJson); os.IsNotExist(err) {
		} else {
			jobContent, err := asJson(ioutil.ReadFile(jobJson))
			if err != nil {
				return err
			}
			allDone := true
			for _, job := range l(m(jobContent, "jobs")) {
				if ms(job, "status") != "completed" {
					allDone = false
				}
			}
			if allDone {
				continue
			}
			log.Print(runId + " is already downloaded but it was in-progress")
		}
		_ = os.MkdirAll(buildDir, 0755)
		err = downloadArtifactsOfRun("apache", mns(run, "id"), buildDir, false)
		if err != nil {
			return errors.Wrap(err, "Can't download artifact of the build "+runId)
		}

	}

	return err
}
