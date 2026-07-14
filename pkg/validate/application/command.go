// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	stdtime "time"

	ramenapi "github.com/ramendr/ramen/api/v1alpha1"
	e2etypes "github.com/ramendr/ramen/e2e/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	basecmd "github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/core"
	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/logging"
	"github.com/ramendr/ramenctl/pkg/ramen"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/s3"
	"github.com/ramendr/ramenctl/pkg/time"
	validatecmd "github.com/ramendr/ramenctl/pkg/validate/command"
	"github.com/ramendr/ramenctl/pkg/validate/summary"
	"github.com/ramendr/ramenctl/pkg/validation"
)

// CommandName is the name of the validate-application command.
const CommandName = "validate-application"

type Command struct {
	*validatecmd.Command
	opts   basecmd.ApplicationOptions
	Report *Report
}

func NewCommand(
	cmd *basecmd.Command,
	cfg *config.Config,
	backend validation.Validation,
	opts basecmd.ApplicationOptions,
) *Command {
	r := NewReport(cfg)
	return &Command{
		Command: validatecmd.New(cmd, cfg, backend, r.Report),
		opts:    opts,
		Report:  r,
	}
}

func (c *Command) passed() {
	c.WriteReport(c.Report)
	console.Completed("Validation completed (%s)", summary.String(c.Report.Summary))
}

func (c *Command) failed() error {
	c.WriteReport(c.Report)
	return errors.New(c.Report.Error())
}

func (c *Command) Run() error {
	c.Report.Application.Name = c.opts.DRPCName
	c.Report.Application.Namespace = c.opts.DRPCNamespace
	if !c.ValidateConfig() {
		return c.failed()
	}
	if !c.validateApplication() {
		return c.failed()
	}
	c.passed()
	return nil
}

func (c *Command) validateApplication() bool {
	console.Step("Validate application")
	c.StartStep("validate application")

	namespaces, ok := c.inspectApplication()
	if !ok {
		return c.FinishStep()
	}

	c.Report.Namespaces = namespaces

	options := gathering.Options{
		Namespaces: namespaces,
		OutputDir:  c.DataDir(),
	}
	if !c.GatherNamespaces(options) {
		return c.FinishStep()
	}

	if !c.gatherS3Data() {
		return c.FinishStep()
	}

	if !c.validateGatheredData() {
		return c.FinishStep()
	}

	c.FinishStep()
	return true
}

func (c *Command) inspectApplication() ([]string, bool) {
	start := time.Now()
	step := &report.Step{Name: "inspect application"}
	c.Logger().Infof("Step %q started", step.Name)

	namespaces, err := c.namespacesToGather()
	if err != nil {
		step.Duration = time.Since(start).Seconds()
		if errors.Is(err, context.Canceled) {
			console.Error("Canceled %s", step.Name)
			step.Status = report.Canceled
			step.Err = fmt.Sprintf("Canceled %s", step.Name)
		} else {
			console.Error("Failed to %s", step.Name)
			step.Status = report.Failed
			step.Err = fmt.Sprintf(
				"Failed to inspect application %q in namespace %q",
				c.opts.DRPCName, c.opts.DRPCNamespace,
			)
		}
		c.Logger().Errorf("Step %q %s: %s", c.Current.Name, step.Status, err)
		c.Current.AddStep(step)

		return nil, false
	}

	step.Duration = time.Since(start).Seconds()
	step.Status = report.Passed
	c.Current.AddStep(step)

	console.Pass("Inspected application")
	c.Logger().Infof("Step %q passed", step.Name)

	return namespaces, true
}

// gatherS3Data inspects application S3 profiles and gathers data. It returns false only if the
// user cancelled, otherwise true if there were errors during inspection, as those will be reported
// in the validation results.
func (c *Command) gatherS3Data() bool {
	profiles, prefix, err := c.inspectS3Profiles()
	if err != nil {
		return !errors.Is(err, context.Canceled)
	}
	return c.gatherS3Profiles(profiles, prefix)
}

