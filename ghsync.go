package main

import (
	"fmt"
	"log"
	"net/http"

	"bufio"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var (
	ghsyncCmd = &cobra.Command{
		Use:           "ghsync",
		Short:         "ghsync syncs configuration across GitHub repositories",
		Long:          "TODO",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          run,
	}
	cfg struct {
		Labels map[string]string `mapstructure:"labels"`
	}
	configFile string
)

func init() {
	ghsyncCmd.PersistentFlags().String("token", "", "GitHub token to use for API authentication")
	must(viper.BindPFlag("token", ghsyncCmd.PersistentFlags().Lookup("token")))
	must(viper.BindEnv("token", "GITHUB_TOKEN"))

	ghsyncCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "ghsync.yaml", "Configuration file in YAML format")

	log.SetFlags(0)
}

func run(cmd *cobra.Command, args []string) error {
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		abort(fmt.Errorf("Error reading config file: %s", err))
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		abort(fmt.Errorf("Error parsing configuration: %s", err))
	}

	var httpClient *http.Client
	if token := viper.GetString("token"); token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		httpClient = oauth2.NewClient(oauth2.NoContext, ts)
	}
	c := github.NewClient(httpClient)

	repoReader := bufio.NewReader(os.Stdin)
	repoData, err := ioutil.ReadAll(repoReader)
	if err != nil {
		return fmt.Errorf("Error reading repository list: %s", err)
	}
	repos := strings.Split(strings.Trim(string(repoData), " \n"), "\n")

	if err := syncLabels(cfg.Labels, repos, c); err != nil {
		return fmt.Errorf("Error syncing repositories: %s", err)
	}

	return nil
}

func syncLabels(labels map[string]string, repos []string, c *github.Client) error {
	for _, repo := range repos {
		log.Printf("Processing repository %q", repo)
		repoParts := strings.Split(repo, "/")
		if len(repoParts) < 2 {
			return fmt.Errorf("repository %q is missing the owner. Required format: <owner>/<repository>", repo)
		}

		repoLabels := map[string]string{}
		page := 0

		for {
			repoLabelsPage, resp, err := c.Issues.ListLabels(repoParts[0], repoParts[1], &github.ListOptions{
				Page: page,
			})
			if err != nil {
				return fmt.Errorf("could not list existing labels of repo %q: %s", repo, err)
			}

			for _, repoLabel := range repoLabelsPage {
				repoLabels[*repoLabel.Name] = *repoLabel.Color
			}

			if resp.NextPage == 0 {
				break
			}
			page = resp.NextPage
		}

		log.Printf("  Found %d labels", len(repoLabels))

		for label, color := range labels {
			if color[0] == '#' {
				color = color[1:]
			}
			if len(color) != 6 {
				return fmt.Errorf("color %q of label %q for repo %q is in an invalid format. Colors need to be formatted as six hexadecimal digits", color, label, repo)
			}

			repoLabelColor, ok := repoLabels[label]
			if !ok {
				log.Printf("  Creating label %q: %q", label, color)
				_, _, err := c.Issues.CreateLabel(repoParts[0], repoParts[1], &github.Label{
					Name:  github.String(label),
					Color: github.String(color),
				})
				if err != nil {
					return fmt.Errorf("could not create label %q in repo %q: %s", label, repo, err)
				}
			} else if repoLabelColor != color {
				log.Printf("  Updating label %q: %q", label, color)
				_, _, err := c.Issues.EditLabel(repoParts[0], repoParts[1], label, &github.Label{
					Name:  github.String(label),
					Color: github.String(color),
				})
				if err != nil {
					return fmt.Errorf("could not update label %q in repo %q: %s", label, repo, err)
				}
			} else {
				log.Printf("  Label %q: %q already exists", label, color)
			}
		}
	}
	return nil
}

func must(err error) {
	if err != nil {
		abort(err)
	}
}
