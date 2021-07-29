# pr-revert

pr-revert is a tool for reverting pull requests.

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
