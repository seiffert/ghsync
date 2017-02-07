package main

import (
	"fmt"
	"log"

	"time"

	"github.com/google/go-github/github"
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

func syncMilestones(milestones []milestone, repos []repository, service milestoneService) error {
	for _, repo := range repos {
		log.Printf("Processing repository %q", repo)

		repoMilestones, err := fetchMilestones(repo, service)
		if err != nil {
			return err
		}

		log.Printf("  Found %d milestones", len(repoMilestones))

		for _, milestone := range milestones {
			if err := ensureMilestone(milestone, repo, repoMilestones, service); err != nil {
				return err
			}
		}
	}
	return nil
}

func fetchMilestones(repo repository, service milestoneService) ([]*github.Milestone, error) {
	repoMilestones := []*github.Milestone{}
	page := 0
	for {
		repoMilestonesPage, resp, err := service.ListMilestones(repo.Owner, repo.Name, &github.MilestoneListOptions{
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

func ensureMilestone(ms milestone, repo repository, existingMilestones []*github.Milestone, service milestoneService) error {
	if ms.State == "" {
		ms.State = "open"
	}
	if ms.State != "closed" && ms.State != "open" {
		return fmt.Errorf("state %q is invalid. Valid values are \"open\" and \"closed\"", ms.State)
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
		_, _, err := service.CreateMilestone(repo.Owner, repo.Name, &github.Milestone{
			Title:       github.String(ms.Title),
			Description: github.String(ms.Description),
			DueOn:       &dueOn,
			State:       github.String(ms.State),
		})
		if err != nil {
			return fmt.Errorf("could not create milestone %q in repo %q: %s", ms.Title, repo, err)
		}
	} else {
		_, _, err := service.EditMilestone(repo.Owner, repo.Name, *repoMilestone.Number, &github.Milestone{
			Title:       github.String(ms.Title),
			Description: github.String(ms.Description),
			DueOn:       &dueOn,
			State:       github.String(ms.State),
		})
		if err != nil {
			return fmt.Errorf("could not update milestone %q in repo %q: %s", ms.Title, repo, err)
		}
	}
	return nil
}
