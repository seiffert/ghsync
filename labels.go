package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/github"
)

type labelService interface {
	ListLabels(owner, repo string, opt *github.ListOptions) ([]*github.Label, *github.Response, error)
	CreateLabel(owner, repo string, label *github.Label) (*github.Label, *github.Response, error)
	EditLabel(owner, repo, name string, label *github.Label) (*github.Label, *github.Response, error)
}

func syncLabels(labels map[string]string, repos []string, service labelService) error {
	for _, repo := range repos {
		log.Printf("Processing repository %q", repo)
		repoParts := strings.Split(repo, "/")
		if len(repoParts) < 2 {
			return fmt.Errorf("repository %q is missing the owner. Required format: <owner>/<repository>", repo)
		}

		repoLabels, err := fetchLabels(repoParts[0], repoParts[1], service)
		if err != nil {
			return err
		}

		log.Printf("  Found %d labels", len(repoLabels))

		for label, color := range labels {
			if err := ensureLabel(label, color, repoParts[0], repoParts[1], repoLabels, service); err != nil {
				return err
			}
		}
	}
	return nil
}

func fetchLabels(owner, repo string, service labelService) (map[string]string, error) {
	repoLabels := map[string]string{}
	page := 0
	for {
		repoLabelsPage, resp, err := service.ListLabels(owner, repo, &github.ListOptions{
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

func ensureLabel(label, color, owner, repo string, existingLabels map[string]string, service labelService) error {
	if color[0] == '#' {
		color = color[1:]
	}
	if len(color) != 6 {
		return fmt.Errorf("color %q of label %q for repo \"%s/%s\" is in an invalid format. Colors need to be formatted as six hexadecimal digits", color, label, owner, repo)
	}

	repoLabelColor, ok := existingLabels[label]
	if !ok {
		log.Printf("  Creating label %q: %q", label, color)
		_, _, err := service.CreateLabel(owner, repo, &github.Label{
			Name:  github.String(label),
			Color: github.String(color),
		})
		if err != nil {
			return fmt.Errorf("could not create label %q in repo \"%s/%s\": %s", label, owner, repo, err)
		}
	} else if repoLabelColor != color {
		log.Printf("  Updating label %q: %q", label, color)
		_, _, err := service.EditLabel(owner, repo, label, &github.Label{
			Name:  github.String(label),
			Color: github.String(color),
		})
		if err != nil {
			return fmt.Errorf("could not update label %q in repo \"%s/%s\": %s", label, owner, repo, err)
		}
	} else {
		log.Printf("  Label %q: %q already exists", label, color)
	}
	return nil
}
