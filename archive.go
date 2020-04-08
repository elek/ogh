package main

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"time"
)

func archiveBuilds(outputDir string) error {

	runs, err := GetWorkflowRunsOfBranch("apache", "hadoop-ozone", "8247", "master")
	if err != nil {
		return err
	}

	for _, run := range l(m(runs, "workflow_runs")) {

		createdString := ms(run, "created_at")
		created, err := time.Parse(time.RFC3339, createdString)
		if err != nil {
			return errors.Wrap(err, "Can't parse creation time of the build "+createdString)
		}

		runId := mns(run, "run_number")
		buildDir := path.Join(outputDir, created.Format("2006/01/02"), runId)
		log.Info().Msgf("Download artifacts of build %s", runId)

		if _, err := os.Stat(buildDir); os.IsNotExist(err) {
			_ = os.MkdirAll(buildDir, 0755)
			err = downloadArtifactsOfRun("apache", mns(run, "id"), buildDir, false)
			if err != nil {
				return errors.Wrap(err, "Can't download artifact of the build "+runId)
			}
		} else {
			log.Info().Msgf("%s is already downloaded", buildDir)

		}
	}

	return err
}
