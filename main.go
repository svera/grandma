package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/google/go-github/github"
	"github.com/nlopes/slack"
	"github.com/svera/grandma/config"
	"golang.org/x/oauth2"
)

var cfg *config.Config
var githubClient *github.Client
var slackClient *slack.Client

func main() {
	var err error

	if cfg, err = loadConfig(); err != nil {
		fmt.Println(err.Error())
		return
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GithubToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	githubClient = github.NewClient(tc)

	slackClient := slack.New(cfg.SlackToken)

	if amount, err := calculateTotal(githubClient); err == nil {
		if amount > cfg.Maximum {
			notify(amount, slackClient)
		}
	} else {
		fmt.Println(err)
		return
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

// Note that Github API only return a maximum of 100 results per request, no matter
// what you put in the PerPage property
func calculateTotal(githubClient *github.Client) (int, error) {
	repoListOptions := &github.RepositoryListByOrgOptions{
		Type:        "private",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var amount int
	for {
		repos, resp, err := githubClient.Repositories.ListByOrg(cfg.Organization, repoListOptions)
		if err != nil {
			return 0, fmt.Errorf("Error retrieving repositories")
		}
		for n := range pulls(githubClient, repos) {
			amount += n
		}
		if resp.NextPage == 0 {
			break
		}
		repoListOptions.ListOptions.Page = resp.NextPage
	}
	return amount, nil
}

func pulls(githubClient *github.Client, repos []*github.Repository) <-chan int {
	var wg sync.WaitGroup

	pullListOptions := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	out := make(chan int)
	wg.Add(len(repos))
	for _, repo := range repos {
		go func(repo *github.Repository) {
			if pulls, _, err := githubClient.PullRequests.List(cfg.Organization, *repo.Name, pullListOptions); err != nil {
				log.Println(fmt.Errorf("Error retrieving pull request info"))
			} else {
				out <- len(pulls)
			}
			wg.Done()
		}(repo)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
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
