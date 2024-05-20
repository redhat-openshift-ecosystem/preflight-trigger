/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	"github.com/redhat-openshift-ecosystem/preflight-trigger/cmd"

	"sigs.k8s.io/prow/pkg/interrupts"
)

func main() {
	go func() {
		interrupts.WaitForGracefulShutdown()
		os.Exit(128)
	}()
	cmd.Execute()
}
