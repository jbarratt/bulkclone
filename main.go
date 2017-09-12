package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// newClient creates a new github client configured to use the token provided
func newClient(ctx context.Context, token string) *github.Client {

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	return client
}

// allRepos returns all the repo objects for a given organization
func allRepos(ctx context.Context, client *github.Client, org string) ([]*github.Repository, error) {
	// get all pages of results
	opt := &github.RepositoryListByOrgOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 99},
	}
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return allRepos, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRepos, nil
}

// gitClone runs a git clone of a url in the provided directory (cloneRoot)
func gitClone(repoName, repoGitURL, cloneRoot string) error {

	dest := path.Join(cloneRoot, repoName)
	_, err := os.Stat(dest)
	if !os.IsNotExist(err) {
		fmt.Printf("%v has already been cloned, skipping\n", dest)
		return nil
	}

	// shallow clone. Someone may want a deep one at some point
	args := []string{"clone", "--depth", "1", repoGitURL}
	cmd := exec.Command("git", args...)
	cmd.Dir = cloneRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	fmt.Printf("Cloning %v into %v...\n", repoName, cloneRoot)
	err = cmd.Run()
	return err
}

func main() {

	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		log.Fatalf("Specify GITHUB_TOKEN environment variable")
	}
	if len(os.Args) < 3 {
		fmt.Println("Usage: bulkclone <github organization> <path to clone repos to>")
		os.Exit(2)
	}

	org := os.Args[1]
	path := os.Args[2]

	err := os.MkdirAll(path, 0755)
	if err != nil {
		fmt.Printf("Unable to create clone directory: %v\n", err)
		os.Exit(2)
	}

	ctx := context.Background()
	client := newClient(ctx, token)

	fmt.Printf("Gently fetching list of all %s repos... may take a minute\n", org)
	repos, err := allRepos(ctx, client, org)

	if err != nil {
		log.Fatalf("fatal error calling github api: %v", err)
	}
	for _, repo := range repos {
		err := gitClone(*repo.Name, *repo.SSHURL, path)
		if err != nil {
			fmt.Printf("clone failed, giving up on everything now: %v\n", err)
			os.Exit(2)
		}
	}
}
