/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/redhat-openshift-ecosystem/preflight-trigger/cmd"
	"k8s.io/test-infra/prow/interrupts"
	"os"
)

func main() {
	go func() {
		interrupts.WaitForGracefulShutdown()
		os.Exit(128)
	}()
	cmd.Execute()
}