func (c *Command) inspectS3Profiles() ([]*s3.Profile, string, error) {
	start := time.Now()
	step := &report.Step{Name: "inspect S3 profiles"}

	c.Logger().Infof("Step %q started", step.Name)

	profiles, prefix, err := c.s3Info()
	if err != nil {
		step.Duration = time.Since(start).Seconds()
		if errors.Is(err, context.Canceled) {
			step.Status = report.Canceled
			step.Err = fmt.Sprintf("Canceled %s", step.Name)
			console.Error("Canceled %s", step.Name)
		} else {
			step.Status = report.Failed
			step.Err = "Failed to read S3 profiles from hub"
			console.Error("Failed to %s", step.Name)
		}
		c.Logger().Errorf("Step %q %s: %s", c.Current.Name, step.Status, err)
		c.Current.AddStep(step)
		return nil, "", err
	}

	step.Duration = time.Since(start).Seconds()
	step.Status = report.Passed
	c.Current.AddStep(step)

	console.Pass("Inspected S3 profiles")
	c.Logger().Infof("Step %q passed", step.Name)

	return profiles, prefix, nil
}

// gatherS3Profiles gathers S3 data from the given profiles using the specified prefix. Returns
// false only if the user cancelled, otherwise true even if there were errors, as those will be
// reported during validation.
func (c *Command) gatherS3Profiles(profiles []*s3.Profile, prefix string) bool {
	start := time.Now()
	outputDir := c.DataDir()

	c.Logger().Infof("Gathering application S3 data from profiles %q with prefix %q",
		logging.ProfileNames(profiles), prefix)

	var failedProfiles []string
	for r := range c.Backend.GatherS3(c, profiles, []string{prefix}, outputDir) {
		// Store the s3 gather result for validation.
		c.S3Results = append(c.S3Results, r)

		step := &report.Step{
			Name:     fmt.Sprintf("gather S3 profile %q", r.ProfileName),
			Duration: r.Duration,
		}
		if r.Err != nil {
			if errors.Is(r.Err, context.Canceled) {
				msg := fmt.Sprintf("Canceled gather S3 profile %q", r.ProfileName)
				console.Error(msg)
				c.Logger().Errorf("%s: %s", msg, r.Err)
				step.Status = report.Canceled
				step.Err = msg
			} else {
				msg := fmt.Sprintf("Failed to gather S3 profile %q", r.ProfileName)
				console.Error(msg)
				c.Logger().Errorf("%s: %s", msg, r.Err)
				step.Status = report.Failed
				step.Err = fmt.Sprintf("Failed to gather S3 profile %q", r.ProfileName)
				failedProfiles = append(failedProfiles, r.ProfileName)
			}
		} else {
			step.Status = report.Passed
			console.Pass("Gathered S3 profile %q", r.ProfileName)
		}
		c.Current.AddStep(step)
	}

	c.Logger().Infof("Gathered application S3 data in %.2f seconds", time.Since(start).Seconds())

	switch c.Current.Status {
	case report.Canceled:
		c.Current.Err = "Canceled gather S3 profiles"
		return false
	case report.Failed:
		c.Current.Err = fmt.Sprintf(
			"Failed to gather S3 profiles %s",
			strings.Join(failedProfiles, ", "),
		)
		return true
	default:
		return true
	}
}

func (c *Command) namespacesToGather() ([]string, error) {
	set := map[string]struct{}{
		// Gather ramen namespaces to get ramen hub and dr-cluster logs and related resources.
		c.Config().Namespaces.RamenHubNamespace:       {},
		c.Config().Namespaces.RamenDRClusterNamespace: {},
	}

	appNamespaces, err := c.Backend.ApplicationNamespaces(c, c.opts.DRPCName, c.opts.DRPCNamespace)
	if err != nil {
		return nil, err
	}

	for _, ns := range appNamespaces {
		set[ns] = struct{}{}
	}

	return slices.Sorted(maps.Keys(set)), nil
}

// s3Info reads S3 profiles and application prefix from gathered hub data.
func (c *Command) s3Info() ([]*s3.Profile, string, error) {
	// Read S3 profiles from the ramen hub configmap, the source of truth
	// synced to managed clusters.
	hub := c.Env().Hub
	reader := c.OutputReader(hub.Name)
	configMapName := ramen.HubOperatorConfigMapName
	configMapNamespace := c.Config().Namespaces.RamenHubNamespace

	storeProfiles, err := ramen.ClusterProfiles(reader, configMapName, configMapNamespace)
	if err != nil {
		return nil, "", err
	}

	// Get S3 secrets from live hub cluster since gathered data may contain
	// sanitized secrets. On cancellation, return immediately. On other failures,
	// empty credentials will cause S3 operations to fail during gatherS3.
	var profiles []*s3.Profile
	for _, sp := range storeProfiles {
		secret, err := c.Backend.GetSecret(c, hub, sp.S3SecretRef.Name, sp.S3SecretRef.Namespace)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil, "", err
			}
			c.Logger().Warnf("Failed to get S3 secret \"%s/%s\" from cluster %q: %s",
				sp.S3SecretRef.Namespace, sp.S3SecretRef.Name, hub.Name, err)
		}
		profiles = append(profiles, ramen.S3ProfileFromStore(sp, secret))
	}

	prefix, err := ramen.ApplicationS3Prefix(reader, c.opts.DRPCName, c.opts.DRPCNamespace)
	if err != nil {
		return nil, "", err
	}

	return profiles, prefix, nil
}

