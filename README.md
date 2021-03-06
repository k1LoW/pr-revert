# pr-revert

pr-revert is a tool for reverting pull requests.

**:octocat: GitHub Actions for pr-revert is [here](https://github.com/k1LoW/pr-revert-action) !!**

## Usage

### Default behavior

1. Clone repository from remote to temporary working directory.
2. Create a branch to stack revert commits.
3. Revert pull requests.
4. Push branch to remote repository.
5. Create a pull request.
6. Cleanup temporary working directory.

### Revert a pull request by number (#3)

``` console
$ pr-revert --number 3
```

### Revert latest 5 merged pull requests

``` console
$ pr-revert --latest 5
```

### Revert all merged pull requests until 2 hours

``` console
$ pr-revert --until '2 hours'
```

### Revert pull requests without new pull request

``` console
$ pr-revert --latest 3 --no-pull-request
```

### Revert pull requests without cleanup

``` console
$ pr-revert --latest 3 --no-cleanup
```

### Revert pull requests and push to default branch of remote repository

``` console
$ pr-revert --latest 3 --no-branch
```

## Requirements

- Git

### Required Environment Variables

| Environment variable | Description | Default |
| --- | --- | --- |
| `GITHUB_TOKEN` | A GitHub access token. | - |
| `GITHUB_REPOSITORY` | The owner and repository name ( `owner/repo` )| - |
| `GITHUB_SERVER_URL` | The the GitHub server URL. | `https://github.com` |
| `GITHUB_API_URL` | The GitHub API URL | `https://api.github.com` |
| `GITHUB_GRAPHQL_URL` | The GitHub GraphQL API URL | `https://api.github.com/graphql` |

## Environment Variables

| Environment variable | Description |
| --- | --- |
| `PR_REVERT_LATEST` | Can be used instead of the `--latest` option |
| `PR_REVERT_UNTIL` | Can be used instead of the `--until` option |
| `PR_REVERT_NUMBER` | Can be used instead of the `--number` option |
| `PR_REVERT_NO_PUSH` | Can be used instead of the `--no-push` option |
| `PR_REVERT_NO_PULL_REQUEST` | Can be used instead of the `--no-pull-request` option |
| `PR_REVERT_NO_CLEANUP` | Can be used instead of the `--no-cleanup` option |
| `PR_REVERT_NO_BRANCH` | Can be used instead of the `--no-branch` option |

## Install

**deb:**

Use [dpkg-i-from-url](https://github.com/k1LoW/dpkg-i-from-url)

``` console
$ export PR-REVERT_VERSION=X.X.X
$ curl -L https://git.io/dpkg-i-from-url | bash -s -- https://github.com/k1LoW/pr-revert/releases/download/v$PR-REVERT_VERSION/pr-revert_$PR-REVERT_VERSION-1_amd64.deb
```

**RPM:**

``` console
$ export PR-REVERT_VERSION=X.X.X
$ yum install https://github.com/k1LoW/pr-revert/releases/download/v$PR-REVERT_VERSION/pr-revert_$PR-REVERT_VERSION-1_amd64.rpm
```

**apk:**

Use [apk-add-from-url](https://github.com/k1LoW/apk-add-from-url)

``` console
$ export PR-REVERT_VERSION=X.X.X
$ curl -L https://git.io/apk-add-from-url | sh -s -- https://github.com/k1LoW/pr-revert/releases/download/v$PR-REVERT_VERSION/pr-revert_$PR-REVERT_VERSION-1_amd64.apk
```

**homebrew tap:**

```console
$ brew install k1LoW/tap/pr-revert
```

**manually:**

Download binary from [releases page](https://github.com/k1LoW/pr-revert/releases)

**go get:**

```console
$ go get github.com/k1LoW/pr-revert
```

**docker:**

```console
$ docker pull ghcr.io/k1low/pr-revert:latest
```
