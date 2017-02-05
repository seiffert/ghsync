package main

import (
	"testing"

	"math"

	"github.com/google/go-github/github"
)

const itemsPerPage = 3

func TestFetchLabels_OneLabel(t *testing.T) {
	name, color := "foo", "ffffff"
	service := &mockLabelService{
		Labels: []*github.Label{{
			Name:  github.String(name),
			Color: github.String(color),
		}},
	}

	owner, repo := "seiffert", "ghsync"

	labels, err := fetchLabels(owner, repo, service)
	if err != nil {
		t.Fatalf("should not return an error (%s)", err)
	}
	if len(labels) != 1 {
		t.Fatal("should return one label")
	}
	if actualColor, ok := labels[name]; !ok {
		t.Errorf("should return label %q", name)
	} else if actualColor != color {
		t.Errorf("color does not match: %q != %q", actualColor, color)
	}
}

func TestFetchLabels_TwoPages(t *testing.T) {
	service := &mockLabelService{
		Labels: []*github.Label{{
			Name:  github.String("foo"),
			Color: github.String("aaaaaa"),
		}, {
			Name:  github.String("bar"),
			Color: github.String("bbbbbb"),
		}, {
			Name:  github.String("baz"),
			Color: github.String("cccccc"),
		}, {
			Name:  github.String("test"),
			Color: github.String("dddddd"),
		}},
	}

	owner, repo := "seiffert", "ghsync"

	labels, err := fetchLabels(owner, repo, service)
	if err != nil {
		t.Fatalf("should not return an error (%s)", err)
	}
	if len(labels) != len(service.Labels) {
		t.Fatalf("should return %d label, not %d", len(service.Labels), len(labels))
	}
}

func TestEnsureLabel_Create(t *testing.T) {
	name, color := "foo", "ffffff"
	owner, repo := "seiffert", "ghsync"

	service := &mockLabelService{}
	labels := map[string]string{}

	if err := ensureLabel(name, color, owner, repo, labels, service); err != nil {
		t.Fatalf("should not return an error: %s", err)
	}
	if len(service.Labels) != 1 {
		t.Fatal("should have created one label")
	}
	label := service.Labels[0]
	if *label.Name != name {
		t.Fatalf("label name does not match: %s != %s", *label.Name, name)
	}
	if *label.Color != color {
		t.Fatalf("label color does not match: %s != %s", *label.Color, color)
	}
}

func TestEnsureLabel_CreateLeadingHash(t *testing.T) {
	name, color := "foo", "ffffff"
	owner, repo := "seiffert", "ghsync"

	service := &mockLabelService{}
	labels := map[string]string{}

	if err := ensureLabel(name, "#" + color, owner, repo, labels, service); err != nil {
		t.Fatalf("should not return an error: %s", err)
	}
	if len(service.Labels) != 1 {
		t.Fatal("should have created one label")
	}
	label := service.Labels[0]
	if *label.Name != name {
		t.Fatalf("label name does not match: %s != %s", *label.Name, name)
	}
	if *label.Color != color {
		t.Fatalf("label color does not match: %s != %s", *label.Color, color)
	}
}

func TestEnsureLabel_Update(t *testing.T) {
	name, oldColor, newColor := "foo", "ffffff", "000000"
	owner, repo := "seiffert", "ghsync"

	service := &mockLabelService{
		Labels: []*github.Label{{
			Name: github.String(name),
			Color: github.String(oldColor),
		}},
	}

	labels := map[string]string{
		name: oldColor,
	}

	if err := ensureLabel(name, newColor, owner, repo, labels, service); err != nil {
		t.Fatalf("should not return an error: %s", err)
	}
	if len(service.Labels) != 1 {
		t.Fatal("should have created one label")
	}
	label := service.Labels[0]
	if *label.Name != name {
		t.Fatalf("label name does not match: %s != %s", *label.Name, name)
	}
	if *label.Color != newColor {
		t.Fatalf("label color does not match: %s != %s", *label.Color, newColor)
	}
}

func TestEnsureLabel_AlreadyExists(t *testing.T) {
	name, color := "foo", "ffffff"
	owner, repo := "seiffert", "ghsync"

	service := &mockLabelService{
		Labels: []*github.Label{{
			Name: github.String(name),
			Color: github.String(color),
		}},
	}

	labels := map[string]string{
		name: color,
	}

	if err := ensureLabel(name, color, owner, repo, labels, service); err != nil {
		t.Fatalf("should not return an error: %s", err)
	}
	if len(service.Labels) != 1 {
		t.Fatal("should have created one label")
	}
	label := service.Labels[0]
	if *label.Name != name {
		t.Fatalf("label name does not match: %s != %s", *label.Name, name)
	}
	if *label.Color != color {
		t.Fatalf("label color does not match: %s != %s", *label.Color, color)
	}
}

type mockLabelService struct {
	Labels []*github.Label
}

func (s *mockLabelService) ListLabels(owner, repo string, opt *github.ListOptions) ([]*github.Label, *github.Response, error) {
	if opt == nil {
		opt = &github.ListOptions{}
	}

	start := opt.Page * itemsPerPage
	end := int(math.Min(float64(opt.Page*itemsPerPage+itemsPerPage), float64(len(s.Labels))))

	nextPage := opt.Page + 1
	if end == len(s.Labels) {
		nextPage = 0
	}

	return s.Labels[start:end], &github.Response{NextPage: nextPage}, nil
}

func (s *mockLabelService) CreateLabel(owner, repo string, label *github.Label) (*github.Label, *github.Response, error) {
	s.Labels = append(s.Labels, label)
	return label, nil, nil
}

func (s *mockLabelService) EditLabel(owner, repo, name string, label *github.Label) (*github.Label, *github.Response, error) {
	for _, l := range s.Labels {
		if *l.Name == name {
			l.Color = label.Color
		}
	}
	return label, nil, nil
}
