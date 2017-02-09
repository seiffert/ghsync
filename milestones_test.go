package main

import (
	"testing"

	"math"

	"github.com/google/go-github/github"
)

func TestFetchMilestones_OneMilestone(t *testing.T) {
	repo := newRepository("seiffert/ghsync")
	title := "Version 1.0"
	description := "All issues that need to be done before release 1.0"
	service := &mockMilestoneService{
		Milestones: []*github.Milestone{{
			Title:       github.String(title),
			Description: github.String(description),
		}},
	}

	milestones, err := fetchMilestones(repo, service)
	if err != nil {
		t.Fatalf("should not return an error (%s)", err)
	}
	if len(milestones) != 1 {
		t.Fatal("should return one label")
	}
	if title != *milestones[0].Title {
		t.Errorf("should return milestone %q", title)
	}
	if description != *milestones[0].Description {
		t.Errorf("milestone description does not match: %q != %q", description, *milestones[0].Description)
	}
}

func TestFetchMilestones_TwoPages(t *testing.T) {
	repo := newRepository("seiffert/ghsync")
	service := &mockMilestoneService{
		Milestones: []*github.Milestone{{
			Title: github.String("Version 1.0"),
		}, {
			Title: github.String("Version 2.0"),
		}, {
			Title: github.String("Version 3.0"),
		}, {
			Title: github.String("Version 4.0"),
		}},
	}

	milestones, err := fetchMilestones(repo, service)
	if err != nil {
		t.Fatalf("should not return an error (%s)", err)
	}
	if len(milestones) != len(service.Milestones) {
		t.Fatalf("should return %d milestones, not %d", len(service.Milestones), len(milestones))
	}
}

func TestEnsureMilestone_Create(t *testing.T) {
	newMilestone := milestone{
		Title: "Version 1.0",
	}
	repo := newRepository("seiffert/ghsync")

	service := &mockMilestoneService{}
	milestones := []*github.Milestone{}

	if err := ensureMilestone(newMilestone, repo, milestones, service); err != nil {
		t.Fatalf("should not return an error: %s", err)
	}
	if len(service.Milestones) != 1 {
		t.Fatal("should have created one milestone")
	}
	createdMilestone := service.Milestones[0]
	if *createdMilestone.Title != newMilestone.Title {
		t.Fatalf("milestone title does not match: %s != %s", *createdMilestone.Title, newMilestone.Title)
	}
}

func TestEnsureMilestone_Update(t *testing.T) {
	newMilestone := milestone{
		Title:       "Version 1.0",
		Description: "New description",
	}
	repo := newRepository("seiffert/ghsync")

	milestones := []*github.Milestone{{
		Title:  github.String("Version 1.0"),
		Number: github.Int(1),
	}}
	service := &mockMilestoneService{
		Milestones: milestones,
	}

	if err := ensureMilestone(newMilestone, repo, milestones, service); err != nil {
		t.Fatalf("should not return an error: %s", err)
	}
	if len(service.Milestones) != 1 {
		t.Fatal("should have created one milestone")
	}
	milestone := service.Milestones[0]
	if *milestone.Title != newMilestone.Title {
		t.Fatalf("milestone title does not match: %s != %s", *milestone.Title, newMilestone.Title)
	}
	if *milestone.Description != newMilestone.Description {
		t.Fatalf("milestone description does not match: %s != %s", *milestone.Description, newMilestone.Description)
	}
}

type mockMilestoneService struct {
	Milestones []*github.Milestone
}

func (s *mockMilestoneService) ListMilestones(owner, repo string, opt *github.MilestoneListOptions) ([]*github.Milestone, *github.Response, error) {
	if opt == nil {
		opt = &github.MilestoneListOptions{}
	}

	start := opt.Page * itemsPerPage
	end := int(math.Min(float64(opt.Page*itemsPerPage+itemsPerPage), float64(len(s.Milestones))))

	nextPage := opt.Page + 1
	if end == len(s.Milestones) {
		nextPage = 0
	}

	return s.Milestones[start:end], &github.Response{NextPage: nextPage}, nil
}

func (s *mockMilestoneService) CreateMilestone(owner, repo string, milestone *github.Milestone) (*github.Milestone, *github.Response, error) {
	s.Milestones = append(s.Milestones, milestone)
	return milestone, nil, nil
}

func (s *mockMilestoneService) EditMilestone(owner, repo string, number int, milestone *github.Milestone) (*github.Milestone, *github.Response, error) {
	for _, l := range s.Milestones {
		if *l.Number == number {
			l.Description = milestone.Description
		}
	}
	return milestone, nil, nil
}
