package main

import (
	"context"
	"flag"
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

// collectAllRepo objects for a given organization and send them to a channel to be processed on.
func collectAllRepos(ctx context.Context, client *github.Client, org string, repositories chan<- *github.Repository, perPage int) error {
	// get all pages of results
	opt := &github.RepositoryListByOrgOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: perPage},
	}
	defer close(repositories)
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return err
		}
		for _, repo := range repos {
			repositories<- repo
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return nil
}

// fetchRepos fetches all the repositories
func fetchRepos(ctx context.Context, client *github.Client, org string, repositories chan<- *github.Repository, perPage int) {
	fmt.Printf("Gently fetching list of all %s repos... may take a minute\n", org)
	err := collectAllRepos(ctx, client, org, repositories, perPage)
	if err != nil {
		log.Fatalf("fatal error calling github api: %v", err)
	}
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

// Worker clones all the repositories.
func worker(repositories <-chan *github.Repository, path string) {
	for repo := range repositories {
		err := gitClone(*repo.Name, *repo.SSHURL, path)
		if err != nil {
			fmt.Printf("clone failed, giving up on everything now: %v\n", err)
			os.Exit(2)
		}
	}
}

func main() {
	workers := flag.Int("workers", 1, "pass this flag to suggest how many git repos should be pulled in parallel (max is 10)")

	flag.Parse()

	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		log.Fatalf("Specify GITHUB_TOKEN environment variable")
	}
	if len(os.Args) < 3 {
		fmt.Println("Usage: bulkclone <github organization> <path to clone repos to>")
		os.Exit(2)
	}

	if *workers > 10 {
		log.Fatal("Limiting the maximum number of workers to 10")
	}

	org := flag.Arg(0)
	path := flag.Arg(1)

	err := os.MkdirAll(path, 0755)
	if err != nil {
		fmt.Printf("Unable to create clone directory: %v\n", err)
		os.Exit(2)
	}

	ctx := context.Background()
	client := newClient(ctx, token)

	perPage := 99
	repositories := make(chan *github.Repository, perPage)

	// Setup the workers.  They will start blocked.
	for w := 1; w <= *workers; w++ {
		go worker(repositories, path)
	}

	fetchRepos(ctx, client, org, repositories, perPage)

}
