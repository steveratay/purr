package main

import (
	"bytes"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/dustin/go-humanize"
	"strings"
	"sync"
	"time"
)

// PullRequest is a normalised version of PullRequest for the different providers
type PullRequest struct {
	ID         int
	Author     string
	Assignee   string
	Updated    time.Time
	WebLink    string
	Title      string
	Repository string
}

func (p *PullRequest) isWIP() bool {
	return strings.Contains(p.Title, "[WIP]") || strings.Contains(p.Title, "WIP:")
}

func (p *PullRequest) isWhiteListed(watchList *WatchList) bool {
	if len(watchList.UserWhiteList) == 0 {
		return true
	}
	for _, user := range watchList.UserWhiteList {
		if user == p.Author || user == p.Assignee {
			return true
		}
	}
	return false
}

func (p *PullRequest) String() string {

	output := fmt.Sprintf(" • <%s|PR #%d> %s  - _%s_", p.WebLink, p.ID, p.Title, p.Author)
	if p.Assignee != "" {
		output += fmt.Sprintf(", assigned to _%s_", p.Assignee)
	}
	output += fmt.Sprintf(" - updated %s", humanize.Time(p.Updated))
	return output
}

// merge merges several channels into one output channel (fan-in)
func merge(channels ...<-chan *PullRequest) <-chan *PullRequest {
	var wg sync.WaitGroup
	out := make(chan *PullRequest)

	// Start an output goroutine for each input channel in channels. output copies values from prs
	// to out until prs is closed, then calls wg.Done
	output := func(prs <-chan *PullRequest) {
		for pr := range prs {
			out <- pr
		}
		wg.Done()
	}

	wg.Add(len(channels))

	for _, c := range channels {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are done. This must start after
	// the wg.Add call
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// filter removes pull requests that should not show up in the final message, this could
// include PRs marked as Work in Progress or where users are not in the whitelist
func filter(watchList *WatchList, in <-chan *PullRequest) chan *PullRequest {
	out := make(chan *PullRequest)

	go func() {
		for list := range in {
			if !list.isWIP() && list.isWhiteListed(watchList) {
				out <- list
			} else {
				logrus.Debugf("filtered pr '%s'", list.Title)
			}
		}
		close(out)
	}()
	return out
}

// format converts all pull requests into a message that is grouped by repo formatted for slack
func format(prs <-chan *PullRequest) fmt.Stringer {
	grouped := make(map[string][]*PullRequest)
	numPRs := 0
	var oldest *PullRequest
	lastUpdated := time.Now()
	for pr := range prs {
		if pr.Updated.Before(lastUpdated) {
			oldest = pr
			lastUpdated = pr.Updated
		}
		numPRs++
		if _, ok := grouped[pr.Repository]; !ok {
			grouped[pr.Repository] = make([]*PullRequest, 0)
		}
		grouped[pr.Repository] = append(grouped[pr.Repository], pr)
	}

	buf := &bytes.Buffer{}
	for repo, prs := range grouped {
		fmt.Fprintf(buf, "*%s*\n", repo)
		for i := range prs {
			fmt.Fprintf(buf, "%s\n", prs[i])
		}
		fmt.Fprint(buf, "\n")
	}

	if numPRs > 0 {
		fmt.Fprintf(buf, "\nThere are currently %d open pull requests", numPRs)
		fmt.Fprintf(buf, " and the oldest (<%s|PR #%d>) was updated %s\n", oldest.WebLink, oldest.ID, humanize.Time(oldest.Updated))
	}
	return buf
}
