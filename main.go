package main

import (
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/pkg/errors"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/skratchdot/open-golang/open"
	"github.com/urfave/cli"
)

var version string
var commit string
var date string

//repository object reference in the format org/repo@branch#id
type Reference struct {
	Org    string
	Repo   string
	Branch string
	Id     string
}

func ParseReference(str string) Reference {
	ref := Reference{
		Org:    "apache",
		Repo:   "hadoop-ozone",
		Branch: "master",
		Id:     "",
	}
	refStr := str
	if strings.Contains(refStr, "/") {
		ref.Org = strings.Split(refStr, "/")[0]
		refStr = refStr[len(ref.Org)+1:]
	}
	if strings.Contains(refStr, "#") {
		ref.Id = strings.Split(refStr, "#")[1]
		refStr = refStr[0 : len(refStr)-len(ref.Id)-1]
	}
	if strings.Contains(refStr, "@") {
		ref.Branch = strings.Split(refStr, "@")[1]
		refStr = refStr[0 : len(refStr)-len(ref.Branch)-1]
	}
	if len(refStr) > 0 {
		ref.Repo = refStr
	}
	return ref
}

var app *cli.App = cli.NewApp()

func init() {
	app.Name = "ogh"
	app.Usage = "Helper cli for Apache Hadoop Ozone development"
	app.Description = "Various helper scripts to query github API to make the development faster."
	app.Version = fmt.Sprintf("%s (%s, %s)", version, commit, date)

	app.Commands = append(app.Commands, []cli.Command{
		{
			Name:    "review",
			Aliases: []string{"r"},
			Usage:   "Show the review queue (all READY pull requests)",
			Action: func(c *cli.Context) error {
				ref := ParseReference(c.Args().Get(0))
				return run(false, "", ref)
			},
		},
		{
			Name:    "pull-requests",
			Aliases: []string{"pr"},
			Usage:   "Show all the available pull requests",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "user",
					Usage: "Github user or organization name",
					Value: "",
				},
			},
			Action: func(c *cli.Context) error {
				ref := ParseReference(c.Args().Get(0))
				return run(true, c.String("username"), ref)
			},
		},
		{
			Name:    "mine",
			Aliases: []string{"m"},
			Usage:   "Show results of the pr from the current user",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "user",
					Usage: "Github user name. (Default: current user)",
					Value: "",
				},
			},
			Action: func(c *cli.Context) error {
				ref := ParseReference(c.Args().Get(0))
				return run(true, getUser(c), ref)
			},
		},
		{
			Name:  "artifacts",
			Usage: "Download build artifacts",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "user",
					Usage: "Github user or organization name",
					Value: "apache",
				},
				cli.StringFlag{
					Name:  "dir",
					Usage: "Destination dir to save the downloaded artifacts",
					Value: "/tmp",
				},
				cli.BoolFlag{
					Name:  "all",
					Usage: "If not used, only the failed artifacts will be downloaded.",
				},
			},
			Action: func(c *cli.Context) error {
				return downloadArtifacts(c.String("user"), c.Args().Get(0), c.String("dir"), c.Bool("all"))
			},
		},
		{
			Name:  "profile",
			Usage: "Profile github run based on downloaded arguments",
			Action: func(c *cli.Context) error {
				return profile(c.Args()[0])
			},
		},
		{
			Name:    "builds",
			Aliases: []string{"b"},
			Usage:   "Print results of branch builds.",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "user",
					Usage: "Github user or organization name",
					Value: "apache",
				},
				cli.StringFlag{
					Name:  "workflow",
					Usage: "Id of the workflow to list the builds",
					Value: "8247",
				},
				cli.StringFlag{
					Name:  "branch",
					Usage: "Check the builds of this specific run",
					Value: "master",
				},
			},
			Action: func(c *cli.Context) error {
				return listBuilds(c.String("user"), c.String("branch"), c.Int("workflow"))
			},
		},
		{
			Name:    "upcomming",
			Aliases: []string{"uc"},
			Usage:   "Show builds sent to the local fork",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "user",
					Usage: "Github user or organization name (can be set by GITHUB_USER, default to the local user)",
					Value: "",
				},
			},
			Action: func(c *cli.Context) error {
				return listForkBuilds(getUser(c))
			},
		},
		{
			Name:      "archive",
			Usage:     "Save artifacts and build results of master builds to a specific dir.",
			ArgsUsage: "destination directory to save the artifacts",
			Action: func(c *cli.Context) error {
				dir := "/tmp"
				if c.NArg() > 0 {
					dir = c.Args().Get(0)
				}
				return archiveBuilds(dir)
			},
		},
		{
			Name:  "jira",
			Usage: "Jira related helper methods",
			Subcommands: []cli.Command{
				{
					Name:  "close",
					Usage: "Close jira with proper fix version",
					Action: func(c *cli.Context) error {
						if c.NArg() > 0 {
							return CloseJira(c.Args().Get(0))
						} else {
							return errors.New("Please specify the jira ID")
						}
					},
				},
				{
					Name:  "open",
					Usage: "Open jira for a specific pull request",
					Action: func(c *cli.Context) error {
						if c.NArg() > 0 {
							return OpenJira(c.Args().Get(0), getProject(c))
						} else {
							return errors.New("Please specify the pull request ID")
						}
					},
				},
			},
		},
		{
			Name:  "report",
			Usage: "Generate HTML report from archived directory structure.",
			Action: func(c *cli.Context) error {
				ex, err := os.Executable()
				if err != nil {
					return err
				}
				dir := filepath.Dir(ex)

				if c.NArg() > 0 {
					dir = c.Args().Get(0)
				}
				return generateReport(dir)
			},
		},
		{
			Name:    "rerun",
			Aliases: []string{"rr"},
			Usage:   "Rerun a build of the specific PR.",
			Action: func(c *cli.Context) error {
				return rerun("apache", c.Args().Get(0))
			},
		},
	}...)
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	if len(os.Args) == 2 {
		_, err := strconv.Atoi(os.Args[1])
		if err == nil {
			err := open.Start("http://github.com/apache/hadoop-ozone/pull/" + os.Args[1])
			if err != nil {
				panic(err)
			}
			return
		}
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("%-v", err)
		panic(err)
	}

}

