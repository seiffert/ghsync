package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/github"
	"time"
)

type milestone struct {
	Title       string `mapstructure:"title"`
	Description string `mapstructure:"description"`
	State       string `mapstructure:"state"`
	DueDate     string `mapstructure:"due"`
}

type milestoneService interface {
	ListMilestones(owner, repo string, opt *github.MilestoneListOptions) ([]*github.Milestone, *github.Response, error)
	CreateMilestone(owner, repo string, milestone *github.Milestone) (*github.Milestone, *github.Response, error)
	EditMilestone(owner, repo string, number int, milestone *github.Milestone) (*github.Milestone, *github.Response, error)
}

func syncMilestones(milestones []milestone, repos []string, service milestoneService) error {
	for _, repo := range repos {
		log.Printf("Processing repository %q", repo)
		repoParts := strings.Split(repo, "/")
		if len(repoParts) < 2 {
			return fmt.Errorf("repository %q is missing the owner. Required format: <owner>/<repository>", repo)
		}

		repoMilestones, err := fetchMilestones(repoParts[0], repoParts[1], service)
		if err != nil {
			return err
		}

		log.Printf("  Found %d milestones", len(repoMilestones))

		for _, milestone := range milestones {
			if err := ensureMilestone(milestone, repoParts[0], repoParts[1], repoMilestones, service); err != nil {
				return err
			}
		}
	}
	return nil
}

func fetchMilestones(owner, repo string, service milestoneService) ([]*github.Milestone, error) {
	repoMilestones := []*github.Milestone{}
	page := 0
	for {
		repoMilestonesPage, resp, err := service.ListMilestones(owner, repo, &github.MilestoneListOptions{
			State: "all",
			ListOptions: github.ListOptions{
				Page: page,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("could not list existing milestones of repo %q: %s", repo, err)
		}
		repoMilestones = append(repoMilestones, repoMilestonesPage...)

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return repoMilestones, nil
}

func ensureMilestone(ms milestone, owner, repo string, existingMilestones []*github.Milestone, service milestoneService) error {
	if ms.State == "" {
		ms.State = "open"
	}
	if ms.State != "closed" && ms.State != "open" {
		return fmt.Errorf("state %q is invalid. Valid values are \"open\" and \"closed\".", ms.State)
	}

	var repoMilestone *github.Milestone
	for _, existingMilestone := range existingMilestones {
		if ms.Title == *existingMilestone.Title {
			repoMilestone = existingMilestone
		}
	}
	dueOn, err := time.Parse(time.RFC3339, ms.DueDate)
	if err != nil {
		return fmt.Errorf("due date %q is formatted invalid. Valid format: %q", ms.DueDate, time.RFC3339)
	}

	if repoMilestone == nil {
		log.Printf("  Creating milestone %q", ms.Title)
		_, _, err := service.CreateMilestone(owner, repo, &github.Milestone{
			Title:       github.String(ms.Title),
			Description: github.String(ms.Description),
			DueOn:       &dueOn,
			State:       github.String(ms.State),
		})
		if err != nil {
			return fmt.Errorf("could not create milestone %q in repo \"%s/%s\": %s", ms.Title, owner, repo, err)
		}
	} else {
		_, _, err := service.EditMilestone(owner, repo, *repoMilestone.Number, &github.Milestone{
			Title:       github.String(ms.Title),
			Description: github.String(ms.Description),
			DueOn:       &dueOn,
			State:       github.String(ms.State),
		})
		if err != nil {
			return fmt.Errorf("could not update milestone %q in repo \"%s/%s\": %s", ms.Title, owner, repo, err)
		}
	}
	return nil
}