func (c *Command) validateGatheredData() bool {
	log := c.Logger()

	start := time.Now()
	step := &report.Step{Name: "validate data"}
	defer func() {
		step.Duration = time.Since(start).Seconds()
		c.Current.AddStep(step)
	}()

	s := &c.Report.ApplicationStatus

	drpc, err := c.validateHub(&s.Hub)
	if err != nil {
		step.Status = report.Failed
		step.Err = "Failed to validate hub"
		msg := "Failed to validate hub"
		console.Error(msg)
		log.Errorf("%s: %s", msg, err)
		return false
	}

	if err := c.validatePrimaryCluster(&s.PrimaryCluster, drpc); err != nil {
		step.Status = report.Failed
		step.Err = fmt.Sprintf("Failed to validate primary cluster %q", s.PrimaryCluster.Name)
		msg := "Failed to validate primary cluster"
		console.Error(msg)
		log.Errorf("%s: %s", msg, err)
		return false
	}

	if err := c.validateSecondaryCluster(&s.SecondaryCluster, drpc); err != nil {
		step.Status = report.Failed
		step.Err = fmt.Sprintf("Failed to validate secondary cluster %q", s.SecondaryCluster.Name)
		msg := "Failed to validate secondary cluster"
		console.Error(msg)
		log.Errorf("%s: %s", msg, err)
		return false
	}

	c.validateS3Status(&s.S3)

	if summary.HasIssues(c.Report.Summary) {
		step.Status = report.Failed
		step.Err = fmt.Sprintf("Validation failed (%s)", summary.String(c.Report.Summary))
		msg := "Issues found during validation"
		console.Error(msg)
		log.Errorf("%s: %s", msg, summary.String(c.Report.Summary))
		return false
	}

	step.Status = report.Passed
	console.Pass("Application validated")
	return true
}

func (c *Command) validateHub(
	s *report.ApplicationStatusHub,
) (*ramenapi.DRPlacementControl, error) {
	log := c.Logger()
	reader := c.OutputReader(c.Env().Hub.Name)
	drpc, err := ramen.ReadDRPC(reader, c.opts.DRPCName, c.opts.DRPCNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to read drpc: %w", err)
	}
	log.Debugf("Read drpc \"%s/%s\"", drpc.Namespace, drpc.Name)
	c.validateDRPC(&s.DRPC, drpc)
	return drpc, nil
}

func (c *Command) validatePrimaryCluster(
	s *report.ApplicationStatusCluster,
	drpc *ramenapi.DRPlacementControl,
) error {
	cluster, err := ramen.PrimaryCluster(c, drpc)
	if err != nil {
		return fmt.Errorf("failed to find primary cluster: %w", err)
	}
	s.Name = cluster.Name
	return c.validateVRG(&s.VRG, cluster, drpc, ramenapi.PrimaryState)
}

func (c *Command) validateSecondaryCluster(
	s *report.ApplicationStatusCluster,
	drpc *ramenapi.DRPlacementControl,
) error {
	cluster, err := ramen.SecondaryCluster(c, drpc)
	if err != nil {
		return fmt.Errorf("failed to find secondary cluster: %w", err)
	}
	s.Name = cluster.Name
	return c.validateVRG(&s.VRG, cluster, drpc, ramenapi.SecondaryState)
}

func (c *Command) validateS3Status(s *report.ApplicationS3Status) {
	c.validatedS3ProfileStatus(&s.Profiles)
}

