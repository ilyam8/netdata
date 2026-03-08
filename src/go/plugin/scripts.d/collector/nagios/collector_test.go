// SPDX-License-Identifier: GPL-3.0-or-later

package nagios

import (
	"context"
	"testing"
	"time"

	"github.com/netdata/netdata/go/plugins/pkg/metrix"
	"github.com/netdata/netdata/go/plugins/plugin/framework/chartengine"
	"github.com/netdata/netdata/go/plugins/plugin/framework/charttpl"
	"github.com/netdata/netdata/go/plugins/plugin/framework/runtimecomp"
	"github.com/netdata/netdata/go/plugins/plugin/go.d/pkg/collecttest"
	"github.com/netdata/netdata/go/plugins/plugin/scripts.d/collector/nagios/internal/execution"
	"github.com/netdata/netdata/go/plugins/plugin/scripts.d/collector/nagios/internal/output"
	"github.com/netdata/netdata/go/plugins/plugin/scripts.d/collector/nagios/internal/spec"
)

func TestCollector_ChartTemplateYAML(t *testing.T) {
	templateYAML := New().ChartTemplateYAML()
	collecttest.AssertChartTemplateSchema(t, templateYAML)

	specYAML, err := charttpl.DecodeYAML([]byte(templateYAML))
	if err != nil {
		t.Fatalf("decode template: %v", err)
	}
	if err := specYAML.Validate(); err != nil {
		t.Fatalf("validate template: %v", err)
	}
	if _, err := chartengine.Compile(specYAML, 1); err != nil {
		t.Fatalf("compile template: %v", err)
	}
}

func TestCollector_InitCollectCleanup(t *testing.T) {
	svc := newFakeExecutionService()
	coll := NewWithExecutionService(svc)
	coll.Config.JobConfig = spec.JobConfig{
		Name:   "check_disk",
		Plugin: "/bin/true",
	}

	if err := coll.Init(context.Background()); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if len(svc.attached) != 1 {
		t.Fatalf("expected one attach call, got %d", len(svc.attached))
	}
	if got := svc.attached[0].Spec.Name; got != "check_disk" {
		t.Fatalf("attached job name = %q, want %q", got, "check_disk")
	}

	svc.snapshot = execution.Snapshot{
		Jobs: []execution.JobSnapshot{
			{
				JobID:      "",
				JobName:    "check_disk",
				State:      "OK",
				Attempt:    1,
				MaxAttempt: 3,
				PerfSamples: []output.PerfDatum{
					{Label: "used", Unit: "KB", Value: 30},
				},
			},
		},
	}

	cc := mustCycleController(t, coll.MetricStore())
	cc.BeginCycle()
	if err := coll.Collect(context.Background()); err != nil {
		cc.AbortCycle()
		t.Fatalf("collect failed: %v", err)
	}
	cc.CommitCycleSuccess()

	read := coll.MetricStore().Read(metrix.ReadRaw())
	flat := coll.MetricStore().Read(metrix.ReadFlatten())
	assertMetricMissing(t, read, "nagios.scheduler.running", metrix.Labels{})
	assertMetricValue(t, flat, "nagios.job.state", metrix.Labels{"nagios_job": "check_disk", "nagios.job.state": "ok"}, 1)
	assertMetricValue(t, flat, "nagios.true.bytes_used_value", metrix.Labels{"nagios_job": "check_disk", metrix.MeasureSetFieldLabel: "value"}, 30000)
	point, ok := read.MeasureSet("nagios.true.bytes_used", metrix.Labels{"nagios_job": "check_disk"})
	if !ok {
		t.Fatalf("expected raw measureset for nagios.true.bytes_used")
	}
	if got := point.Values[0]; got != 30000 {
		t.Fatalf("unexpected raw measureset value: got=%f want=%f", got, float64(30000))
	}
	meta, ok := flat.MetricMeta("nagios.true.bytes_used_value")
	if !ok {
		t.Fatalf("expected metric metadata for nagios.true.bytes_used_value")
	}
	if meta.Unit != "bytes" {
		t.Fatalf("unexpected unit metadata: %q", meta.Unit)
	}
	if meta.ChartFamily != "true" {
		t.Fatalf("unexpected chart family metadata: %q", meta.ChartFamily)
	}
	if !meta.Float {
		t.Fatalf("expected float metadata to be true")
	}

	coll.Cleanup(context.Background())
	if svc.detached != 1 {
		t.Fatalf("expected one detach call, got %d", svc.detached)
	}
}

func TestCollector_NewUsesSharedExecutionService(t *testing.T) {
	coll := New()
	if coll.service != sharedExecutionService {
		t.Fatalf("expected New() to use shared execution service")
	}
}

type fakeExecutionService struct {
	attached []execution.Registration
	detached int
	snapshot execution.Snapshot
}

func newFakeExecutionService() *fakeExecutionService {
	return &fakeExecutionService{}
}

func (f *fakeExecutionService) Attach(reg execution.Registration) (*execution.JobHandle, error) {
	f.attached = append(f.attached, reg)
	return &execution.JobHandle{}, nil
}

func (f *fakeExecutionService) Detach(_ *execution.JobHandle) {
	f.detached++
}

func (f *fakeExecutionService) Snapshot() execution.Snapshot {
	return f.snapshot
}

func (f *fakeExecutionService) BindRuntimeService(runtimecomp.Service) error {
	return nil
}

func mustCycleController(t *testing.T, store metrix.CollectorStore) metrix.CycleController {
	t.Helper()
	managed, ok := metrix.AsCycleManagedStore(store)
	if !ok {
		t.Fatalf("store does not expose cycle control")
	}
	return managed.CycleController()
}

func assertMetricValue(t *testing.T, r metrix.Reader, name string, labels metrix.Labels, want float64) {
	t.Helper()
	got, ok := r.Value(name, labels)
	if !ok {
		t.Fatalf("missing metric %s labels=%v", name, labels)
	}
	if diff := got - want; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("metric mismatch %s labels=%v got=%f want=%f", name, labels, got, want)
	}
}

func assertMetricMissing(t *testing.T, r metrix.Reader, name string, labels metrix.Labels) {
	t.Helper()
	if _, ok := r.Value(name, labels); ok {
		t.Fatalf("unexpected metric %s labels=%v", name, labels)
	}
}

func TestCollector_CheckValidatesConfig(t *testing.T) {
	coll := NewWithExecutionService(newFakeExecutionService())
	coll.Config.JobConfig = spec.JobConfig{Name: "invalid-without-plugin"}
	if err := coll.Check(context.Background()); err == nil {
		t.Fatalf("expected check to fail for missing plugin")
	}

	coll.Config.JobConfig.Plugin = "/bin/true"
	if err := coll.Check(context.Background()); err != nil {
		t.Fatalf("expected check to pass, got %v", err)
	}
}

func TestCollector_DefaultTimingDefaultsFromSpec(t *testing.T) {
	coll := NewWithExecutionService(newFakeExecutionService())
	coll.Config.JobConfig = spec.JobConfig{
		Name:   "defaults",
		Plugin: "/bin/true",
	}
	if err := coll.Init(context.Background()); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if coll.jobSpec.CheckInterval != 5*time.Minute {
		t.Fatalf("unexpected check interval default: %s", coll.jobSpec.CheckInterval)
	}
	if coll.jobSpec.RetryInterval != time.Minute {
		t.Fatalf("unexpected retry interval default: %s", coll.jobSpec.RetryInterval)
	}
}
