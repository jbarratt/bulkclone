# BulkClone

Simple go tool to that, given an org name, makes a shallow clone all the repositories inside it.

This can be useful if you're exploring an organization's codebase, for doing backups, or ... probably other things? I wrote this joining a new company and wanting to look at languages/style choices in some depth.

This includes any repos for an organization you have access to (public or private).

Requires

* `GITHUB_TOKEN` env var to be set
* org name on the cli
* path to clone into (will be created if needed)

It will try and detect already-cloned repositories and skip them, so you can run it periodically to fetch fresh repos.

### Getting a token

1. Go to the [github token page](https://github.com/settings/tokens)
2. Click 'Generate new Token'
3. Check the first box (**repo** _Full control of private repositories_)
4. Store the token somewhere secure (maybe your 1Password or LastPass vault?)
5. Set the `GITHUB_TOKEN` environment variable

### Running

1. go build -o bulkclone main.go
2. ./bulkclone yourorg ~/work/clones/yourorg

### Future improvements

To be fair, I just wrote this to fill a need, and probably won't hack on it further. The hacky CLI versions of this tool don't work with high repo counts.

But if it almost-works for you, PR's welcome!

Some things that could be improved:

* Optional/flagged deep vs shallow cloning
* Parallelism of clones (right now it's one at a time, github can probably handle a bit more read i/o)
* Overall progress indication ("repo 12 of 27...")


