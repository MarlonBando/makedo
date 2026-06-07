package engine

import (
	"fmt"
	"os"
	"path/filepath"
)

type RunContext struct {
	MkTmpDir  string
	MkEnvFile string
	Registry  *Registry
}

func NewRunContext() (*RunContext, error) {
	tmpDir, err := os.MkdirTemp("", "makedo-env-")
	if err != nil {
		return nil, err
	}

	mkEnvFile := filepath.Join(tmpDir, ".mkenv")
	envContent := fmt.Sprintf("MAKEDO_ENV='%s'\n", mkEnvFile)
	err = os.WriteFile(mkEnvFile, []byte(envContent), 0600)
	if err != nil {
		return nil, err
	}

	return &RunContext{
		MkTmpDir:  tmpDir,
		MkEnvFile: mkEnvFile,
		Registry:  NewRegistry(),
	}, nil
}

func (ctx *RunContext) Cleanup() {
	os.RemoveAll(ctx.MkTmpDir)
	if ctx.Registry != nil {
		ctx.Registry.KillAll()
	}
}
