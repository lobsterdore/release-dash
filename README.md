# release-dash

Very much a WIP.

Will be a dashboard for visualising commits in pipelines so developers know what will go out.
Useful for pipelines that have non-prd environments with a manual gate.

## How to run

For a full list of available environment vars see [config/configuration.go](config/configuration.go).

### Requirements

* GO 1.14
* Github PAT

### Native

```
# Assuming direnv is installed
echo "export GITHUB_PAT=\"[PAT]\"" > .envrc
direnv allow
make run
```

### Docker

This method is not fully working yet, I need to figure out what to do with the HTML templates and
the binary.

```
# Assuming direnv is installed
echo "export GITHUB_PAT=\"[PAT]\"" > .envrc
direnv allow
make docker_run
```
