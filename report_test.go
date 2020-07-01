package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindFailures(t *testing.T) {
	failures, err := findFailures("testdata/2020/06/11/1020", "org.apache.hadoop.ozone.om.TestOzoneManagerHAWithData")
	assert.Nil(t, err)
	assert.Len(t, failures, 1)
	assert.Equal(t, "testMultipartUploadWithOneOmNodeDown", failures[0].Method)
}

func TestReadFailuresFromJUnitReport(t *testing.T) {
	testReport := "testdata/2020/06/11/1020/it-hdds-om/hadoop-ozone/integration-test/TEST-org.apache.hadoop.ozone.om.TestOzoneManagerHAWithData.xml"
	failures, err := readFailuresFromJUnitReport(testReport)
	assert.Nil(t, err)
	assert.Len(t, failures, 1)
	assert.Equal(t, "testMultipartUploadWithOneOmNodeDown", failures[0].Method)
}

func TestReadRobotFailingTests(t *testing.T) {
	testReport := "testdata/2020/06/30/1335/acceptance"
	failures, err := readRobotFailingTests(testReport)
	assert.Nil(t, err)
	assert.Len(t, failures, 3)
}