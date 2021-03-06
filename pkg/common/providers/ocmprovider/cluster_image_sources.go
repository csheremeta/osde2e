package ocmprovider

import (
	"log"
	"math/rand"
)

var clusterImageSources = map[string]string{"quay-primary": `imageContentSources:
- mirrors:
  - quay.io/openshift-release-dev/ocp-release
  - 950916221866.dkr.ecr.us-east-1.amazonaws.com/ocp-release
  source: quay.io/openshift-release-dev/ocp-release
- mirrors:
  - quay.io/openshift-release-dev/ocp-v4.0-art-dev
  - 950916221866.dkr.ecr.us-east-1.amazonaws.com/ocp-art-dev
  source: quay.io/openshift-release-dev/ocp-v4.0-art-dev
- mirrors:
  - quay.io/app-sre/managed-upgrade-operator
  - 950916221866.dkr.ecr.us-east-1.amazonaws.com/managed-upgrade-operator
  source: quay.io/app-sre/managed-upgrade-operator
- mirrors:
  - quay.io/app-sre/managed-upgrade-operator-registry
  - 950916221866.dkr.ecr.us-east-1.amazonaws.com/managed-upgrade-operator-registry
  source: quay.io/app-sre/managed-upgrade-operator-registry`,
	"ecr-primary": `imageContentSources:
- mirrors:
  - 950916221866.dkr.ecr.us-east-1.amazonaws.com/ocp-release
  - quay.io/openshift-release-dev/ocp-release
  source: quay.io/openshift-release-dev/ocp-release
- mirrors:
  - 950916221866.dkr.ecr.us-east-1.amazonaws.com/ocp-art-dev
  - quay.io/openshift-release-dev/ocp-v4.0-art-dev
  source: quay.io/openshift-release-dev/ocp-v4.0-art-dev
- mirrors:
  - 950916221866.dkr.ecr.us-east-1.amazonaws.com/managed-upgrade-operator
  - quay.io/app-sre/managed-upgrade-operator
  source: quay.io/app-sre/managed-upgrade-operator
- mirrors:
  - 950916221866.dkr.ecr.us-east-1.amazonaws.com/managed-upgrade-operator-registry
  - quay.io/app-sre/managed-upgrade-operator-registry
  source: quay.io/app-sre/managed-upgrade-operator-registry`,
	"ecr-only": `imageContentSources:
- mirrors:
  - 950916221866.dkr.ecr.us-east-1.amazonaws.com/ocp-release
  source: quay.io/openshift-release-dev/ocp-release
- mirrors:
  - 950916221866.dkr.ecr.us-east-1.amazonaws.com/ocp-art-dev
  source: quay.io/openshift-release-dev/ocp-v4.0-art-dev
- mirrors:
  - 950916221866.dkr.ecr.us-east-1.amazonaws.com/managed-upgrade-operator
  source: quay.io/app-sre/managed-upgrade-operator
- mirrors:
  - 950916221866.dkr.ecr.us-east-1.amazonaws.com/managed-upgrade-operator-registry
  source: quay.io/app-sre/managed-upgrade-operator-registry`,
	"quay-only": `imageContentSources:
- mirrors:
  - quay.io/openshift-release-dev/ocp-release
  source: quay.io/openshift-release-dev/ocp-release
- mirrors:
  - quay.io/openshift-release-dev/ocp-v4.0-art-dev
  source: quay.io/openshift-release-dev/ocp-v4.0-art-dev
- mirrors:
  - quay.io/app-sre/managed-upgrade-operator
  source: quay.io/app-sre/managed-upgrade-operator
- mirrors:
  - quay.io/app-sre/managed-upgrade-operator-registry
  source: quay.io/app-sre/managed-upgrade-operator-registry`}

func (o *OCMProvider) ChooseImageSource(choice string) (source string) {
	var ok bool
	if choice == "random" || choice == "" {
		var sources []string
		for key := range clusterImageSources {
			sources = append(sources, key)
		}
		choice = sources[rand.Intn(len(sources))]
	}
	if source, ok = clusterImageSources[choice]; !ok {
		log.Printf("Image source not found: %s", choice)
		return ""
	}

	log.Printf("Choice: %s", choice)
	log.Println(source)
	return source
}
