# Release Dash ![build](https://github.com/lobsterdore/release-dash/workflows/Docker/badge.svg)

[![release-dash homepage](/docs/screenshot.png)](https://dash.techpunch.co.uk/)

A dashboard for visualising commits in pipelines so developers know what will go out.
Useful for pipelines that have non-prd environments with a manual gate, developers can
check the dashboard to see what will go out when a release is pushed from one
environment to another.

The dashboard requires a Github PAT to run, any repo that the PAT has read access to
will be inspected, if a ```.releasedash.yml``` file is found in the root of the repo
then it will appear on the board. For each repo a list of tags is supplied via the
```.releasedash.yml``` config file, these tags are diffed to produce a changelog.

## How to run

### Requirements

* GO 1.14>
* Github PAT

### Native

```BASH
# Assuming direnv is installed
echo "export GITHUB_PAT=\"[PAT]\"" > .envrc
direnv allow
make run
```

### Docker

```BASH
# Assuming direnv is installed
echo "export GITHUB_PAT=\"[PAT]\"" > .envrc
direnv allow
make docker_run
```

### Kubes

A [basic helm chart](https://github.com/lobsterdore/release-dash-helm) if available
for running on Kubernetes.

### Configuration via Environment vars

For a full list of available environment vars see [config/configuration.go](config/configuration.go).

Some key env vars are:

* ```GITHUB_PAT``` the Github Personal Access Token used to read repos
* ```GITHUB_CHANGELOG_FETCH_TIMER_SECONDS``` (default 60) controls the interval for grabbing
the changelog for each repository that should appear on the dashboard
* ```GITHUB_REPO_FETCH_TIMER_SECONDS``` (default 600) controls the interval for looking up repos
that should appear on the dashboard
* ```LOGGING_LEVEL``` (default error) verbosity of logs, set to debug to see detailed info on each
repo read by the dash

## How to register repos and commits

When started the service will kick of two background processes, one to grab a list of
repos that should appear on the dashboard and another to grab the changelogs for
all of these repos.

For a repo and it's commits to appear on the dashboard the following criteria must
be met:

* Readable via the Github Personal Access Token supplied to the dashboard
* A file called '.releasedash.yml' exists in the repos root with a list of
environment tags
* Environment tags must exist to perform diffs

### Accessiable via GH PAT

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

environment_tags: [ LIST OF ENVIRONMENT TAGS ]
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

The ```environment_tags``` list powers the changelog, the release dash will do
a diff between each tag starting from the first one. In the example above the
release dash will display the following diffs:

```BASH
dev -> stg
stg -> prd
```

The changelog for all repos is fetched via a background task on a regular tick
interval which can be controlled via the ```GITHUB_CHANGELOG_FETCH_TIMER_SECONDS```
env var in [config/configuration.go](config/configuration.go)).

### Environment tags

As noted in the [Configuration via .releasedash.yml](#configuration-via-releasedashyml)
section, tags are used to perform diffs between different environments, this
diff is then used to show any changes that are in place between environments for a
service repo.

The recommended way to maintain release tags is to force update them after a
release of a service to a given environment, the flow of a release pipeline might
look something like this:

* Build and test service repo
* Deploy sevice to dev environment
* Force update ```dev``` tag in repo post deployment
* Repeat the same process for other envs

To force push an environment tag you can do the following:

```BASH
git tag -a -f dev v0.0.1 -m "Updating dev to v0.0.1"
git push --force origin tag dev
```
