# release-dash

![docker](https://github.com/lobsterdore/release-dash/workflows/Docker/badge.svg)

A dashboard for visualising commits in pipelines so developers know what will go out.
Useful for pipelines that have non-prd environments with a manual gate.

![release-dash homepage](/docs/screenshot.png)

## How to run

For a full list of available environment vars see [config/configuration.go](config/configuration.go).

### Requirements

* GO 1.14
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
