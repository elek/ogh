package main

import (
	"encoding/xml"
	"fmt"
	"github.com/pkg/errors"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type RobotReport struct {
	Statistics RobotStatistics `xml:"statistics"`
}

type RobotStatistics struct {
	Suite RobotSuite `xml:"suite"`
}

type RobotSuite struct {
	Stat []RobotStat `xml:"stat"`
}
type RobotStat struct {
	Name string `xml:"name,attr"`
	Pass int    `xml:"pass,attr"`
	Fail int    `xml:"fail,attr"`
}

type TestFailure struct {
	ClassTimeout bool
	Method       string
	Unknown      bool
	ResultFile   string //relative reference to the result file
}
type TestResult struct {
	Name     string
	Failures []TestFailure
}

type JobResult struct {
	Name         string
	Artifact     string
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

type TestSuite struct {
	TestCase []TestCase `xml:"testcase"`
}

type TestCase struct {
	Name  string  `xml:"name,attr"`
	Error []Error `xml:"error"`
}

type Error struct {
	Message string `xml:"message,attr"`
}

func readFailuresFromJUnitReport(filePath string) ([]TestFailure, error) {
	results := make([]TestFailure, 0)

	report := TestSuite{}
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return results, err
	}
	xml.Unmarshal(content, &report)
	for _, testcase := range report.TestCase {
		if len(testcase.Error) > 0 {
			results = append(results, TestFailure{Method: testcase.Name})
		}
	}
	return results, nil

}

//try to find failure reports in a root dir based on the FQDN name of the test
func findFailures(dir string, testName string) ([]TestFailure, error) {
	results := make([]TestFailure, 0)
	testFileName := "TEST-" + testName + ".xml"
	err := filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && info.Name() == testFileName {
			failures, err := readFailuresFromJUnitReport(filePath)
			if err != nil {
				return err
			}
			results = append(results, failures...)
		}
		return nil
	})
	return results, err
}

func readFailingTests(dir string) ([]TestResult, error) {
	if path.Base(dir) == "acceptance" {
		return readRobotFailingTests(dir)
	} else {
		return readJunitFailingTests(dir)
	}
}

func readRobotFailingTests(dir string) ([]TestResult, error) {
	results := make([]TestResult, 0)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		//acceptance dir exists only in case of errors
		return results, nil
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), "xml") {

			reportFile := path.Join(dir, file.Name())
			content, err := ioutil.ReadFile(reportFile)
			if err != nil {
				return results, errors.Wrap(err, "Can't open "+reportFile)
			}

			robot := RobotReport{}
			err = xml.Unmarshal(content, &robot)
			if err != nil {
				return results, errors.Wrap(err, "Can't parse XML of "+reportFile)
			}

			for _, suite := range robot.Statistics.Suite.Stat {
				if suite.Fail > 0 {
					results = append(results, TestResult{Name: suite.Name})
				}
			}

		}
	}
	return results, nil
}

func readJunitFailingTests(dir string) ([]TestResult, error) {
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
				failures, err := findFailures(dir, trimmedLine)
				if err != nil {
					return results, err
				}
				testResult := TestResult{
					Name:     trimmedLine,
					Failures: failures,
				}
				results = append(results, testResult)
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

		failingTests, err := readFailingTests(path.Join(root, buildPath, JobToArtifactName(ms(job, "name"))))
		if err != nil {
			return b, err
		}
		jobResult := JobResult{
			Name:         ms(job, "name"),
			Artifact:     JobToArtifactName(ms(job, "name")),
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

	funcs := template.FuncMap{
		"limitEnd": func(length int, content string) string {
			if len(content) > length {
				return "..." + content[len(content)-length:len(content)]
			} else {
				return content
			}
		},
		"limit": func(length int, content string) string {
			if len(content) > length {
				return content[0:length] + "..."
			} else {
				return content
			}
		},
		"shortPackage": func(content string) string {
			return strings.Replace(content, "org.apache.hadoop", "o.a.h", -1)
		},
	}
	template := template.Must(template.New("index").Funcs(funcs).Parse(string(indexTemplate)))

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

					buildDir := path.Join(year, month, day, build)
					excludeFile := path.Join(dir, buildDir, "exclude")

					buildNo, _ := strconv.Atoi(build)

					//with an empty exclude file, we can ignore any dir
					//builds older than 872 doesn't have good descriptors
					if _, err := os.Stat(excludeFile); os.IsNotExist(err) && buildNo > 872 {
						buildDirs = append(buildDirs, buildDir)
					}
				}
			}
		}
	}
	return buildDirs, nil
}
