package crosslink

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tools "go.opentelemetry.io/build-tools"
	"golang.org/x/mod/modfile"
)

// TODO: Print warning if there are modules that were not found in the repository but there are requirements for them in the dependency tree.
// TODO: Logging to alert user what changes have been made what attempts were made

func Crosslink(rc runConfig) {
	var err error
	defer rc.logger.Sync()

	if rc.RootPath == "" {
		rc.RootPath, err = tools.FindRepoRoot()
		if err != nil {
			panic("Could not find repo root directory")
		}
	}

	if _, err := os.Stat(filepath.Join(rc.RootPath, "go.mod")); err != nil {
		panic("Invalid root directory, could not locate go.mod file")
	}

	// identify and read the root module
	rootModPath := filepath.Join(rc.RootPath, "go.mod")
	rootModFile, err := os.ReadFile(rootModPath)
	if err != nil {
		panic(fmt.Sprintf("Could not read go.mod file in root path: %v", err))
	}
	rootModulePath := modfile.ModulePath(rootModFile)

	graph, err := buildDepedencyGraph(rc, rootModulePath)
	if err != nil {
		panic(fmt.Sprintf("failed to build dependency graph: %v", err))
	}

	for _, moduleInfo := range graph {
		// do not do anything with excluded
		// TODO: Readdress what `excluded` means more concretely. Should crosslink ignore
		// all references to excluded module or only for replacing and pruning?
		// If an exluded module is named should that stop crosslink from making edits to that go.mod file?
		//if _, exists := rc.excludedPaths[modName]; exists {
		//	continue
		//}

		err = insertReplace(&moduleInfo, rc)
		if err != nil {
			panic(fmt.Sprintf("failed to insert replace statements: %v", err))
		}

		err = pruneReplace(rootModulePath, &moduleInfo, rc)

		if err != nil {
			panic(fmt.Sprintf("error pruning replace statements: %v", err))
		}

		err = writeModules(moduleInfo)
		if err != nil {
			panic(fmt.Sprintf("error writing go.mod files: %v", err))
		}
	}
}

func insertReplace(module *moduleInfo, rc runConfig) error {
	// modfile type that we will work with then write to the mod file in the end
	mfParsed, err := modfile.Parse("go.mod", module.moduleContents, nil)
	if err != nil {
		return err
	}

	for reqModule := range module.requiredReplaceStatements {
		// skip excluded
		if _, exists := rc.ExcludedPaths[reqModule]; exists {
			if rc.Verbose {
				rc.logger.Sugar().Infof("Excluded Module %s, ignoring replace", reqModule)
			}
			continue
		}

		localPath, err := filepath.Rel(mfParsed.Module.Mod.Path, reqModule)
		if err != nil {
			return err
		}
		if localPath == "." || localPath == ".." {
			localPath += "/"
		} else if !strings.HasPrefix(localPath, "..") {
			localPath = "./" + localPath
		}
		var loggerStr string
		// see if replace statement already exists for module. Verify if it's the same. If it does not exist then add it.
		// AddReplace should handle all of these conditions in terms of add and/or verifying
		// https://cs.opensource.google/go/go/+/master:src/cmd/vendor/golang.org/x/mod/modfile/rule.go;l=1296?q=addReplace
		if oldReplace, exists := containsReplace(mfParsed.Replace, reqModule); exists {
			if rc.Overwrite {
				loggerStr = fmt.Sprintf("Overwriting: Module: %s Old: %s => %s New: %s => %s", mfParsed.Module.Mod.Path, reqModule, oldReplace.New.Path, reqModule, localPath)
				mfParsed.AddReplace(reqModule, "", localPath, "")
			} else {
				loggerStr = fmt.Sprintf("Replace already exists: Module: %s : %s => %s \n run with -overwrite flag if update is desired", mfParsed.Module.Mod.Path, reqModule, oldReplace.New.Path)
			}
		} else {
			// does not contain a replace statement. Insert it
			loggerStr = fmt.Sprintf("Inserting replace: Module: %s : %s => %s", mfParsed.Module.Mod.Path, reqModule, localPath)
			mfParsed.AddReplace(reqModule, "", localPath, "")
		}
		if rc.Verbose {
			rc.logger.Sugar().Info(loggerStr)
		}

	}
	module.moduleContents, err = mfParsed.Format()
	if err != nil {
		return err
	}

	return nil
}

// Identifies if a replace statement already exists for a given module name
func containsReplace(replaceStatments []*modfile.Replace, modName string) (*modfile.Replace, bool) {
	for _, repStatement := range replaceStatments {
		if repStatement.Old.Path == modName {
			return repStatement, true
		}
	}
	return nil, false
}