func getUser(c *cli.Context) string {
	userName := c.String("user")

	if userName == "" {
		userName = os.Getenv("GITHUB_USER")
	}

	if userName == "" {
		user, err := user.Current()
		if err == nil {
			userName = user.Username
		}
	}

	return userName
}

//recursive parent search for a .git directory
func findGitDir(dir string) string {
	if _, err := os.Stat(path.Join(dir, ".git")); os.IsNotExist(err) {
		if dir == "/" {
			return ""
		} else {
			return findGitDir(filepath.Dir(dir))
		}
	} else {
		return path.Join(dir, ".git")
	}
}

func JiraNameFromGithubProject(githubProject string) string {
	if githubProject == "ozone" {
		return "HDDS"
	}
	project := strings.ReplaceAll(githubProject, "incubator-", "")
	return strings.ToUpper(project)
}

//return the defined project (or try to auto-detect)
func getProject(c *cli.Context) string {
	project := c.String("project")

	if project == "" {

		worktree := memfs.New()

		wd, err := os.Getwd()
		if err == nil {

			st := filesystem.NewStorage(osfs.New(findGitDir(wd)), cache.NewObjectLRUDefault())

			repository, err := git.Open(st, worktree)
			if err != nil {
				log.Err(err)
			}
			remotes, err := repository.Remotes()
			if err != nil {
				log.Err(err)
			}
			for _, remote := range remotes {
				if remote.Config().Name == "origin" {
					remoteUrl := remote.Config().URLs[0]
					parts := strings.Split(remoteUrl, "/")
					reponame := parts[len(parts)-1]
					project = strings.ReplaceAll(reponame, ".git", "")
				}
			}
		}

	}

	if project == "" {
		project = "ozone"
	}
	return project
}
