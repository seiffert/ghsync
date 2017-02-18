# ghsync

[![Build Status](https://travis-ci.org/seiffert/ghsync.svg?branch=master)](https://travis-ci.org/seiffert/ghsync)
[![Go Report Card](https://goreportcard.com/badge/github.com/seiffert/ghsync)](https://goreportcard.com/report/github.com/seiffert/ghsync)
[![Code Climate](https://codeclimate.com/github/seiffert/ghsync/badges/gpa.svg)](https://codeclimate.com/github/seiffert/ghsync)

**ghsync** is a CLI tool that allows you to sync configuration specified in a configuration file to multiple GitHub 
repositories. This is particularly useful when you work with many repositories in your organization that should all 
follow certain standard configurations. 

## Example Usage

```bash
$ cat ghrepos.yaml
labels:
  foo: '#ffffff'
  bar: '#000000'
  
milestones:
  - title: Version 0.1
    description: All issues that need to be done before release 0.1
    due: 2017-02-06T12:00:00Z
    state: closed
  - title: Version 1.0
    description: All issues that need to be done before release 1.0
    due: 2017-03-02T12:00:00Z
    state: open
  - title: Version 1.1
    description: All issues that need to be done before release 1.1
    due: 2017-03-03T12:00:00Z
    state: open
$ echo seiffert/ghsync | ghsync --config ghsync.yaml
Syncing Labels
Processing repository "seiffert/ghsync"
  Found 7 labels
  Creating label "foo": "ffffff"
  Creating label "bar": "000000"
  
Syncing Milestones
Processing repository "seiffert/ghsync"
  Found 1 milestones
  Creating milestone "Version 1.0"
  Creating milestone "Version 1.1"
```

To authenticate against the GitHub API, **ghsync** requires a GitHub access token.
Generate a token one in [your account settings](https://github.com/settings/tokens) and pass it either as environment 
variable `GITHUB_TOKEN` or via the `--token` option:

```bash
$ echo seiffert/ghsync | ghsync --token <GITHUB_TOKEN> --config ghsync.yaml
```

Instead of `echo`ing the repositories manually, you can use my tool [**ghrepos**](https://github.com/seiffert/ghrepos)
to get a list of repositories with a specific topic:

```
$ ghrepos -o seiffert github | ghsync --config ghsync.yaml
Processing repository "seiffert/ghrepos"
  Found 7 labels
  Creating label "bar": "000000"
  Creating label "foo": "ffffff"
Processing repository "seiffert/ghsync"
  Found 9 labels
  Label "foo": "ffffff" already exists
  Label "bar": "000000" already exists
```

## Installation

To install **ghsync**, download a binary from the provided
[GitHub releases](https://github.com/seiffert/ghsync/releases) and put it into a folder that is part of your 
system's `$PATH`.

## Contribution

If you have ideas for improving this little tools or just a question, please don't hesitate to open an
[issue](https://github.com/seiffert/ghsync/issues/new) or even fork this repository and create a pull-request!
