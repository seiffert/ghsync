package main

import (
	"fmt"
	"strings"
)

func newRepository(in string) repository {
	parts := strings.Split(in, "/")
	return repository{
		Owner: parts[0],
		Name:  parts[1],
	}
}

type repository struct {
	Owner string
	Name  string
}

func (r repository) String() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}
