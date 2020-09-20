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
echo "GITHUB_PAT=\"[PAT]\"" .envrc
direnv allow
go run main.go
```

### Docker

This method is not fully working yet, I need to figure out what to do with the HTML templates and
the binary.

```
# Assuming direnv is installed
echo "GITHUB_PAT=\"[PAT]\"" .envrc
direnv allow
docker build . -t release_dash
docker run -itd  -e GITHUB_PAT -p 8080:8080 release_dash
```

## Todo

* Tests of any kind
* Proper UI
* Binary + Templates shipped together
