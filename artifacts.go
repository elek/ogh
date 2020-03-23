package main

import (
	"archive/zip"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

func downloadArtifacts(org string, buildId int) error {
	fork := org

	cacheKey := "runs-" + org + "-"

	apiUrl := "https://api.github.com/repos/" + fork + "/hadoop-ozone/actions/"
	apiUrl += "runs?per_page=20"
	apiGetter := func() ([]byte, error) {
		return readGithubApiV3(apiUrl)
	}
	runs, err := asJson(cachedGet(apiGetter, cacheKey))
	if err != nil {
		return err
	}

	for _, run := range l(m(runs, "workflow_runs")) {
		if mn(run, "run_number") == buildId {
			id := strconv.Itoa(mn(run, "id"))

			apiGetter := func() ([]byte, error) {
				return readGithubApiV3("https://api.github.com/repos/" + fork + "/hadoop-ozone/actions/runs/" + id + "/artifacts")
			}
			artifacts, err := asJson(cachedGet(apiGetter, fork+"-artifact-"+id))
			if err != nil {
				return err
			}

			for _, artifact := range l(m(artifacts, "artifacts")) {
				name := ms(artifact, "name")
				destDir := path.Join("/tmp/" + strconv.Itoa(buildId))
				err := os.MkdirAll(destDir, 0755)
				if err != nil {
					return err
				}
				println("Downloading results of " + name + " to " + destDir)
				err = downloadAndExtract(name, ms(artifact, "archive_download_url"), destDir)
				if err != nil {
					return err
				}
			}
		}
	}

	return err
}

func downloadAndExtract(name string, url string, destinationDir string) error {

	zipPath := path.Join(destinationDir, name+".zip")

	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		resp, err := callGithubApiV3(url)
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
			println("Exctracting " + f.FileInfo().Name() + " to " + destFile)
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
