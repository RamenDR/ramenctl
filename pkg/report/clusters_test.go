// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"testing"

	ramenapi "github.com/ramendr/ramen/api/v1alpha1"
	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/ramen"
	"github.com/ramendr/ramenctl/pkg/report"
)

func TestReportClusterStatusEqual(t *testing.T) {
	c1 := testClusterStatus()

	t.Run("equal to self", func(t *testing.T) {
		c2 := c1
		checkClustersEqual(t, c1, c2)
	})
	t.Run("equal clusters", func(t *testing.T) {
		c2 := testClusterStatus()
		checkClustersEqual(t, c1, c2)
	})
}

func TestReportClusterStatusNotEqual(t *testing.T) {
	c1 := testClusterStatus()

	t.Run("not equal to nil", func(t *testing.T) {
		var c2 *report.ClustersStatus
		checkClustersNotEqual(t, c1, c2)
	})

	// Hub drClusters tests

	t.Run("hub drclusters nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters.Value = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drcluster conditions nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters.Value[0].Conditions = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drclusters state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drcluster name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters.Value[0].Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drcluster phase", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters.Value[0].Phase = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drcluster conditions", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters.Value[0].Conditions[0].State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})

	// Hub drPolicies tests

	t.Run("hub drpolicies nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy conditions nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value[0].Conditions = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicies state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value[0].Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy drclusters", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value[0].DRClusters[0] = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy scheduling interval", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value[0].SchedulingInterval = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy peer classes state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value[0].PeerClasses.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy peer classes storage class name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value[0].PeerClasses.Value[0].StorageClassName = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy peer classes grouping", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value[0].PeerClasses.Value[0].Grouping = false
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy conditions", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value[0].Conditions[0].State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})

	// Hub ramen configmap tests

	t.Run("hub ramen configmap s3storeprofiles value nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap namespace", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.Namespace = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap deleted", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.Deleted = report.ValidatedBool{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Resource does not exist",
			},
			Value: true,
		}
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen deployment ramen controller type state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Deployment.RamenControllerType.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen deployment ramen controller type", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Deployment.RamenControllerType.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3ProfileName = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles s3bucket state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3Bucket.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles s3bucket value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3Bucket.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles s3CompatibleEndpoint state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3CompatibleEndpoint.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles s3CompatibleEndpoint value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3CompatibleEndpoint.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles s3Region state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3Region.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles s3Region value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3Region.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles secret name state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.Name.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles secret name value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.Name.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles secret namespace state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.Namespace.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles secret namespace value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.Namespace.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles secret deleted", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.Deleted.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles secret accesskey state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.AWSAccessKeyID.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles secret accesskey value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.AWSAccessKeyID.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles secret secretkey state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.AWSSecretAccessKey.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles secret secretkey value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.AWSSecretAccessKey.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles caCertificate state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].CACertificate.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})

	t.Run("hub ramen configmap s3storeprofiles caCertificate value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].CACertificate.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})

	// Hub ramen deployment tests

	t.Run("hub ramen deployment conditions nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Deployment.Conditions = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen deployment name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Deployment.Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen deployment namespace", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Deployment.Namespace = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen deployment deleted", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Deployment.Deleted = report.ValidatedBool{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Resource does not exist",
			},
			Value: true,
		}
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen deployment replicas", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Deployment.Replicas = report.ValidatedInteger{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Expecting 1 replica",
			},
			Value: 0,
		}
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen deployment conditions", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Deployment.Conditions[0].State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})

	// Managed cluster tests

	t.Run("clusters nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles value nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen deployment conditions nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.Deployment.Conditions = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap namespace", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.Namespace = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen deployment ramen controller type state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.Deployment.RamenControllerType.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen deployment ramen controller type", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.Deployment.RamenControllerType.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles profile name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3ProfileName = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles s3bucket state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3Bucket.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles s3bucket value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3Bucket.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles s3CompatibleEndpoint state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3CompatibleEndpoint.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles s3CompatibleEndpoint value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3CompatibleEndpoint.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles s3Region state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3Region.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles s3Region value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3Region.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles secret name state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.Name.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles secret name value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.Name.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles secret namespace state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.Namespace.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles secret namespace value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.Namespace.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles secret deleted", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.Deleted.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles secret accesskey state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.AWSAccessKeyID.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles secret accesskey value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.AWSAccessKeyID.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles secret secretkey state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.AWSSecretAccessKey.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles secret secretkey value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.AWSSecretAccessKey.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles caCertificate state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].CACertificate.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles caCertificate value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].CACertificate.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen deployment name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.Deployment.Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen deployment namespace", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.Deployment.Namespace = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen deployment replicas", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.Deployment.Replicas = report.ValidatedInteger{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Expecting 1 replica",
			},
			Value: 2,
		}
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen deployment conditions", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.Deployment.Conditions[0].State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})

	// ClustersS3Status tests

	t.Run("s3 status profiles nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.S3.Profiles.Value = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("s3 status profiles state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.S3.Profiles.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("s3 status profiles description", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.S3.Profiles.Description = "S3 profiles not available"
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("s3 status profile name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.S3.Profiles.Value[0].Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("s3 status profile accessible state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.S3.Profiles.Value[0].Accessible.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("s3 status profile accessible value", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.S3.Profiles.Value[0].Accessible.Value = false
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("s3 status profile accessible description", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.S3.Profiles.Value[0].Accessible.Description = "connection refused"
		checkClustersNotEqual(t, c1, c2)
	})
}

func TestReportClusterStatusMarshaling(t *testing.T) {
	c1 := testClusterStatus()
	data, err := yaml.Marshal(c1)
	if err != nil {
		t.Fatal(err)
	}
	c2 := &report.ClustersStatus{}
	if err := yaml.Unmarshal(data, c2); err != nil {
		t.Fatal(err)
	}
	checkClustersEqual(t, c1, c2)
}

func testClusterStatus() *report.ClustersStatus {
	c := &report.ClustersStatus{
		Hub: report.ClustersStatusHub{
			DRClusters: report.ValidatedDRClustersList{
				Validated: report.Validated{
					State: report.OK,
				},
				Value: []report.DRClusterSummary{
					{
						Name:  "dr1",
						Phase: "Available",
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Fenced",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Clean",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Validated",
							},
						},
					},
					{
						Name:  "dr2",
						Phase: "Available",
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Fenced",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Clean",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Validated",
							},
						},
					},
				},
			},
			DRPolicies: report.ValidatedDRPoliciesList{
				Validated: report.Validated{
					State: report.OK,
				},
				Value: []report.DRPolicySummary{
					{
						Name:               "dr-policy-1m",
						DRClusters:         []string{"dr1", "dr2"},
						SchedulingInterval: "1m",
						PeerClasses: report.ValidatedPeerClassesList{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: []report.PeerClassesSummary{
								{
									StorageClassName: "rook-ceph-block",
									ReplicationID:    "rook-ceph-replication-1",
									Grouping:         true,
								},
								{
									StorageClassName: "rook-cephfs-fs1",
									Grouping:         true,
								},
							},
						},
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Validated",
							},
						},
					},
					{
						Name:               "dr-policy-5m",
						DRClusters:         []string{"dr1", "dr2"},
						SchedulingInterval: "5m",
						PeerClasses: report.ValidatedPeerClassesList{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: []report.PeerClassesSummary{
								{
									StorageClassName: "rook-ceph-block",
									ReplicationID:    "rook-ceph-replication-1",
									Grouping:         true,
								},
								{
									StorageClassName: "rook-cephfs-fs1",
									Grouping:         true,
								},
							},
						},
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Validated",
							},
						},
					},
				},
			},
			Ramen: report.RamenSummary{
				ConfigMap: report.ConfigMapSummary{
					Name:      ramen.HubOperatorConfigMapName,
					Namespace: "ramen-system",
					Deleted: report.ValidatedBool{
						Validated: report.Validated{
							State: report.OK,
						},
					},
					S3StoreProfiles: report.ValidatedS3StoreProfilesList{
						Validated: report.Validated{
							State: report.OK,
						},
						Value: []report.S3StoreProfilesSummary{
							{
								S3ProfileName: "s3-profile-dr1",
								S3Bucket: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "bucket",
								},
								S3CompatibleEndpoint: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "http://example-cluster:30000",
								},
								S3Region: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "us-west-1",
								},
								S3SecretRef: report.S3SecretSummary{
									Name: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "ramen-s3-secret-dr1",
									},
									Namespace: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "ramen-system",
									},
									Deleted: report.ValidatedBool{
										Validated: report.Validated{State: report.OK},
									},
									AWSAccessKeyID: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
										Value:     helpers.FakeAWSKeyIDFingerprint,
									},
									AWSSecretAccessKey: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
										Value:     helpers.FakeAWSKeyFingerprint,
									},
								},
								// CACertificate is optional, empty is OK if hub also has no cert.
								CACertificate: report.ValidatedFingerprint{
									Validated: report.Validated{State: report.OK},
								},
							},
							{
								S3ProfileName: "s3-profile-dr2",
								S3Bucket: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "bucket",
								},
								S3CompatibleEndpoint: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "http://example-cluster:30000",
								},
								S3Region: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "us-east-1",
								},
								S3SecretRef: report.S3SecretSummary{
									Name: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "ramen-s3-secret-dr2",
									},
									Namespace: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "ramen-system",
									},
									Deleted: report.ValidatedBool{
										Validated: report.Validated{State: report.OK},
									},
									AWSAccessKeyID: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
										Value:     helpers.FakeAWSKeyIDFingerprint,
									},
									AWSSecretAccessKey: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
										Value:     helpers.FakeAWSKeyFingerprint,
									},
								},
								CACertificate: report.ValidatedFingerprint{
									Validated: report.Validated{State: report.OK},
								},
							},
						},
					},
				},
				Deployment: report.DeploymentSummary{
					Name:      ramen.HubOperatorName,
					Namespace: "ramen-system",
					Deleted: report.ValidatedBool{
						Validated: report.Validated{
							State: report.OK,
						},
					},
					RamenControllerType: report.ValidatedString{
						Validated: report.Validated{
							State: report.OK,
						},
						Value: string(ramenapi.DRHubType),
					},
					Replicas: report.ValidatedInteger{
						Validated: report.Validated{
							State: report.OK,
						},
						Value: 1,
					},
					Conditions: []report.ValidatedCondition{
						{
							Validated: report.Validated{
								State: report.OK,
							},
							Type: "Available",
						},
						{
							Validated: report.Validated{
								State: report.OK,
							},
							Type: "Progressing",
						},
					},
				},
			},
		},
		Clusters: []report.ClustersStatusCluster{
			{
				Name: "dr1",
				Ramen: report.RamenSummary{
					ConfigMap: report.ConfigMapSummary{
						Name:      ramen.DrClusterOperatorConfigMapName,
						Namespace: "ramen-system",
						Deleted: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
						},
						S3StoreProfiles: report.ValidatedS3StoreProfilesList{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: []report.S3StoreProfilesSummary{
								{
									S3ProfileName: "s3-profile-dr1",
									S3Bucket: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "bucket",
									},
									S3CompatibleEndpoint: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "http://example-cluster:30000",
									},
									S3Region: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "us-west-1",
									},
									S3SecretRef: report.S3SecretSummary{
										Name: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     "ramen-s3-secret-dr1",
										},
										Namespace: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     "ramen-system",
										},
										Deleted: report.ValidatedBool{
											Validated: report.Validated{State: report.OK},
										},
										AWSAccessKeyID: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyIDFingerprint,
										},
										AWSSecretAccessKey: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyFingerprint,
										},
									},
									CACertificate: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
									},
								},
								{
									S3ProfileName: "s3-profile-dr2",
									S3Bucket: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "bucket",
									},
									S3CompatibleEndpoint: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "http://example-cluster:30000",
									},
									S3Region: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "us-east-1",
									},
									S3SecretRef: report.S3SecretSummary{
										Name: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     "ramen-s3-secret-dr2",
										},
										Namespace: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     "ramen-system",
										},
										Deleted: report.ValidatedBool{
											Validated: report.Validated{State: report.OK},
										},
										AWSAccessKeyID: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyIDFingerprint,
										},
										AWSSecretAccessKey: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyFingerprint,
										},
									},
									CACertificate: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
									},
								},
							},
						},
					},
					Deployment: report.DeploymentSummary{
						Name:      ramen.DRClusterOperatorName,
						Namespace: "ramen-system",
						Deleted: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
						},
						RamenControllerType: report.ValidatedString{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: string(ramenapi.DRClusterType),
						},
						Replicas: report.ValidatedInteger{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: 1,
						},
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Available",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Progressing",
							},
						},
					},
				},
			},
			{
				Name: "dr2",
				Ramen: report.RamenSummary{
					ConfigMap: report.ConfigMapSummary{
						Name:      ramen.DrClusterOperatorConfigMapName,
						Namespace: "ramen-system",
						Deleted: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
						},
						S3StoreProfiles: report.ValidatedS3StoreProfilesList{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: []report.S3StoreProfilesSummary{
								{
									S3ProfileName: "s3-profile-dr1",
									S3Bucket: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "bucket",
									},
									S3CompatibleEndpoint: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "http://example-cluster:30000",
									},
									S3Region: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "us-west-1",
									},
									S3SecretRef: report.S3SecretSummary{
										Name: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     "ramen-s3-secret-dr1",
										},
										Namespace: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     "ramen-system",
										},
										Deleted: report.ValidatedBool{
											Validated: report.Validated{State: report.OK},
										},
										AWSAccessKeyID: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyIDFingerprint,
										},
										AWSSecretAccessKey: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyFingerprint,
										},
									},
									CACertificate: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
									},
								},
								{
									S3ProfileName: "s3-profile-dr2",
									S3Bucket: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "bucket",
									},
									S3CompatibleEndpoint: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "http://example-cluster:30000",
									},
									S3Region: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "us-east-1",
									},
									S3SecretRef: report.S3SecretSummary{
										Name: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     "ramen-s3-secret-dr2",
										},
										Namespace: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     "ramen-system",
										},
										Deleted: report.ValidatedBool{
											Validated: report.Validated{State: report.OK},
										},
										AWSAccessKeyID: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyIDFingerprint,
										},
										AWSSecretAccessKey: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyFingerprint,
										},
									},
									CACertificate: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
									},
								},
							},
						},
					},
					Deployment: report.DeploymentSummary{
						Name:      ramen.DRClusterOperatorName,
						Namespace: "ramen-system",
						Deleted: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
						},
						RamenControllerType: report.ValidatedString{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: string(ramenapi.DRClusterType),
						},
						Replicas: report.ValidatedInteger{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: 1,
						},
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Available",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Progressing",
							},
						},
					},
				},
			},
		},
		S3: report.ClustersS3Status{
			Profiles: report.ValidatedClustersS3ProfileStatusList{
				Validated: report.Validated{
					State: report.OK,
				},
				Value: []report.ClustersS3ProfileStatus{
					{
						Name: "minio-on-dr1",
						Accessible: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: true,
						},
					},
					{
						Name: "minio-on-dr2",
						Accessible: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: true,
						},
					},
				},
			},
		},
	}
	return c
}

func checkClustersEqual(t *testing.T, a, b *report.ClustersStatus) {
	if !a.Equal(b) {
		diff := helpers.UnifiedDiff(t, a, b)
		t.Fatalf("clusters statuses are not equal\n%s", diff)
	}
}

func checkClustersNotEqual(t *testing.T, a, b *report.ClustersStatus) {
	if a.Equal(b) {
		t.Fatalf("clusters statuses are equal\n%s", helpers.MarshalYAML(t, a))
	}
}