func (c *Command) validatedS3ProfileStatus(
	s *report.ValidatedApplicationS3ProfileStatusList,
) {
	if len(c.S3Results) > 0 {
		// Gathered objects from one or more profiles, validate the results.
		s.State = report.OK
		for _, result := range c.S3Results {
			validated := c.validatedS3Profile(result)
			s.Value = append(s.Value, validated)
		}
	} else {
		// Failed to get S3 profiles or application prefix from the gathered hub data.
		s.State = report.Problem
		s.Description = "S3 data not available"
	}

	summary.AddValidation(c.Report.Summary, s)
}

func (c *Command) validatedS3Profile(
	result s3.Result,
) report.ApplicationS3ProfileStatus {
	profileStatus := report.ApplicationS3ProfileStatus{
		Name: result.ProfileName,
	}

	if result.Err != nil {
		profileStatus.Gathered = report.ValidatedBool{
			Validated: report.Validated{
				State:       report.Problem,
				Description: result.Err.Error(),
			},
			Value: false,
		}
	} else {
		profileStatus.Gathered = report.ValidatedBool{
			Validated: report.Validated{
				State: report.OK,
			},
			Value: true,
		}
	}

	summary.AddValidation(c.Report.Summary, &profileStatus.Gathered)
	return profileStatus
}

func (c *Command) validateDRPC(
	s *report.DRPCSummary,
	drpc *ramenapi.DRPlacementControl,
) {
	s.Name = drpc.Name
	s.Namespace = drpc.Namespace

	var err error

	s.ClusterTime, err = ramen.ClusterTime(drpc.Annotations)
	if err != nil {
		c.Logger().Warnf("drpc \"%s/%s\": %s", drpc.Namespace, drpc.Name, err)
	}

	s.Deleted = c.ValidatedDeleted(drpc)
	s.DRPolicy = drpc.Spec.DRPolicyRef.Name
	s.SchedulingInterval = c.validatedDRPCSchedulingInterval(drpc)
	s.LastGroupSyncTime = c.validateLastGroupSyncTime(
		drpc.Status.LastGroupSyncTime, s.ClusterTime, s.SchedulingInterval, true)
	s.Action = c.validatedDRPCAction(string(drpc.Spec.Action))
	s.Phase = c.validatedDRPCPhase(drpc)
	s.Progression = c.validatedDRPCProgression(drpc)
	s.Conditions = c.ValidatedConditions(drpc, drpc.Status.Conditions)
}

func (c *Command) validateVRG(
	s *report.VRGSummary,
	cluster *e2etypes.Cluster,
	drpc *ramenapi.DRPlacementControl,
	stableState ramenapi.State,
) error {
	log := c.Logger()
	reader := c.OutputReader(cluster.Name)
	vrgName := drpc.Name
	vrgNamespace := ramen.VRGNamespace(drpc)

	vrg, err := ramen.ReadVRG(reader, vrgName, vrgNamespace)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to read vrg from cluster %q: %w", cluster.Name, err)
		}

		log.Debugf("vrg \"%s/%s\" missing in cluster %q", vrgNamespace, vrgName, cluster.Name)
		s.Name = vrgName
		s.Namespace = vrgNamespace
		s.Deleted = c.ValidatedDeleted(nil)
		return nil
	}

	log.Debugf("Read vrg \"%s/%s\" from cluster %q", vrgNamespace, vrgName, cluster.Name)
	s.Name = vrgName
	s.Namespace = vrgNamespace

	s.ClusterTime, err = ramen.ClusterTime(vrg.Annotations)
	if err != nil {
		log.Warnf("vrg \"%s/%s\" on cluster %q: %s", vrgNamespace, vrgName, cluster.Name, err)
	}

	s.Deleted = c.ValidatedDeleted(vrg)
	s.SchedulingInterval = c.validatedVRGSchedulingInterval(
		vrg,
		&c.Report.ApplicationStatus.Hub.DRPC,
	)
	s.LastGroupSyncTime = c.validateLastGroupSyncTime(
		vrg.Status.LastGroupSyncTime, s.ClusterTime, s.SchedulingInterval,
		stableState == ramenapi.PrimaryState)
	s.Conditions = c.validatedVRGConditions(vrg)
	s.ProtectedPVCs = c.validatedProtectedPVCs(cluster, vrg)
	s.PVCGroups = c.pvcGroups(vrg)
	s.State = c.validatedVRGState(vrg, stableState)

	return nil
}

