// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package cmd

import (
	"fmt"
	"github.com/rayburgemeestre/jirahours/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-git.v4"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

func init() {
	rootCmd.AddCommand(fetchCmd)
}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch all remotes on all repositories",
	Long:  `Runs git fetch --all in each repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		repos := viper.GetStringSlice("repositories")
		fmt.Println("Reading", len(repos), "repositories.")
		wg = sync.WaitGroup{}

		// Running fetch concurrently apparently is not a good idea:
		// ...
		// ERROR IN: /home/ray/projects/cmdaemon/cluster-tools reference has changed concurrently
		// error: reference has changed concurrently
		// exit status 1
		//
		// It seems able to corrupt repositories:
		// bash$ git fetch upstream
		// error: cannot lock ref 'refs/remotes/upstream/8.0': unable to resolve reference 'refs/remotes/upstream/8.0': reference broken
		// From ssh://bitbucket.brightcomputing.com:7999/cm/cluster-tools
		// ! [new branch]          8.0        -> upstream/8.0  (unable to update local ref)
		// error: cannot lock ref 'refs/remotes/upstream/8.2': unable to resolve reference 'refs/remotes/upstream/8.2': reference broken
		// ! [new branch]          8.2        -> upstream/8.2  (unable to update local ref)
		//
		// I screwed up a few repositories this way so I'm pretty sure there are races when you do "writes" to different
		// repositories in concurrently running go routines.
		//
		// EDIT: I don't know how relevant for above, but I also found this general remark: https://github.com/src-d/go-git/issues/702

		for _, path := range repos {
			wg.Add(1)
			// ...Hence no `go` keyword here!
			fetchAll(path)
		}
		wg.Wait()
	},
}

func fetchAll(path string) {
	defer wg.Done()

	key := viper.GetString("ssh.key")
	key = strings.Replace(key, "$HOME", os.Getenv("HOME"), 1)
	sshKey, err := ioutil.ReadFile(key)
	util.CheckIfError(err)

	signer, err := ssh.ParsePrivateKey([]byte(sshKey))
	util.CheckIfError(err)

	r, err := git.PlainOpen(path)
	util.CheckIfError(err)

	remotes, err := r.Remotes()
	util.CheckIfError(err)

	for _, remote := range remotes {
		urls := remote.Config().URLs
		if len(urls) == 0 {
			continue
		}
		fmt.Println("Fetching remote", remote.Config().Name, urls)
		var err error
		if strings.Contains(urls[0], "ssh://") {
			err = r.Fetch(&git.FetchOptions{
				RemoteName: remote.Config().Name,
				Auth:       &gitssh.PublicKeys{User: "git", Signer: signer},
			})
		} else {
			err = r.Fetch(&git.FetchOptions{
				RemoteName: remote.Config().Name,
			})
		}
		if err == git.NoErrAlreadyUpToDate {
			fmt.Println("Remote was already up-to-date.")
		} else if err != nil {
			fmt.Println("ERROR IN:", path, err.Error())
			util.CheckIfError(err)
		}
	}
}
