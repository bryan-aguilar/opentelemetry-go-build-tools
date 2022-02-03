// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crosslink

import (
	"os"
	"path/filepath"

	tools "go.opentelemetry.io/build-tools"
	"golang.org/x/mod/modfile"
)

func writeModules(module moduleInfo) error {
	mfParsed, err := modfile.Parse("go.mod", module.moduleContents, nil)
	if err != nil {
		return err
	}
	//  now overwrite the existing gomod file
	gomodFile, err := mfParsed.Format()
	if err != nil {
		return err
	}
	//write our updated go.mod file
	err = os.WriteFile(module.moduleFilePath, gomodFile, 0700)
	if err != nil {
		return err
	}

	return nil
}

// attempts to identify a go.mod file at the either the path provided
// or at the root of the git repository. This check will fail if no path is
// provided and the tool was called outside of a git repository.
func identifyRootModule(r string) (string, error) {
	var err error
	rootPath := r
	if rootPath == "" {
		rootPath, err = tools.FindRepoRoot()
		if err != nil {
			return "", err
		}
	}

	if _, err := os.Stat(filepath.Join(rootPath, "go.mod")); err != nil {
		return "", err
	}

	// identify and read the root module
	rootModPath := filepath.Join(rootPath, "go.mod")
	rootModFile, err := os.ReadFile(rootModPath)
	if err != nil {
		return "", err
	}
	return modfile.ModulePath(rootModFile), nil
}