func (c *Command) validatedDRPCPhase(drpc *ramenapi.DRPlacementControl) report.ValidatedString {
	validated := report.ValidatedString{Value: string(drpc.Status.Phase)}

	// We expect stable phase as ok, and anything else as an error. An application is not expected
	// to be in unstable phase (e.g. FailingOver) for a long time. The stable phase depends on the
	// action.

	stablePhase, err := ramen.StablePhase(drpc.Spec.Action)
	if err != nil {
		validated.State = report.Problem
		validated.Description = err.Error()
	} else {
		if drpc.Status.Phase != stablePhase {
			validated.State = report.Problem
			validated.Description = fmt.Sprintf("Waiting for stable phase %q", stablePhase)
		} else {
			validated.State = report.OK
		}
	}

	summary.AddValidation(c.Report.Summary, &validated)
	return validated
}

func (c *Command) validatedDRPCProgression(
	drpc *ramenapi.DRPlacementControl,
) report.ValidatedString {
	validated := report.ValidatedString{Value: string(drpc.Status.Progression)}

	// We expect a stable progression (Completed). An application should not be in unstable state
	// for long time, so it we see unstable progression it requires investigation.
	if drpc.Status.Progression != ramenapi.ProgressionCompleted {
		validated.State = report.Problem
		validated.Description = fmt.Sprintf(
			"Waiting for progression %q",
			ramenapi.ProgressionCompleted,
		)
	} else {
		validated.State = report.OK
	}

	summary.AddValidation(c.Report.Summary, &validated)
	return validated
}

func (c *Command) validatedVRGState(
	vrg *ramenapi.VolumeReplicationGroup,
	stableState ramenapi.State,
) report.ValidatedString {
	validated := report.ValidatedString{Value: string(vrg.Status.State)}

	// We expect the stable state. An application should not be in unstable state for long time, so
	// it we see unstable state it requires investigation.
	if vrg.Status.State != stableState {
		validated.State = report.Problem
		validated.Description = fmt.Sprintf("Waiting to become %q", stableState)
	} else {
		validated.State = report.OK
	}

	summary.AddValidation(c.Report.Summary, &validated)
	return validated
}

func (c *Command) validatedProtectedPVCPhase(
	pvc *corev1.PersistentVolumeClaim,
) report.ValidatedString {
	validated := report.ValidatedString{Value: string(pvc.Status.Phase)}

	// Protected PVC must be bound; anything else seen for long time requires investigation.
	if pvc.Status.Phase != corev1.ClaimBound {
		validated.State = report.Problem
		validated.Description = fmt.Sprintf("PVC is not %q", corev1.ClaimBound)
	} else {
		validated.State = report.OK
	}

	summary.AddValidation(c.Report.Summary, &validated)
	return validated
}

func (c *Command) validatedDRPCAction(action string) report.ValidatedString {
	validated := report.ValidatedString{Value: action}
	if slices.Contains(ramen.Actions, action) {
		validated.State = report.OK
	} else {
		validated.State = report.Problem
		validated.Description = fmt.Sprintf("Unknown action %q", action)
	}
	summary.AddValidation(c.Report.Summary, &validated)
	return validated
}

func (c *Command) validatedDRPCSchedulingInterval(
	drpc *ramenapi.DRPlacementControl,
) report.ValidatedDuration {
	log := c.Logger()
	reader := c.OutputReader(c.Env().Hub.Name)

	drPolicy, err := ramen.ReadDRPolicy(reader, drpc.Spec.DRPolicyRef.Name)
	if err != nil {
		log.Warnf("Failed to read drpolicy %q: %s", drpc.Spec.DRPolicyRef.Name, err)

		return c.validatedSchedulingInterval(0,
			fmt.Sprintf("Could not read drpolicy %q", drpc.Spec.DRPolicyRef.Name))
	}

	if drPolicy.Spec.SchedulingInterval == "" {
		return c.validatedSchedulingInterval(0,
			fmt.Sprintf("Missing scheduling interval in drpolicy %q", drPolicy.Name))
	}

	duration, err := ramen.ParseSchedulingInterval(drPolicy.Spec.SchedulingInterval)
	if err != nil {
		log.Warnf("Invalid scheduling interval in drpolicy %q: %s", drPolicy.Name, err)

		return c.validatedSchedulingInterval(0,
			fmt.Sprintf("Invalid scheduling interval in drpolicy %q", drPolicy.Name))
	}

	return c.validatedSchedulingInterval(duration, "")
}

