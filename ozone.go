package main

import (
	"regexp"
)

var acceptanceRE = regexp.MustCompile(`acceptance \(([^)]+)\)`)
var basicRE = regexp.MustCompile(`basic \(([^)]+)\)`)
var integrationRE = regexp.MustCompile(`integration \(([^)]+)\)`)

func JobToArtifactName(job string) string {
	if basicRE.MatchString(job) {
		return basicRE.ReplaceAllString(job, "$1")
	}
	if acceptanceRE.MatchString(job) {
		return acceptanceRE.ReplaceAllString(job, "acceptance-$1")
	}
	if integrationRE.MatchString(job) {
		return integrationRE.ReplaceAllString(job, "it-$1")
	}
	return job
}
