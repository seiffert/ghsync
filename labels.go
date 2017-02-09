package main

import (
	"fmt"
	"log"

	"github.com/google/go-github/github"
)

type labelService interface {
	ListLabels(owner, repo string, opt *github.ListOptions) ([]*github.Label, *github.Response, error)
	CreateLabel(owner, repo string, label *github.Label) (*github.Label, *github.Response, error)
	EditLabel(owner, repo, name string, label *github.Label) (*github.Label, *github.Response, error)
}

func syncLabels(labels map[string]string, repos []repository, service labelService) error {
	for _, repo := range repos {
		log.Printf("Processing repository %q", repo)

		repoLabels, err := fetchLabels(repo, service)
		if err != nil {
			return err
		}

		log.Printf("  Found %d labels", len(repoLabels))

		for label, color := range labels {
			if err := ensureLabel(label, color, repo, repoLabels, service); err != nil {
				return err
			}
		}
	}
	return nil
}

func fetchLabels(repo repository, service labelService) (map[string]string, error) {
	repoLabels := map[string]string{}
	page := 0
	for {
		repoLabelsPage, resp, err := service.ListLabels(repo.Owner, repo.Name, &github.ListOptions{
			Page: page,
		})
		if err != nil {
			return nil, fmt.Errorf("could not list existing labels of repo %q: %s", repo, err)
		}

		for _, repoLabel := range repoLabelsPage {
			repoLabels[*repoLabel.Name] = *repoLabel.Color
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return repoLabels, nil
}

func ensureLabel(label, color string, repo repository, existingLabels map[string]string, service labelService) error {
	if color[0] == '#' {
		color = color[1:]
	}
	if len(color) != 6 {
		return fmt.Errorf("color %q of label %q for repo %q is in an invalid format. Colors need to be formatted as six hexadecimal digits", color, label, repo)
	}

	repoLabelColor, ok := existingLabels[label]
	if !ok {
		log.Printf("  Creating label %q: %q", label, color)
		_, _, err := service.CreateLabel(repo.Owner, repo.Name, &github.Label{
			Name:  github.String(label),
			Color: github.String(color),
		})
		if err != nil {
			return fmt.Errorf("could not create label %q in repo %q: %s", label, repo, err)
		}
	} else if repoLabelColor != color {
		log.Printf("  Updating label %q: %q", label, color)
		_, _, err := service.EditLabel(repo.Owner, repo.Name, label, &github.Label{
			Name:  github.String(label),
			Color: github.String(color),
		})
		if err != nil {
			return fmt.Errorf("could not update label %q in repo %q: %s", label, repo, err)
		}
	} else {
		log.Printf("  Label %q: %q already exists", label, color)
	}
	return nil
}
