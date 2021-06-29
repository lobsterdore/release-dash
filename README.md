# Release Dash

![build](https://github.com/lobsterdore/release-dash/workflows/Docker/badge.svg)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
[![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/lobsterdore/release-dash?style=flat)](https://github.com/lobsterdore/release-dash/tags)

> A dashboard for visualising commits in pipelines.

[![release-dash homepage](/docs/screenshot.png)](https://dash.techpunch.co.uk/)

## What it does

Designed for pipelines that have multiple environments with manual gates, developers
can check the dashboard to see what will go out when a release is pushed from one
environment to another.

The dashboard requires a Github PAT to run, any repo that the PAT has read access to
will be inspected, if a ```.releasedash.yml``` file is found in the root of the repo
then it will appear on the board. For each repo a list of tags or branches is supplied
via the ```.releasedash.yml``` config file, these tags/branches are diffed to produce
a changelog of commits.

## Useful links

* [Demo site](https://dash.techpunch.co.uk/)
* [Docker images](https://hub.docker.com/r/lobsterdore/release-dash)
* [Example of a registered repo with tags](https://github.com/lobsterdore/release-dash-test-repo-1)
* [Example of a registered repo with branches](https://github.com/lobsterdore/release-dash-test-repo-2)

## Use case

This dashboard is intended to show developers, and potentially people that manage
developers, commits that are sitting in environments waiting to be released. It's useful for:

* Ensuring that commits are pushed to production regularly
* Showing a pile up of unreleased commits
* Checking if changes needs to be pushed before committing more changes

The dashboard makes an assumption that the commit history of your repos is readable by people
other than the developer that committed a given change, some discipline is required on the part
of committers to produce legible messages.

The dashboard will show all commits between tags/branches, this will include features, chores,
tech debt and bug fixes, this makes the dashboard suitable as an internal tool only since it will
show commits that are not useful for people that have no investment in the inner workings of a
project and it's repository.

## How to run

### Requirements

* GO 1.14> (for native)
* Github PAT

### Docker

```BASH
# Assuming direnv is installed
echo "export GITHUB_PAT=\"[PAT]\"" > .envrc
direnv allow
# See https://hub.docker.com/r/lobsterdore/release-dash for versions
docker run -dit -e GITHUB_PAT -p 8080:8080 lobsterdore/release-dash:[VERSION]
```

### Kubernetes

A [basic helm chart](https://github.com/lobsterdore/release-dash-helm) is available
for running on Kubes.

### Binary releases

Download a OS specific binary from
[https://github.com/lobsterdore/release-dash/releases](https://github.com/lobsterdore/release-dash/releases),
you will need to setup environment vars as appropriate for the target OS.

### Native

```BASH
git clone git@github.com:lobsterdore/release-dash.git
# Assuming direnv is installed
echo "export GITHUB_PAT=\"[PAT]\"" > .envrc
direnv allow
make run
```

### Configuration

The following environment variables are available:

|Variable|Default|Description|
|---|---|---|
|CACHE_CLEANUP_INTERVAL_SECONDS|300|Time between cache purges, see [https://github.com/patrickmn/go-cache](https://github.com/patrickmn/go-cache)|
|CACHE_DEFAULT_EXPIRATION_SECONDS|1800|Time to keep cached Repo and Changelog data for, should be greater than fetch timers|
|GITHUB_CHANGELOG_FETCH_TIMER_SECONDS|180|Time between fetches of diffs for each repo and environment|
|GITHUB_PAT|~|Github Personal Access Token used to read repos|
|GITHUB_REPO_FETCH_TIMER_SECONDS|900|Time between fetches of repo list|
|GITHUB_URL_DEFAULT|~|URL for Github API, defaults to standard Github API URL|
|GITHUB_URL_UPLOAD|~|URL for Github Uploads, defaults to standard Github Upload URL|
|LOGGING_LEVEL|error|Level for logs, see [https://github.com/rs/zerolog](https://github.com/rs/zerolog)|
|SERVER_HOST|0.0.0.0|Host to bind web server to|
|SERVER_PORT|8080|Port to bind web server to|
|SERVER_TIMEOUT_IDLE|65|Idle timeout for connections|
|SERVER_TIMEOUT_WRITE|10|Write timeout for connections|
|SERVER_TIMEOUT_SERVER|10|Overall timeout for connections|
|SERVER_TIMEOUT_READ|10|Read timeout for connections|

## How to register repos and commits

When started the service will kick of two background processes, one to grab a list of
repos that should appear on the dashboard and another to grab the changelogs for
all of these repos.

For a repo and it's commits to appear on the dashboard the following criteria must
be met:

* Readable via the Github Personal Access Token supplied to the dashboard
* A file called '.releasedash.yml' exists in the repos root with a list of
environment tags or environment branches
* Environment tags or environment branches must exist to perform diffs

### Accessible via GH PAT

The Github Personal Access Token added to this service needs read access to
all repos that should appear on the dashboard. The list of repos is fetched
via a background task on a regular tick interval, the interval can be
controlled via the ```GITHUB_REPO_FETCH_TIMER_SECONDS``` env var in
[config/configuration.go](config/configuration.go)).

### Configuration via releasedash.yml

A ```.releasedash.yml``` file needs to exist in the root of a repo, please see
the [example file here](https://github.com/lobsterdore/release-dash-test-repo-1/blob/main/.releasedash.yml).

The layout of the file should be like so:

```YAML
---
# Using tags for the changelog diffs

environment_tags: [ LIST OF ENVIRONMENT TAGS ]
name: [NAME OF SERVICE ON DASHBOARD]
```

```YAML
---
# Using branches for the changelog diffs

environment_branches: [ LIST OF ENVIRONMENT TAGS ]
name: [NAME OF SERVICE ON DASHBOARD]
```

For example:

```YAML
---

environment_tags:
  - dev
  - stg
  - prd
name: release-dash-test-repo-1
```

```YAML
---

environment_branches:
  - development
  - pre-prod
  - prod
name: release-dash-test-repo-2
```

The ```environment_tags``` and ```environment_branches``` lists power the
changelog, the release dash will do a diff between each tag/branch starting from
the first one. In the examples above the release dash will display the following
diffs:

```BASH
dev -> stg
stg -> prd
```

```BASH
development -> pre-prod
pre-prod -> prod
```

The changelog for all repos is fetched via a background task on a regular tick
interval which can be controlled via the ```GITHUB_CHANGELOG_FETCH_TIMER_SECONDS```
env var in [config/configuration.go](config/configuration.go)).

### Diffs via environment tags or environment branches

As noted in the [Configuration via .releasedash.yml](#configuration-via-releasedashyml)
section, tags/branches are used to perform diffs between different environments, this
diff is then used to show any changes that are in place between environments for a
service repo.

Each environment should have a tag or branch in each repository registered with the release-dash,
for example your dev environment should have a 'dev' tag in each repo, this tag should
match the commit hash of the version that is deployed to a given environment, if you just deployed
hash bb5ac8 to dev then your dev tag should point to bb5ac8. These tags/branches allow for easy
diffing between environments without having to know which version numbers are deployed.

The recommended way to maintain environment tags/branches is to force update them after a
deployment of a service to a given env, the flow of a release pipeline might
look something like this if environment_tags are in use:

* Build and test service repo
* Deploy service to dev environment
* Force update ```dev``` tag in repo post deployment
* Repeat the same tagging process for other envs

To force push an environment tag you can do the following:

```BASH
git tag -a -f dev v0.0.1 -m "Updating dev to v0.0.1"
git push --force origin tag dev
```

A similar process can be followed if environment_branches are in use for a given repo.