func (c *Command) validatedVRGSchedulingInterval(
	vrg *ramenapi.VolumeReplicationGroup,
	drpc *report.DRPCSummary,
) report.ValidatedDuration {
	// Metro DR does not use async replication.
	if vrg.Spec.Async == nil {
		return report.ValidatedDuration{}
	}

	if vrg.Spec.Async.SchedulingInterval == "" {
		return c.validatedSchedulingInterval(0, "Missing scheduling interval in vrg")
	}

	duration, err := ramen.ParseSchedulingInterval(vrg.Spec.Async.SchedulingInterval)
	if err != nil {
		c.Logger().Warnf("Invalid scheduling interval in vrg: %s", err)

		return c.validatedSchedulingInterval(0, "Invalid scheduling interval in vrg")
	}

	if drpc.SchedulingInterval.State == report.OK && duration != drpc.SchedulingInterval.Value {
		return c.validatedSchedulingInterval(duration,
			fmt.Sprintf("Does not match drpolicy %q interval %s",
				drpc.DRPolicy, drpc.SchedulingInterval.Value))
	}

	return c.validatedSchedulingInterval(duration, "")
}

func (c *Command) validatedSchedulingInterval(
	duration stdtime.Duration,
	description string,
) report.ValidatedDuration {
	validated := report.ValidatedDuration{Value: duration}

	if description != "" {
		validated.State = report.Problem
		validated.Description = description
	} else {
		validated.State = report.OK
	}

	summary.AddValidation(c.Report.Summary, &validated)

	return validated
}

// validateLastGroupSyncTime checks if replication is fresh by comparing
// lastGroupSyncTime with clusterTime and schedulingInterval. This is the
// shared logic for DRPC and VRG, matching the VolumeSynchronizationDelay
// alert in ramen.
func (c *Command) validateLastGroupSyncTime(
	lastGroupSyncTime *metav1.Time,
	clusterTime *stdtime.Time,
	schedulingInterval report.ValidatedDuration,
	primary bool,
) report.ValidatedTime {
	if lastGroupSyncTime == nil {
		if primary {
			return c.validatedLastGroupSyncTime(nil, report.Warning,
				"Waiting for first volume synchronization")
		}

		return c.validatedLastGroupSyncTime(nil, report.OK, "")
	}

	// metav1.Time.UnmarshalJSON converts timestamps to local time. We convert to UTC
	// for consistency with other timestamps in YAML and HTML reports.
	// https://github.com/kubernetes/kubernetes/issues/102316
	t := lastGroupSyncTime.UTC()

	// Cannot validate without cluster time or valid scheduling interval - should not happen.
	if clusterTime == nil ||
		schedulingInterval.State != report.OK ||
		schedulingInterval.Value == 0 {
		return report.ValidatedTime{Value: &t}
	}

	// Thresholds match ramen VolumeSynchronizationDelay alert (>= 3 critical, > 2 warning).
	intervals := float64(clusterTime.Sub(t)) / float64(schedulingInterval.Value)

	if intervals >= 3 {
		return c.validatedLastGroupSyncTime(&t, report.Problem,
			"Replication is exceeding 3x the scheduling interval")
	}

	if intervals > 2 {
		return c.validatedLastGroupSyncTime(&t, report.Warning,
			"Replication is exceeding 2x the scheduling interval")
	}

	return c.validatedLastGroupSyncTime(&t, report.OK, "")
}

func (c *Command) validatedLastGroupSyncTime(
	value *stdtime.Time,
	state report.ValidationState,
	description string,
) report.ValidatedTime {
	validated := report.ValidatedTime{
		Validated: report.Validated{State: state, Description: description},
		Value:     value,
	}

	if state != "" {
		summary.AddValidation(c.Report.Summary, &validated)
	}

	return validated
}

