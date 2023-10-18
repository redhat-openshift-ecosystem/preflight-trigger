/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	"github.com/redhat-openshift-ecosystem/preflight-trigger/cmd"

	"k8s.io/test-infra/prow/interrupts"
)

func main() {
	go func() {
		interrupts.WaitForGracefulShutdown()
		os.Exit(128)
	}()
	cmd.Execute()
}
