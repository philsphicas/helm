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

package action

import (
	"bytes"
	"time"

	"github.com/pkg/errors"

	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
)

// Wait is the action for waiting for a given release.
//
// It provides the implementation of 'helm wait'.
type Wait struct {
	cfg *Configuration

	Timeout     time.Duration
	WaitForJobs bool
}

// NewWait creates a new Wait object with the given configuration.
func NewWait(cfg *Configuration) *Wait {
	return &Wait{
		cfg: cfg,
	}
}

// Run executes 'helm wait' against the given release.
func (r *Wait) Run(name string) error {
	if err := r.cfg.KubeClient.IsReachable(); err != nil {
		return err
	}

	r.cfg.Log("preparing wait of %s", name)
	currentRelease, err := r.prepareWait(name)
	if err != nil {
		return err
	}

	r.cfg.Log("performing wait of %s", name)
	if _, err := r.performWait(currentRelease); err != nil {
		return err
	}

	return nil
}

// prepareWait finds the current release and prepares to wait for its resources.
func (r *Wait) prepareWait(name string) (*release.Release, error) {
	if err := chartutil.ValidateReleaseName(name); err != nil {
		return nil, errors.Errorf("prepareWait: Release name is invalid: %s", name)
	}

	currentRelease, err := r.cfg.Releases.Last(name)
	if err != nil {
		return nil, err
	}

	r.cfg.Log("waiting for %s (current: v%d)", name, currentRelease.Version)

	return currentRelease, nil
}

func (r *Wait) performWait(targetRelease *release.Release) (*release.Release, error) {
	target, err := r.cfg.KubeClient.Build(bytes.NewBufferString(targetRelease.Manifest), false)
	if err != nil {
		return targetRelease, errors.Wrap(err, "unable to build kubernetes objects from release manifest")
	}

	if r.WaitForJobs {
		if err := r.cfg.KubeClient.WaitWithJobs(target, r.Timeout); err != nil {
			return targetRelease, errors.Wrapf(err, "release %s not ready", targetRelease.Name)
		}
	} else {
		if err := r.cfg.KubeClient.Wait(target, r.Timeout); err != nil {
			return targetRelease, errors.Wrapf(err, "release %s not ready", targetRelease.Name)
		}
	}

	return targetRelease, nil
}
