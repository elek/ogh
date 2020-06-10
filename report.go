package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

type TestResult struct {
	Name string
}

type JobResult struct {
	Status       string
	Conclusion   string
	FailingTests []TestResult
}

type BuildResult struct {
	ID           string
	Dir          string
	Date         string
	Link         string
	CommitString string
	Conclusion   string
	TestResults  map[string]JobResult
}

func readFailingTests(dir string) ([]TestResult, error) {
	results := make([]TestResult, 0)
	summaryFile := path.Join(dir, "summary.txt")
	if _, err := os.Stat(summaryFile); err == nil {
		summary, err := ioutil.ReadFile(summaryFile)
		if err != nil {
			return results, err
		}
		for _, line := range strings.Split(string(summary), "\n") {
			trimmedLine := strings.Trim(line, " ")
			if len(trimmedLine) > 0 {
				results = append(results, TestResult{Name: trimmedLine})
			}
		}
	}
	return results, nil
}
func parseBuildResults(root string, buildPath string) (BuildResult, error) {
	b := BuildResult{}
	jobs, err := asJson(ioutil.ReadFile(path.Join(root, buildPath, "job.json")))
	if err != nil {
		return b, err
	}
	run, err := asJson(ioutil.ReadFile(path.Join(root, buildPath, "run.json")))
	if err != nil {
		return b, err
	}

	b.Date = ms(run, "created_at")
	b.Dir = buildPath
	b.CommitString = ms(run, "head_commit", "message")
	b.Conclusion = ms(run, "conclusion")
	b.TestResults = make(map[string]JobResult)
	b.ID = mns(run, "run_number")
	b.Link = ms(run, "html_url")
	for _, job := range l(m(jobs, "jobs")) {
		failingTests, err := readFailingTests(path.Join(root, buildPath, ms(job, "name")))
		if err != nil {
			return b, err
		}
		jobResult := JobResult{
			Status:       ms(job, "status"),
			Conclusion:   ms(job, "conclusion"),
			FailingTests: failingTests,
		}
		b.TestResults[ms(job, "name")] = jobResult
	}
	return b, nil
}
func generateReport(dir string) error {
	buildDirs, err := listBuildDirs(dir)
	if err != nil {
		return err
	}

	builds := make([]BuildResult, 0)

	for _, buildDir := range buildDirs {
		br, err := parseBuildResults(dir, buildDir)
		if err != nil {
			fmt.Println(err)
		}
		builds = append(builds, br)
	}
	return renderIndex(path.Join(dir, "templates"), path.Join(dir, "docs"), builds)
}

func renderIndex(templateDir string, destinationDir string, results []BuildResult) error {

	indexTemplate, err := ioutil.ReadFile(path.Join(templateDir, "index.html"))
	if err != nil {
		return err
	}
	template := template.Must(template.New("index").Parse(string(indexTemplate)))

	destFile := path.Join(destinationDir, "index.html")
	destWriter, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer destWriter.Close()
	err = template.Execute(destWriter, results)
	if err != nil {
		log.Println("executing template:", err)
	}
	return nil
}

func getSortedNumberSubdirs(dir string) []string {
	result := make([]string, 0)
	directories, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("WARNING: Can't list directory " + dir + " " + err.Error())
		return result
	}
	for _, directory := range directories {
		if directory.IsDir() {
			if _, err := strconv.Atoi(directory.Name()); err == nil {
				result = append(result, directory.Name())
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		iv, _ := strconv.Atoi(result[i])
		jv, _ := strconv.Atoi(result[j])
		return iv > jv
	})
	return result
}
func listBuildDirs(dir string) ([]string, error) {

	buildDirs := make([]string, 0)
	for _, year := range getSortedNumberSubdirs(dir) {
		for _, month := range getSortedNumberSubdirs(path.Join(dir, year)) {
			for _, day := range getSortedNumberSubdirs(path.Join(dir, year, month)) {
				for _, build := range getSortedNumberSubdirs(path.Join(dir, year, month, day)) {
					buildDirs = append(buildDirs, path.Join(year, month, day, build))
				}
			}
		}
	}
	return buildDirs, nil
}
