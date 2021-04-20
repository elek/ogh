package main

import (
	js "github.com/elek/go-utils/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"time"
)

//generate profile/flamegraph of a build based on the downloaded artifacts
func profile(dir string) error {
	jobJson := path.Join(dir, "job.json")
	if _, err := os.Stat(jobJson); os.IsNotExist(err) {
		return errors.New("jon.json couldn't be found in dir " + jobJson)
	}
	job, err := js.AsJson(ioutil.ReadFile(jobJson))
	if err != nil {
		return err
	}
	sum := 0
	for _, jobrun := range js.L(js.M(job, "jobs")) {
		name := js.MS(jobrun, "name")
		end := time.Unix(js.ME(time.RFC3339, jobrun, "completed_at")/1000, 0)
		start := time.Unix(js.ME(time.RFC3339, jobrun, "started_at")/1000, 0)
		duration := int(end.Sub(start).Seconds())
		sum += duration
		println("build;" + name + ";" + strconv.Itoa(duration))

	}
	return nil
}
