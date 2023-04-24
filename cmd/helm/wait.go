/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"helm.sh/helm/v3/cmd/helm/require"
	"helm.sh/helm/v3/pkg/action"
)

const waitDesc = `
This command waits for a release to be ready.
`

func newWaitCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	client := action.NewWait(cfg)

	cmd := &cobra.Command{
		Use:   "wait <RELEASE>",
		Short: "wait for a release to be ready",
		Long:  waitDesc,
		Args:  require.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return compListReleases(toComplete, args, cfg)
			}

			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.Run(args[0]); err != nil {
				return err
			}

			fmt.Fprintf(out, "Wait was a success! Happy Helming!\n")
			return nil
		},
	}

	f := cmd.Flags()
	f.DurationVar(&client.Timeout, "timeout", 300*time.Second, "time to wait for any individual Kubernetes operation (like Jobs for hooks)")
	f.BoolVar(&client.WaitForJobs, "wait-for-jobs", false, "if set, will wait until all Jobs have been completed before declaring the release as ready. It will wait for as long as --timeout")

	return cmd
}
