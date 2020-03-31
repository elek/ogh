package main

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"path"
	"time"
)

type getter func() ([]byte, error)

type isCacheValid func(string) (bool, error)

//finished runs can be cached for forever
func buildResultCache(filename string) (bool, error) {

	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return false, errors.Wrap(err, "Couldn't load the cachefile: "+filename)
		}
		jsonData := make(map[string]interface{})
		err = json.Unmarshal(data, &jsonData)
		if err != nil {
			return false, errors.Wrap(err, "Couldn't parse the cachefile to json: "+filename)
		}

		//in case of any uncompleted job, we need to refresh it
		for _, job := range l(m(jsonData, "jobs")) {
			if ms(job, "status") != "completed" {
				return timeCache3min(filename)
			}
		}
		return true, nil
	}
	return false, nil
}
func cachedGet3min(getter getter, key string) ([]byte, error) {
	return cachedGet(getter, key, timeCache3min)
}

func timeCache3min(cacheFile string) (bool, error) {
	if stat, err := os.Stat(cacheFile); !os.IsNotExist(err) {
		if stat.ModTime().Add(3 * time.Minute).After(time.Now()) {
			return true, nil
		}
	}
	return false, nil
}

func cachedGet(getter getter, key string, cacheValidator isCacheValid) ([]byte, error) {
	oghCache := os.Getenv("OGH_CACHE")

	if oghCache == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			oghCache = path.Join(home, ".cache", "ogh")
		}
	}
	cacheFile := ""
	if oghCache != "" {
		cacheFile = path.Join(oghCache, key)
	}

	if cacheFile != "" {
		_ = os.MkdirAll(oghCache, 0700)

		valid, err := cacheValidator(cacheFile)
		if err != nil {
			println("Couldn't validate cache file " + cacheFile + " " + err.Error())
		}
		if err == nil && valid {
			log.Debug().Msgf("'%s' is read from the cache", key)
			return ioutil.ReadFile(cacheFile)
		}
	}
	result, err := getter()
	if cacheFile != "" {
		err = ioutil.WriteFile(cacheFile, result, 0600)
		if err != nil {
			return nil, err
		}
	}
	return result, err
}
