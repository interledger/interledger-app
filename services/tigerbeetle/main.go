package main

import (
	"errors"
	"path/filepath"
	"runtime"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/kustomize"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployTigerBeetle(ctx *pulumi.Context) error {
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		return errors.New("Could not get directory path for kubernetes module.")
	}

	_, err := kustomize.NewDirectory(ctx, "tigerbeetle",
		kustomize.DirectoryArgs{
			Directory: pulumi.String(filepath.Join(filepath.Dir(moduleDir), "deploy/base")),
		},
	)
	if err != nil {
		return err
	}

	return nil
}
