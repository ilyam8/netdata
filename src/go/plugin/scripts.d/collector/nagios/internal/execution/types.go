// SPDX-License-Identifier: GPL-3.0-or-later

package execution

import (
	"time"

	"github.com/netdata/netdata/go/plugins/plugin/scripts.d/collector/nagios/internal/output"
	"github.com/netdata/netdata/go/plugins/plugin/scripts.d/collector/nagios/internal/runtime"
	"github.com/netdata/netdata/go/plugins/plugin/scripts.d/collector/nagios/internal/spec"
	"github.com/netdata/netdata/go/plugins/plugin/scripts.d/pkg/timeperiod"
)

// Registration describes one Nagios job attached to the shared execution service.
type Registration struct {
	Spec             spec.JobSpec
	Runner           runtime.JobRunner
	RegisterPerfdata func(spec.JobSpec, output.PerfDatum)
	Periods          *timeperiod.Set
	UserMacros       map[string]string
	Vnode            runtime.VnodeInfo
}

// Snapshot captures the public execution-service view over the internal runtime scheduler.
type Snapshot struct {
	Running   int64
	Queued    int64
	Scheduled int64
	Started   uint64
	Finished  uint64
	Skipped   uint64
	Next      time.Duration

	Jobs []JobSnapshot
}

// JobSnapshot contains one job's exported execution/runtime state.
type JobSnapshot struct {
	JobID   string
	JobName string

	State      string
	Attempt    int
	MaxAttempt int

	Running     bool
	Retrying    bool
	PeriodSkip  bool
	CPUMissing  bool
	Duration    time.Duration
	CPUTime     time.Duration
	RSS         int64
	DiskRead    int64
	DiskWrite   int64
	PerfSamples []output.PerfDatum
}

func registrationToRuntime(reg Registration) runtime.JobRegistration {
	return runtime.JobRegistration{
		Spec:             reg.Spec,
		Runner:           reg.Runner,
		RegisterPerfdata: reg.RegisterPerfdata,
		Periods:          reg.Periods,
		UserMacros:       reg.UserMacros,
		Vnode:            reg.Vnode,
	}
}

func snapshotFromRuntime(src runtime.SchedulerSnapshot) Snapshot {
	out := Snapshot{
		Running:   src.Running,
		Queued:    src.Queued,
		Scheduled: src.Scheduled,
		Started:   src.Started,
		Finished:  src.Finished,
		Skipped:   src.Skipped,
		Next:      src.Next,
	}
	if len(src.Jobs) == 0 {
		return out
	}
	out.Jobs = make([]JobSnapshot, 0, len(src.Jobs))
	for _, job := range src.Jobs {
		out.Jobs = append(out.Jobs, JobSnapshot{
			JobID:       job.JobID,
			JobName:     job.JobName,
			State:       job.State,
			Attempt:     job.Attempt,
			MaxAttempt:  job.MaxAttempt,
			Running:     job.Running,
			Retrying:    job.Retrying,
			PeriodSkip:  job.PeriodSkip,
			CPUMissing:  job.CPUMissing,
			Duration:    job.Duration,
			CPUTime:     job.CPUTime,
			RSS:         job.RSS,
			DiskRead:    job.DiskRead,
			DiskWrite:   job.DiskWrite,
			PerfSamples: clonePerfDatumList(job.PerfSamples),
		})
	}
	return out
}

func clonePerfDatumList(src []output.PerfDatum) []output.PerfDatum {
	if len(src) == 0 {
		return nil
	}
	out := make([]output.PerfDatum, 0, len(src))
	for _, datum := range src {
		item := datum
		item.Min = cloneFloatPtr(datum.Min)
		item.Max = cloneFloatPtr(datum.Max)
		item.Warn = cloneThresholdRange(datum.Warn)
		item.Crit = cloneThresholdRange(datum.Crit)
		out = append(out, item)
	}
	return out
}

func cloneFloatPtr(v *float64) *float64 {
	if v == nil {
		return nil
	}
	cp := *v
	return &cp
}

func cloneThresholdRange(r *output.ThresholdRange) *output.ThresholdRange {
	if r == nil {
		return nil
	}
	return &output.ThresholdRange{
		Raw:       r.Raw,
		Inclusive: r.Inclusive,
		Low:       cloneFloatPtr(r.Low),
		High:      cloneFloatPtr(r.High),
	}
}
