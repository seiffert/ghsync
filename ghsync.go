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

	repos, err := readInRepositories()
	if err != nil {
		return fmt.Errorf("Error reading repository list: %s", err)
	}

	if err := syncLabels(cfg.Labels, repos, c.Issues); err != nil {
		return fmt.Errorf("Error syncing repositories: %s", err)
	}

	return nil
}

func readInRepositories() ([]string, error) {
	repoReader := bufio.NewReader(os.Stdin)
	repoData, err := ioutil.ReadAll(repoReader)
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.Trim(string(repoData), " \n"), "\n"), nil
}

func must(err error) {
	if err != nil {
		abort(err)
	}
}
