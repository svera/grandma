package main

import (
	"fmt"
	"log"

	"github.com/google/go-github/github"
	"github.com/nlopes/slack"
	"github.com/svera/vigilante/config"
	"golang.org/x/oauth2"
)

var cfg *config.Config

func main() {
	var cfg *config.Config
	var err error

	if cfg, err = loadConfig(); err != nil {
		fmt.Println(err.Error())
		return
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GithubToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	githubClient := github.NewClient(tc)

	slackClient := slack.New(cfg.SlackToken)

	if amount := calculateTotal(githubClient); amount > cfg.Maximum {
		notify(amount, slackClient)
	}
}

func loadConfig() (*config.Config, error) {
	var data []byte
	var err error
	if data, err = config.Load("/etc/vigilante.yml"); err != nil {
		return nil, err
	}
	return config.Parse(data)
}

func calculateTotal(githubClient *github.Client) int {
	repoListOptions := &github.RepositoryListByOrgOptions{
		Type:        "private",
		ListOptions: github.ListOptions{PerPage: 999},
	}
	pullListOptions := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 999},
	}

	// get all pages of results
	var amount int
	for {
		repos, resp, err := githubClient.Repositories.ListByOrg("magento-mcom", repoListOptions)
		if err != nil {
			fmt.Errorf("Error retrieving repositories")
		}
		for _, repoData := range repos {
			if pulls, _, err := githubClient.PullRequests.List("magento-mcom", *repoData.Name, pullListOptions); err != nil {
				log.Println(fmt.Errorf("Error retrieving pull request info"))
			} else {
				amount += len(pulls)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		repoListOptions.ListOptions.Page = resp.NextPage
	}
	return amount
}

func notify(number int, slackClient *slack.Client) error {
	params := slack.PostMessageParameters{
		Markdown: true,
	}
	_, _, err := slackClient.PostMessage(
		cfg.Channel,
		fmt.Sprintf("You lazy asses! There are %d pull requests waiting to be merged!", number),
		params,
	)
	if err != nil {
		return err
	}
	return nil
}
