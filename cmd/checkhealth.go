/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"github.com/gobuffalo/envy"
	osconfigv1 "github.com/openshift/api/config/v1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"log"
	ctrlrt "sigs.k8s.io/controller-runtime"
	ctrlrtc "sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"time"
)

// checkhealthCmd represents the checkhealth command
var checkhealthCmd = &cobra.Command{
	Use:   "checkhealth",
	Short: "Verify cluster under test is ready.",
	Long: `Check that the /readyz endpoint is healthy. Also check that
the cluster is not in a degraded state. We do this by checking the
Degraded, Available, and Progressing status' of clusteroperator
resources. We also check the clusterversion resource to ensure its
Progrssing status is not true.'`,
	PreRun: checkhealthPreRun,
	Run:    checkhealthRun,
}

var kclient *kubernetes.Clientset
var oclient ctrlrtc.Client

func init() {
	rootCmd.AddCommand(checkhealthCmd)
}

func checkhealthPreRun(cmd *cobra.Command, args []string) {
	config, err := ctrlrt.GetConfig()
	if err != nil {
		log.Fatalf("Error getting config: %s", err)
	}

	kclient = kubernetes.NewForConfigOrDie(config)

	scheme := runtime.NewScheme()
	if err = osconfigv1.AddToScheme(scheme); err != nil {
		log.Fatalf("Error adding OpenShift config API to scheme: %s", err)
	}
	oclient, err = ctrlrtc.New(config, ctrlrtc.Options{Scheme: scheme})
	if err != nil {
		log.Fatalf("Error getting OpenShift config client: %s", err)
	}
}

func checkhealthRun(cmd *cobra.Command, args []string) {
	checkreadyz()

	oht := envy.Get("OPERATOR_HEALTH_TIMEOUT", "10")
	timeout, err := strconv.Atoi(oht)
	if err != nil {
		log.Fatalf("Error converting OPERATOR_HEALTH_TIMEOUT envvar to int: %s", err)
	}

	for i := 0; i < timeout; i++ {
		var cvok bool
		cook := checkclusteroperators()
		if cook {
			cvok = checkclusterversion()
		}
		if !cvok {
			time.Sleep(time.Minute)
		} else {
			log.Println("cluster operators are ok, cluster version is ok")
			break
		}
	}
}

func checkreadyz() {
	var readyz []byte
	oht := envy.Get("OPERATOR_HEALTH_TIMEOUT", "10")
	timeout, err := strconv.Atoi(oht)
	if err != nil {
		log.Fatalf("Error converting OPERATOR_HEALTH_TIMEOUT envvar to int: %s", err)
	}

	for i := 0; i < timeout; i++ {
		readyz, err = kclient.CoreV1().RESTClient().Get().AbsPath("/readyz").DoRaw(context.TODO())
		if string(readyz) == "ok" {
			log.Println("/readyz is ok")
			return
		} else {
			log.Printf("/readyz check failed, will check again in 1 minute")
			time.Sleep(time.Minute)
		}
	}

	log.Fatalf("/readyz took longer than %d minutes to succeed", timeout)
}

func checkclusterversion() bool {
	cvl := osconfigv1.ClusterVersionList{}
	err := oclient.List(context.Background(), &cvl)
	if err != nil {
		log.Fatalf("Unable to get ClusterVersionList: %v", err)
	}

	for _, cv := range cvl.Items {
		for _, cond := range cv.Status.Conditions {
			if cond.Type == "Progressing" && cond.Status == "True" {
				log.Println("ClusterVersion not ready")
				return false
			}
		}
	}
	log.Println("ClusterVersion is ready")
	return true
}

func checkclusteroperators() bool {
	oht := envy.Get("OPERATOR_HEALTH_TIMEOUT", "10")
	_, err := strconv.Atoi(oht)
	if err != nil {
		log.Fatalf("Error converting OPERATOR_HEALTH_TIMEOUT envvar to int: %s", err)
	}

	col := osconfigv1.ClusterOperatorList{}
	err = oclient.List(context.Background(), &col)
	if err != nil {
		log.Fatalf("Unable to list cluster operators: %v", err)
	}

	for _, co := range col.Items {
		for _, conds := range co.Status.Conditions {
			switch {
			case conds.Type == "Degraded" && conds.Status == "True":
				log.Printf("ClusterOperator %s is degraded", co.Name)
				return false
			case conds.Type == "Available" && conds.Status == "False":
				log.Printf("ClusterOperator %s is not available", co.Name)
				return false
			case conds.Type == "Progressing" && conds.Status == "True":
				log.Printf("ClusterOperator %s is progressing", co.Name)
				return false
			}
		}
	}
	log.Println("All cluster operators are ready")
	return true
}