func (c *Command) validatedProtectedPVCs(
	cluster *e2etypes.Cluster,
	vrg *ramenapi.VolumeReplicationGroup,
) []report.ProtectedPVCSummary {
	log := c.Logger()

	// Protected PVCs becomes stale on a secondary cluster:
	// https://github.com/RamenDR/ramenctl/issues/286.
	if vrg.Status.State == ramenapi.SecondaryState {
		log.Debugf(
			"Skipping protected pvcs on cluster %q for vrg state %q",
			cluster.Name, vrg.Status.State,
		)
		return nil
	}

	reader := c.OutputReader(cluster.Name)
	var protectedPVCs []report.ProtectedPVCSummary

	for i := range vrg.Status.ProtectedPVCs {
		ppvc := &vrg.Status.ProtectedPVCs[i]
		ps := report.ProtectedPVCSummary{
			Name:        ppvc.Name,
			Namespace:   ppvc.Namespace,
			Replication: c.protectedPVCReplication(ppvc),
			Conditions:  c.validatedProtectedPVCConditions(vrg, ppvc),
		}

		if pvc, err := core.ReadPVC(reader, ppvc.Name, ppvc.Namespace); err != nil {
			log.Warnf("failed to read pvc \"%s/%s\" from cluster %q: %s",
				ppvc.Namespace, ppvc.Name, cluster.Name, err)
			ps.Deleted = c.ValidatedDeleted(nil)
		} else {
			log.Debugf("Read pvc \"%s/%s\" from cluster %q", pvc.Namespace, pvc.Name, cluster.Name)
			ps.Deleted = c.ValidatedDeleted(pvc)
			ps.Phase = c.validatedProtectedPVCPhase(pvc)
		}

		protectedPVCs = append(protectedPVCs, ps)
	}

	return protectedPVCs
}

func (c *Command) validatedVRGConditions(
	vrg *ramenapi.VolumeReplicationGroup,
) []report.ValidatedCondition {
	var conditions []report.ValidatedCondition
	for i := range vrg.Status.Conditions {
		condition := &vrg.Status.Conditions[i]
		// On the secondary cluster most conditions are unused.
		if condition.Reason == ramen.VRGConditionReasonUnused {
			continue
		}
		// DataProtected behaves differently for volrep and volsync. Since a workload can have both
		// volsync protected pvcs and volrep protected pvcs we seem to have now way to validate this
		// condition.
		if condition.Type == ramen.VRGConditionTypeDataProtected {
			continue
		}
		if !ramen.IsKnownVRGCondition(condition.Type) {
			c.Logger().Warnf("Skipping validation for unknown VRG condition: %+v", *condition)
			continue
		}
		validated := validatecmd.ValidatedCondition(vrg, condition, metav1.ConditionTrue)
		summary.AddValidation(c.Report.Summary, &validated)
		conditions = append(conditions, validated)
	}
	return conditions
}

func (c *Command) protectedPVCReplication(ppvc *ramenapi.ProtectedPVC) report.ReplicationType {
	// TODO: report external replication.
	if ppvc.ProtectedByVolSync {
		return report.Volsync
	}
	return report.Volrep
}

func (c *Command) validatedProtectedPVCConditions(
	vrg *ramenapi.VolumeReplicationGroup,
	ppvc *ramenapi.ProtectedPVC,
) []report.ValidatedCondition {
	log := c.Logger()

	var conditions []report.ValidatedCondition
	for i := range ppvc.Conditions {
		condition := &ppvc.Conditions[i]

		// DataProtected exists only with volrep and has confusing and unhelpful semantics. Status
		// is False in the stable state and True during some part of Relocate phase.
		if condition.Type == ramen.VRGConditionTypeDataProtected {
			continue
		}

		// Volsync PVsRestored condition is always stale on the primary after failover or
		// relocate, but the application is fine.
		if condition.Type == ramen.VRGConditionTypeVolSyncPVsRestored &&
			condition.ObservedGeneration != vrg.Generation {
			log.Debugf(
				"Skipping stale protected PVC condition: observed generation %d does not match vrg generation: %+v",
				condition.ObservedGeneration,
				vrg.Generation,
				condition,
			)
			continue
		}

		validated := validatecmd.ValidatedCondition(vrg, condition, metav1.ConditionTrue)
		summary.AddValidation(c.Report.Summary, &validated)
		conditions = append(conditions, validated)
	}
	return conditions
}

func (c *Command) pvcGroups(vrg *ramenapi.VolumeReplicationGroup) []report.PVCGroupsSummary {
	if len(vrg.Status.PVCGroups) == 0 {
		return nil
	}

	groups := make([]report.PVCGroupsSummary, 0, len(vrg.Status.PVCGroups))
	for _, group := range vrg.Status.PVCGroups {
		if len(group.Grouped) > 0 {
			groups = append(groups, report.PVCGroupsSummary{Grouped: group.Grouped})
		}
	}
	return groups
}
