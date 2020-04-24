package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/skratchdot/open-golang/open"
	"github.com/urfave/cli"
	"os"
	"os/user"
	"strconv"
)

var version string
var commit string
var date string

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

	app := cli.NewApp()
	app.Name = "ogh"
	app.Usage = "Helper cli for Apache Hadoop Ozone development"
	app.Description = "Various helper scripts to query github API to make the development faster."
	app.Version = fmt.Sprintf("%s (%s, %s)", version, commit, date)
	app.Commands = []cli.Command{
		{
			Name:    "review",
			Aliases: []string{"r"},
			Usage:   "Show the review queue (all READY pull requests)",
			Action: func(c *cli.Context) error {
				return run(false, "")
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
				return run(true, c.String("user"))
			},
		},
		{
			Name:    "mine",
			Aliases: []string{"pr"},
			Usage:   "Show results of the pr from the current user",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "user",
					Usage: "Github user name. (Default: current user)",
					Value: "",
				},
			},
			Action: func(c *cli.Context) error {
				userName := c.String("user")
				if userName == "" {
					user, err := user.Current()
					if err == nil {
						userName = user.Username
					}
				}
				return run(true, userName)
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
			Name:    "rerun",
			Aliases: []string{"rr"},
			Usage:   "Rerun a build of the specific PR.",
			Action: func(c *cli.Context) error {
				return rerun("apache", c.Args().Get(0))
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}

}

