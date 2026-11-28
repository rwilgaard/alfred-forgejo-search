# alfred-forgejo-search

Alfred workflow for searching Forgejo repositories, issues, and pull requests. Also compatible with Gitea, as the workflow uses the Gitea SDK.

## Requirements

- [Alfred 6](https://www.alfredapp.com/) with Powerpack (the details, login and error windows use Alfred UI)
- A Forgejo instance with API access

## Installation

1. Download the latest `.alfredworkflow` from [Releases](https://github.com/rwilgaard/alfred-forgejo-search/releases)
2. Double-click to import into Alfred
3. Set `forgejo_url` in workflow configuration (no trailing slash, e.g. `https://git.example.com`)
4. Trigger `fg` → press ⏎ on "You're not logged in" → paste your access token into the login window and press ⏎ (the token is checked against Forgejo before it's stored in your Keychain)

### Creating an API token

1. Go to your Forgejo instance → **Settings** → **Applications** (the login window's **Create a token…** button opens this page)
2. Under **Manage Access Tokens**, enter a token name (e.g. `alfred`)
3. Select **Public only** or **All repositories** depending on your needs
4. Enable **Issue → Read**, **Repository → Read**, and **User → Read**
5. Click **Generate Token** and copy it — it won't be shown again

## Usage

### `fg [query]` — search repositories

Type to filter your repositories. Repositories you use often rank higher, both with an empty query and among equally good matches.

| Key | Action |
|-----|--------|
| ⏎ | Open the repository in the browser |
| ⌘⏎ | Open the repository's details window |

⎋ in the details window returns to the search with your query kept.

### Details window

Shows the repository's description and details, and its open issues or pull requests (first 50). Every action has a button, labelled with its shortcut:

| Button | Shortcut | Result |
|--------|----------|--------|
| Issues / Pull requests | ⌘1 / ⌘2 | Switch the list |
| Preview | ⌘Y | Show the selected item's description; ↩ in the preview opens it |
| Open in Browser | ⌘O | Open the repository |
| Copy HTTPS / Copy SSH | ⌘C / ⌥⌘C | Copy a clone URL |

Type to filter the list, ↑/↓ to select, ↩ or double-click to open the selected issue or pull request.

### Errors

If refreshing the repository cache in the background fails, an error window shows the full message. ↩ retries, ⌘L opens the login window, ⌘C copies the message.

To remove the stored token, type `workflow:clearauth` after the keyword.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `forgejo_url` | — | Base URL of your Forgejo instance |
| `repo_keyword` | `fg` | Alfred keyword to trigger the workflow |
| `cache_age` | `360` | Repository cache TTL in minutes |
| `check_for_updates` | on | Check GitHub for a new version once a day |

## Development

```sh
make build          # build arch-specific binaries
make dev            # build, install into Alfred in place, and reload
make package-alfred # build + zip into .alfredworkflow
make zip-alfred     # zip into .alfredworkflow without rebuilding (for CI/CD)
make fmt            # format with gofumpt
make release V=x.y.z        # bump version, package, and tag locally
```

Requires Go 1.27.4+ and `golangci-lint`.
