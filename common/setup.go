package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/osde2e/pkg/config"
	"github.com/openshift/osde2e/pkg/events"
	"github.com/openshift/osde2e/pkg/metadata"
	"github.com/openshift/osde2e/pkg/osd"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// Setup cluster before testing begins.
var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	defer ginkgo.GinkgoRecover()
	cfg := config.Cfg

	err := setupCluster(cfg)
	events.HandleErrorWithEvents(err, events.InstallSuccessful, events.InstallFailed).ShouldNot(HaveOccurred(), "failed to setup cluster for testing")

	if len(cfg.Addons.IDs) > 0 {
		err = installAddons(cfg)
		events.HandleErrorWithEvents(err, events.InstallAddonsSuccessful, events.InstallAddonsFailed).ShouldNot(HaveOccurred(), "failed while installing addons")
	}

	if len(cfg.Kubeconfig.Contents) == 0 {
		// Give the cluster some breathing room.
		log.Println("OSD cluster installed. Sleeping for 600s.")
		time.Sleep(600 * time.Second)
	}

	return []byte{}
}, func(data []byte) {
	// only needs to run once
})

// Collect logs after each test
var _ = ginkgo.AfterSuite(func() {
	log.Printf("Getting logs for cluster '%s'...", config.Cfg.Cluster.ID)
	getLogs()
})
var _ = ginkgo.JustAfterEach(getLogs)

func getLogs() {
	defer ginkgo.GinkgoRecover()
	cfg := config.Cfg

	if OSD == nil {
		log.Println("OSD was not configured. Skipping log collection...")
	} else if cfg.Cluster.ID == "" {
		log.Println("CLUSTER_ID is not set, likely due to a setup failure. Skipping log collection...")
	} else {
		logs, err := OSD.FullLogs(cfg.Cluster.ID)
		Expect(err).NotTo(HaveOccurred(), "failed to collect cluster logs")
		writeLogs(cfg, logs)
	}
}

// setupCluster brings up a cluster, waits for it to be ready, then returns it's name.
func setupCluster(cfg *config.Config) (err error) {
	// if TEST_KUBECONFIG has been set, skip configuring OCM
	if len(cfg.Kubeconfig.Contents) > 0 || len(cfg.Kubeconfig.Path) > 0 {
		return useKubeconfig(cfg)
	}

	// create a new cluster if no ID is specified
	if cfg.Cluster.ID == "" {
		if cfg.Cluster.Name == "" {
			cfg.Cluster.Name = clusterName(cfg)
		}

		if cfg.Cluster.ID, err = OSD.LaunchCluster(cfg); err != nil {
			return fmt.Errorf("could not launch cluster: %v", err)
		}
	} else {
		log.Printf("CLUSTER_ID of '%s' was provided, skipping cluster creation and using it instead", cfg.Cluster.ID)

		if cfg.Cluster.Name == "" {
			cluster, err := OSD.GetCluster(cfg.Cluster.ID)
			if err != nil {
				return fmt.Errorf("could not retrieve cluster information from OCM: %v", err)
			}

			if cluster.Name() == "" {
				return fmt.Errorf("cluster name from OCM is empty, and this shouldn't be possible")
			}

			cfg.Cluster.Name = cluster.Name()
			log.Printf("CLUSTER_NAME not provided, retrieved %s from OCM.", cfg.Cluster.Name)
		}
	}

	metadata.Instance.ClusterName = cfg.Cluster.Name
	metadata.Instance.ClusterID = cfg.Cluster.ID

	if err = OSD.WaitForClusterReady(cfg); err != nil {
		return fmt.Errorf("failed waiting for cluster ready: %v", err)
	}

	if cfg.Kubeconfig.Contents, err = OSD.ClusterKubeconfig(cfg.Cluster.ID); err != nil {
		return fmt.Errorf("could not get kubeconfig for cluster: %v", err)
	}

	return nil
}

// installAddons installs addons onto the cluster
func installAddons(cfg *config.Config) (err error) {
	num, err := OSD.InstallAddons(cfg)
	if err != nil {
		return fmt.Errorf("could not install addons: %s", err.Error())
	}
	if num > 0 {
		if err = OSD.WaitForClusterReady(cfg); err != nil {
			return fmt.Errorf("failed waiting for cluster ready: %v", err)
		}
	}

	return nil
}

// useKubeconfig reads the path provided for a TEST_KUBECONFIG and uses it for testing.
func useKubeconfig(cfg *config.Config) (err error) {
	_, err = clientcmd.RESTConfigFromKubeConfig(cfg.Kubeconfig.Contents)
	if err != nil {
		log.Println("Not an existing Kubeconfig, attempting to read file instead...")
	} else {
		log.Println("Existing valid kubeconfig!")
		return nil
	}

	cfg.Kubeconfig.Contents, err = ioutil.ReadFile(cfg.Kubeconfig.Path)
	if err != nil {
		return fmt.Errorf("failed reading '%s' which has been set as the TEST_KUBECONFIG: %v", cfg.Kubeconfig.Path, err)
	}
	log.Printf("Using a set TEST_KUBECONFIG of '%s' for Origin API calls.", cfg.Kubeconfig.Path)
	return nil
}

// cluster name format must be short enough to support all versions
func clusterName(cfg *config.Config) string {
	vers := strings.TrimPrefix(cfg.Cluster.Version, osd.VersionPrefix)
	safeVersion := strings.Replace(vers, ".", "-", -1)
	return "ci-cluster-" + safeVersion + "-" + cfg.Suffix
}

func randomStr(length int) (str string) {
	chars := "0123456789abcdefghijklmnopqrstuvwxyz"
	for i := 0; i < length; i++ {
		c := string(chars[rand.Intn(len(chars))])
		str += c
	}
	return
}

func writeLogs(cfg *config.Config, m map[string][]byte) {
	for k, v := range m {
		name := k + "-log.txt"
		filePath := filepath.Join(cfg.ReportDir, name)
		err := ioutil.WriteFile(filePath, v, os.ModePerm)
		Expect(err).NotTo(HaveOccurred(), "failed to write log '%s'", filePath)
	}
}
