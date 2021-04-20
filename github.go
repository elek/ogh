package main

import (
	"bytes"
	"encoding/json"
	"github.com/markbates/pkger"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path"
	"strings"
)


func callGithubApiV3(method string, url string) (*http.Response, error) {
	client := &http.Client{}
	log.Debug().Msgf("%s url from GITHUB api: %s ", method, url)

	req, err := http.NewRequest(method, url, nil)
	req.Header.Add("Authorization", "token "+GetToken())
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 299 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error().Msg("Can't read the body of the response: " + err.Error())
		} else {
			log.Error().Msgf(string(body))
		}
		return nil, errors.New(method + " url is failed (" + resp.Status + "): " + url)
	}
	return resp, nil
}

func readGithubApiV3(url string) ([]byte, error) {
	client := &http.Client{}
	log.Debug().Msgf("Reading url from GITHUB api: %s ", url)

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "token "+GetToken())
	req.Header.Add("Accept", "application/vnd.github.antiope-preview+json")
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 299 {
		log.Error().Msgf(string(body))
		return nil, errors.New("Reading url is failed (" + resp.Status + "): " + url)
	}
	return body, nil
}

func readPrWithGraphql(ref Reference) ([]byte, error) {
	client := &http.Client{}

	f, err := pkger.Open("/pr.graphql")
	if err != nil {
		return nil, err
	}

	graphql, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	queryPayload := make(map[string]string)
	queryString := string(graphql)
	queryString = strings.Replace(queryString,
		"owner: \"apache\", name: \"hadoop-ozone\"",
		"owner: \""+ref.Org+"\", name: \""+ref.Repo+"\"", 1)
	queryPayload["query"] = queryString
	query, err := json.Marshal(queryPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewReader(query))
	req.Header.Add("Authorization", "token "+GetToken())
	req.Header.Add("Accept", "application/vnd.github.antiope-preview+json")
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func GetToken() string {
	token := os.Getenv("GITHUB_TOKEN");
	if token != "" {
		return token
	}
	token = getTokenFromHubConfig()
	if token != "" {
		return token
	}
	return getTokenFromGhConfig()
}

func getTokenFromHubConfig() string {
	usr, err := user.Current()
	if err != nil {
		return ""
	}
	hubConfigFile := path.Join(usr.HomeDir, ".config", "hub")
	if _, err := os.Stat(hubConfigFile); os.IsNotExist(err) {
		return ""
	}
	data, err := ioutil.ReadFile(hubConfigFile)
	if err != nil {
		return ""
	}

	hubConfig := make(map[string]interface{})
	err = yaml.Unmarshal(data, &hubConfig)
	if err != nil {
		return ""
	}
	users := l(m(hubConfig, "github.com"))
	if len(users) > 0 {
		return m(users[0], "oauth_token").(string)
	}
	return ""

}

func getTokenFromGhConfig() string {
	usr, err := user.Current()
	if err != nil {
		return ""
	}
	hubConfigFile := path.Join(usr.HomeDir, ".config", "gh", "config.yml")
	if _, err := os.Stat(hubConfigFile); os.IsNotExist(err) {
		return ""
	}
	data, err := ioutil.ReadFile(hubConfigFile)
	if err != nil {
		return ""
	}

	hubConfig := make(map[string]interface{})
	err = yaml.Unmarshal(data, &hubConfig)
	if err != nil {
		return ""
	}
	users := l(m(hubConfig, "github.com"))
	if len(users) > 0 {
		return m(users[0], "oauth_token").(string)
	}
	return ""

}
