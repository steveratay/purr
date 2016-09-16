package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"strings"
	"sync"
)

func trawlGitHub(conf *Config, watchList *WatchList) <-chan *PullRequest {

	out := make(chan *PullRequest)

	// create a sync group that is used to close the out channel when all github repos has been
	// trawled
	var wg sync.WaitGroup

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: conf.GitHubToken})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)

	// check for wildcards in the repo name and expand them into individual repos
	var repos []string
	for _, repoName := range watchList.Repositories {
		repoParts := strings.Split(repoName, "/")
		if len(repoParts) != 2 {
			logrus.Errorf("%s is not a valid GitHub repository\n", repoName)
			continue
		}
		if repoParts[1] != "*" {
			repos = append(repos, repoName)
			continue
		}
		logrus.Debugf("expanding wildcard on %s", repoName)
		allRepos, _, err := client.Repositories.List(repoParts[0], nil)
		if err != nil {
			logrus.Error(err)
			continue
		}
		for i := range allRepos {
			repos = append(repos, fmt.Sprintf("%s/%s", repoParts[0], *allRepos[i].Name))
		}
	}

	// spin out each request to find PRs on a repo into a separate goroutine so we fetch them
	// asynchronous
	for _, repo := range repos {

		// increment the wait group
		wg.Add(1)

		go func(repoName string) {
			// when finished, decrement the wait group
			defer wg.Done()
			logrus.Debugf("Starting fetch from %s", repoName)

			parts := strings.Split(repoName, "/")

			// nextPage keeps track of of the current page of the paginataed response from the
			// GitHub API
			nextPage := 1
			for {
				// options for the request for PRs
				options := &github.PullRequestListOptions{
					State:     "open",
					Sort:      "updated",
					Direction: "desc",
					ListOptions: github.ListOptions{
						Page: nextPage,
					},
				}

				// get the pull requests
				pullRequests, resp, err := client.PullRequests.List(parts[0], parts[1], options)
				if err != nil {
					logrus.Errorf("While fetching PRs from GitHub (%s/%s): %s", parts[0], parts[1], err)
					return
				}

				// transform the GitHub pull request struct into a provider agnostic struct
				for _, pr := range pullRequests {
					pullRequest := &PullRequest{
						ID:         *pr.Number,
						Author:     *pr.User.Login,
						Updated:    *pr.UpdatedAt,
						WebLink:    *pr.HTMLURL,
						Title:      *pr.Title,
						Repository: fmt.Sprintf("%s/%s", parts[0], parts[1]),
					}
					if pr.Assignee != nil {
						pullRequest.Assignee = *pr.Assignee.Login
					}

					// push to the outchannel
					out <- pullRequest
				}

				// the GitHub API returns 0 as the LastPage if there are no more pages of result
				if resp.LastPage == 0 {
					break
				}
				nextPage++

			}
		}(repo)
	}

	// Spin off a go routine that will close the channel when all repos have finished
	go func() {
		wg.Wait()
		logrus.Debugf("Done with github")
		close(out)
	}()

	return out
}
