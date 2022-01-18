//go:build e2e
// +build e2e

package e2e

import (
	"context"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"testing"
	//cpomanifests "github.com/openshift/hypershift/control-plane-operator/controllers/hostedcontrolplane/manifests"
	"github.com/openshift/hypershift/hypershift-operator/controllers/manifests"

	configv1 "github.com/openshift/api/config/v1"
	hyperv1 "github.com/openshift/hypershift/api/v1alpha1"
	"github.com/openshift/hypershift/support/globalconfig"
	e2eutil "github.com/openshift/hypershift/test/e2e/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// TestGusetClusterKubeAdmin executes a suite of guest cluster tests which ensure that guest cluster admin user is
// behaving as expected on the guest cluster.
func TestGusetClusterKubeAdmin(t *testing.T) {
	t.Parallel()

	g := NewWithT(t)

	ctx, cancel := context.WithCancel(testContext)
	defer cancel()

	client := e2eutil.GetClientOrDie()

	// Create a cluster with oauth nil
	clusterOpts := globalOpts.DefaultClusterOptions()

	//globalConfig.OAuth = nil
	hostedCluster := e2eutil.CreateAWSCluster(t, ctx, client, clusterOpts, globalOpts.ArtifactDir)
	nodepool := &hyperv1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: hostedCluster.Namespace,
			Name:      hostedCluster.Name,
		},
	}
	err := client.Get(testContext, crclient.ObjectKeyFromObject(nodepool), nodepool)
	g.Expect(err).NotTo(HaveOccurred(), "failed to get nodepool")

	// Sanity check the cluster by waiting for the nodes to report ready
	t.Logf("Waiting for guest client to become available")
	guestClient := e2eutil.WaitForGuestClient(t, testContext, client, hostedCluster)
	e2eutil.WaitForNReadyNodes(t, testContext, guestClient, *nodepool.Spec.NodeCount)
	kubeadminPasswordSecret := manifests.KubeadminPasswordSecret(hostedCluster.Namespace, hostedCluster.Name)
	err = client.Get(testContext, crclient.ObjectKeyFromObject(kubeadminPasswordSecret), kubeadminPasswordSecret)
	g.Expect(err).NotTo(HaveOccurred(), "failed to get kubeadmin password secret")
	namespace := manifests.HostedControlPlaneNamespace(hostedCluster.Namespace, hostedCluster.Name).Name
	cp := &hyperv1.HostedControlPlane{}
	err = client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: hostedCluster.Name}, cp)
	if err != nil {
		t.Errorf("Failed to get hostedcontrolplane: %v", err)
	}
	globalConfig, err := globalconfig.ParseGlobalConfig(ctx, cp.Spec.Configuration)
	if err != nil {
		t.Errorf("failed to parse global config: %w", err)
	}

	t.Logf("Updating cluster  oAuth: %s", globalConfig.OAuth)
	err = client.Get(testContext, crclient.ObjectKeyFromObject(hostedCluster), hostedCluster)
	g.Expect(err).NotTo(HaveOccurred(), "failed to get hostedcluster")

	globalConfig.OAuth = &cp.Spec.Configuration.OAuth{
		// ApiVersion: "oauth.openshift.io/v1",
		// Kind:       "OAuth",
		metav1.ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
		Spec: globalconfig.OAuthSpec{
			IdentityProviders: []globalconfig.IdentityProvider{
				{
					github: &globalconfig.GithubIdentityProvider{
						Name:         "github",
						Type:         "github",
						ClientID:     "123456789",
						ClientSecret: kubeadminPasswordSecret.Data["kubeadmin-password"],
						Organizations: []string{
							"openshift",
						},
						MappingMethod: "claim",
					},
				},
			},
			secretRefs: []globalconfig.SecretRef{
				{
					Name: "github-client-secret",
					Key:  "github-client-secret",
				},
			},
		},
	}
	//cp.Spec.Configuration = globalconfig.UpdateGlobalConfig(cp.Spec.Configuration, globalConfig.OAuth)
	err = client.Update(testContext, hostedCluster)
	g.Expect(err).NotTo(HaveOccurred(), "failed update hostedcluster image")
	err = client.Get(testContext, crclient.ObjectKeyFromObject(hostedCluster), hostedCluster)
	g.Expect(err).NotTo(HaveOccurred(), "failed to get hostedcluster")
	kubeadminPasswordSecret = manifests.KubeadminPasswordSecret(hostedCluster.Namespace, hostedCluster.Name)
	err = client.Get(testContext, crclient.ObjectKeyFromObject(kubeadminPasswordSecret), kubeadminPasswordSecret)
	g.Expect(err).NotTo(BeEmpty(), "couldn't find any kubeadmin password secret")
}
