package main

import (
//"fmt"
//"github.com/Sirupsen/logrus"
//"github.com/xanzy/go-gitlab"
//"sync"
)

//func trawlGitLab(conf *Config) <-chan *PullRequest {
//	out := make(chan *PullRequest)
//
//	// create a sync group that is used to close the out channel when all gitlab repos has been
//	// trawled
//	var wg sync.WaitGroup
//
//	client := gitlab.NewClient(nil, conf.GitLabToken)
//	if err := client.SetBaseURL(conf.GitlabURL + "/api/v3"); err != nil {
//		usageAndExit(err.Error(), 1)
//	}
//
//	status := "opened"
//	options := &gitlab.ListMergeRequestsOptions{State: &status}
//
//	// spin out each request to find PR on a repo into a separate goroutine
//	for _, repo := range conf.GitLabRepos {
//
//		// increment
//		wg.Add(1)
//
//		go func(repoName string) {
//			defer wg.Done()
//
//			pullRequests, _, err := client.MergeRequests.ListMergeRequests(repoName, options)
//			if err != nil {
//				logrus.Errorf("While fetching PRs from GitLab (%s): %s", repoName, err)
//				return
//			}
//			for _, pr := range pullRequests {
//				out <- &PullRequest{
//					ID:         pr.IID,
//					Author:     pr.Author.Username,
//					Assignee:   pr.Assignee.Username,
//					Updated:    *pr.UpdatedAt,
//					WebLink:    fmt.Sprintf("%s/%s/merge_requests/%d", conf.GitlabURL, repoName, pr.IID),
//					Title:      pr.Title,
//					Repository: repoName,
//				}
//			}
//		}(repo)
//	}
//
//	go func() {
//		wg.Wait()
//		logrus.Debugf("Done with gitlab")
//		close(out)
//	}()
//
//	return out
//}
