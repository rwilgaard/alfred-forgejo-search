# alfred-forgejo-search

Alfred workflow for searching Forgejo repositories, issues, and pull requests. Also compatible with Gitea, as the workflow uses the Gitea SDK.

## Requirements

- [Alfred](https://www.alfredapp.com/) with Powerpack
- A Forgejo instance with API access

## Installation

1. Download the latest `.alfredworkflow` from [Releases](https://github.com/rwilgaard/alfred-forgejo-search/releases)
2. Double-click to import into Alfred
3. Set `forgejo_url` in workflow configuration (no trailing slash, e.g. `https://git.example.com`)
4. Trigger `fg` → press ⏎ on "You're not logged in" → enter your API token

### Creating an API token

1. Go to your Forgejo instance → **Settings** → **Applications**
2. Under **Manage Access Tokens**, enter a token name (e.g. `alfred`)
3. Select **Public only** or **All repositories** depending on your needs
4. Enable **Issue → Read**, **Repository → Read**, and **User → Read**
5. Click **Generate Token** and copy it — it won't be shown again

## Usage

### `fg [query]` — search repositories

| Action | Result |
|--------|--------|
| ⏎ | Open repository in browser |
| ⌘⏎ | Copy HTTPS clone URL |
| ⌥⏎ | Copy SSH clone URL |
| ⌃⏎ | Show open issues |
| ⇧⏎ | Show open pull requests |

### Issues / Pull Requests sub-menu

Type to filter. Select item to open in browser. Select `← owner/repo` to go back to repository list (restores previous search query).

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `forgejo_url` | — | Base URL of your Forgejo instance |
| `repo_keyword` | `fg` | Alfred keyword to trigger the workflow |
| `cache_age` | `360` | Repository cache TTL in minutes |

## Development

```sh
make build          # build arch-specific binaries
make package-alfred # build + zip into .alfredworkflow
make fmt            # format with gofumpt
make release VERSION=x.y.z  # bump version, package, tag, push
```

Requires Go 1.26+ and `golangci-lint`.
