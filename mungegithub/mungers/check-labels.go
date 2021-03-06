/*
Copyright 2015 The Kubernetes Authors.

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

package mungers

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/test-infra/mungegithub/features"
	githubhelper "k8s.io/test-infra/mungegithub/github"
	"k8s.io/test-infra/mungegithub/mungers/mungerutil"
	"k8s.io/test-infra/mungegithub/options"

	"bytes"
	"io/ioutil"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

type labelAccessor interface {
	AddLabel(label *github.Label) error
	GetLabels() ([]*github.Label, error)
}

// CheckLabelsMunger will check that the labels specified in the labels yaml file
// are created.
type CheckLabelsMunger struct {
	labelFilePath string
	prevHash      string
	labelAccessor labelAccessor
	features      *features.Features
	readFunc      func() ([]byte, error)
}

func init() {
	RegisterMungerOrDie(&CheckLabelsMunger{})
}

// Name is the name usable in --pr-mungers
func (c *CheckLabelsMunger) Name() string { return "check-labels" }

// RequiredFeatures is a slice of 'features' that must be provided.
func (c *CheckLabelsMunger) RequiredFeatures() []string { return []string{features.RepoFeatureName} }

// Initialize will initialize the munger.
func (c *CheckLabelsMunger) Initialize(config *githubhelper.Config, features *features.Features) error {
	c.labelAccessor = config
	c.features = features
	c.readFunc = func() ([]byte, error) {
		return ioutil.ReadFile(c.labelFilePath)
	}

	return c.validateFilePath()
}

func (c *CheckLabelsMunger) validateFilePath() error {
	if len(c.labelFilePath) == 0 {
		return fmt.Errorf("no 'label-file' option specified, cannot check labels")
	}
	if _, err := os.Stat(c.labelFilePath); os.IsNotExist(err) {
		return fmt.Errorf("failed to stat the check label config: %v", err)
	}
	return nil
}

// EachLoop is called at the start of every munge loop
func (c *CheckLabelsMunger) EachLoop() error {
	fileContents, err := c.readFunc()
	if err != nil {
		glog.Errorf("Failed to read the check label config: %v", err)
		return err
	}
	hash := mungerutil.GetHash(fileContents)
	if c.prevHash != hash {
		// Get all labels from file.
		fileLabels := map[string][]*github.Label{}
		if err := yaml.NewYAMLToJSONDecoder(bytes.NewReader(fileContents)).Decode(&fileLabels); err != nil {
			return fmt.Errorf("Failed to decode the check label config: %v", err)
		}

		// Get all labels from repository.
		repoLabels, err := c.labelAccessor.GetLabels()
		if err != nil {
			return err
		}
		c.addMissingLabels(repoLabels, fileLabels["labels"])
		c.prevHash = hash
	}
	return nil
}

// addMissingLabels will not remove any labels. It will add those which are present in the yaml file and not in
// the repository.
func (c *CheckLabelsMunger) addMissingLabels(repoLabels, fileLabels []*github.Label) {
	repoLabelSet := sets.NewString()
	for _, repoLabel := range repoLabels {
		repoLabelSet.Insert(*repoLabel.Name)
	}

	// Compare against labels in local file.
	for _, label := range fileLabels {
		if !repoLabelSet.Has(*label.Name) {
			err := c.labelAccessor.AddLabel(label)
			if err != nil {
				glog.Errorf("Error %s in adding label %s", err, *label.Name)
			}
		}
	}
}

// RegisterOptions registers options for this munger; returns any that require a restart when changed.
func (c *CheckLabelsMunger) RegisterOptions(opts *options.Options) sets.String {
	opts.RegisterString(&c.labelFilePath, "label-file", "", "Path from repository root to file containing list of labels.")
	opts.RegisterUpdateCallback(func(changed sets.String) error {
		if changed.Has("label-file") {
			return c.validateFilePath()
		}
		return nil
	})
	return nil
}

// Munge is unused by this munger.
func (c *CheckLabelsMunger) Munge(obj *githubhelper.MungeObject) {}
