// SPDX-License-Identifier: GPL-3.0-or-later

package execution

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/netdata/netdata/go/plugins/pkg/metrix"
	"github.com/netdata/netdata/go/plugins/plugin/framework/runtimecomp"
	"github.com/netdata/netdata/go/plugins/plugin/go.d/pkg/ndexec"
	"github.com/netdata/netdata/go/plugins/plugin/scripts.d/collector/nagios/internal/runtime"
	"github.com/netdata/netdata/go/plugins/plugin/scripts.d/collector/nagios/internal/spec"
)

type runtimeServiceMock struct {
	registered   []runtimecomp.ComponentConfig
	unregistered []string
	producers    map[string]func() error
	unprod       []string
}

func (m *runtimeServiceMock) RegisterComponent(cfg runtimecomp.ComponentConfig) error {
	m.registered = append(m.registered, cfg)
	return nil
}

func (m *runtimeServiceMock) UnregisterComponent(name string) {
	m.unregistered = append(m.unregistered, name)
}

func (m *runtimeServiceMock) RegisterProducer(name string, fn func() error) error {
	if m.producers == nil {
		m.producers = make(map[string]func() error)
	}
	m.producers[name] = fn
	return nil
}

func (m *runtimeServiceMock) UnregisterProducer(name string) {
	m.unprod = append(m.unprod, name)
	delete(m.producers, name)
}

func testJobSpec(name string) spec.JobSpec {
	return spec.JobSpec{
		Name:             name,
		Plugin:           "/bin/true",
		CheckInterval:    time.Hour,
		RetryInterval:    time.Hour,
		Timeout:          time.Second,
		MaxCheckAttempts: 1,
	}
}

func testJobRegistration(name string) Registration {
	return Registration{
		Spec: testJobSpec(name),
		Runner: func(context.Context, runtime.JobRuntime, time.Duration) ([]byte, string, ndexec.ResourceUsage, error) {
			return nil, "", ndexec.ResourceUsage{}, nil
		},
		Emitter: runtime.NewNoopEmitter(),
	}
}

func TestNewExecutionServiceDefaults(t *testing.T) {
	tests := map[string]struct {
		cfg           ExecutionServiceConfig
		wantWorkers   int
		wantQueueSize int
	}{
		"default zero values": {
			cfg:           ExecutionServiceConfig{},
			wantWorkers:   defaultWorkers,
			wantQueueSize: defaultQueueSize,
		},
		"explicit values": {
			cfg:           ExecutionServiceConfig{Workers: 4, QueueSize: 16},
			wantWorkers:   4,
			wantQueueSize: 16,
		},
		"negative values normalize": {
			cfg:           ExecutionServiceConfig{Workers: -1, QueueSize: -1},
			wantWorkers:   defaultWorkers,
			wantQueueSize: defaultQueueSize,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc, err := NewExecutionService(test.cfg)
			if err != nil {
				t.Fatalf("NewExecutionService() error = %v", err)
			}
			t.Cleanup(svc.close)

			if got := svc.workerCount(); got != test.wantWorkers {
				t.Fatalf("workerCount() = %d, want %d", got, test.wantWorkers)
			}
			if got := svc.queueCapacity(); got != test.wantQueueSize {
				t.Fatalf("queueCapacity() = %d, want %d", got, test.wantQueueSize)
			}

			snapshot := svc.Snapshot()
			if snapshot.Scheduled != 0 || len(snapshot.Jobs) != 0 {
				t.Fatalf("unexpected non-empty initial snapshot: %+v", snapshot)
			}
		})
	}
}

func TestExecutionServiceAttachDetachSnapshotLifecycle(t *testing.T) {
	svc, err := NewExecutionService(ExecutionServiceConfig{})
	if err != nil {
		t.Fatalf("NewExecutionService() error = %v", err)
	}
	t.Cleanup(svc.close)

	handle, err := svc.Attach(testJobRegistration("job-1"))
	if err != nil {
		t.Fatalf("Attach() error = %v", err)
	}
	if handle == nil || handle.JobID() == "" {
		t.Fatalf("Attach() returned invalid handle: %#v", handle)
	}

	snapshot := svc.Snapshot()
	if snapshot.Scheduled != 1 {
		t.Fatalf("snapshot.Scheduled = %d, want 1", snapshot.Scheduled)
	}
	if len(snapshot.Jobs) != 1 {
		t.Fatalf("len(snapshot.Jobs) = %d, want 1", len(snapshot.Jobs))
	}
	if snapshot.Jobs[0].JobID != handle.JobID() {
		t.Fatalf("snapshot job id = %q, want %q", snapshot.Jobs[0].JobID, handle.JobID())
	}
	if snapshot.Jobs[0].JobName != "job-1" {
		t.Fatalf("snapshot job name = %q, want %q", snapshot.Jobs[0].JobName, "job-1")
	}

	svc.Detach(handle)

	snapshot = svc.Snapshot()
	if snapshot.Scheduled != 0 {
		t.Fatalf("snapshot.Scheduled after detach = %d, want 0", snapshot.Scheduled)
	}
	if len(snapshot.Jobs) != 0 {
		t.Fatalf("len(snapshot.Jobs) after detach = %d, want 0", len(snapshot.Jobs))
	}
	if got := svc.attachedJobCount(); got != 0 {
		t.Fatalf("attachedJobCount() = %d, want 0", got)
	}
}

func TestExecutionServiceDetachNilOrUnknownHandleIsSafe(t *testing.T) {
	svc, err := NewExecutionService(ExecutionServiceConfig{})
	if err != nil {
		t.Fatalf("NewExecutionService() error = %v", err)
	}
	t.Cleanup(svc.close)

	svc.Detach(nil)
	svc.Detach(&JobHandle{jobID: "missing"})

	snapshot := svc.Snapshot()
	if snapshot.Scheduled != 0 || len(snapshot.Jobs) != 0 {
		t.Fatalf("unexpected snapshot after safe detaches: %+v", snapshot)
	}
}

func TestExecutionServiceCloseStopsActiveAndInactiveJobs(t *testing.T) {
	tests := map[string]struct {
		attach bool
	}{
		"no active jobs": {},
		"active job":     {attach: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc, err := NewExecutionService(ExecutionServiceConfig{})
			if err != nil {
				t.Fatalf("NewExecutionService() error = %v", err)
			}

			if test.attach {
				if _, err := svc.Attach(testJobRegistration("job-1")); err != nil {
					t.Fatalf("Attach() error = %v", err)
				}
			}

			svc.close()
			svc.close()

			snapshot := svc.Snapshot()
			if snapshot.Scheduled != 0 || len(snapshot.Jobs) != 0 {
				t.Fatalf("Snapshot() after close = %+v, want zero-ish value", snapshot)
			}

			_, err = svc.Attach(testJobRegistration("job-2"))
			if !errors.Is(err, errExecutionServiceClosed) {
				t.Fatalf("Attach() after close error = %v, want %v", err, errExecutionServiceClosed)
			}
		})
	}
}

func TestExecutionServiceBindRuntimeServiceLifecycle(t *testing.T) {
	svc, err := NewExecutionService(ExecutionServiceConfig{})
	if err != nil {
		t.Fatalf("NewExecutionService() error = %v", err)
	}

	mockSvc := &runtimeServiceMock{}
	if err := svc.BindRuntimeService(mockSvc); err != nil {
		t.Fatalf("BindRuntimeService() error = %v", err)
	}
	if err := svc.BindRuntimeService(mockSvc); err != nil {
		t.Fatalf("second BindRuntimeService() error = %v", err)
	}

	if len(mockSvc.registered) != 1 {
		t.Fatalf("len(registered) = %d, want 1", len(mockSvc.registered))
	}
	cfg := mockSvc.registered[0]
	if cfg.Name != executionRuntimeComponentName {
		t.Fatalf("component name = %q, want %q", cfg.Name, executionRuntimeComponentName)
	}
	if cfg.Store == nil {
		t.Fatalf("component store is nil")
	}
	if cfg.Module != "nagios_execution" {
		t.Fatalf("component module = %q, want %q", cfg.Module, "nagios_execution")
	}
	if cfg.JobName != "service" {
		t.Fatalf("component job name = %q, want %q", cfg.JobName, "service")
	}
	producerName := executionRuntimeComponentName + ".producer"
	if _, ok := mockSvc.producers[producerName]; !ok {
		t.Fatalf("producer %q not registered", producerName)
	}

	svc.close()

	if len(mockSvc.unprod) != 1 {
		t.Fatalf("len(unregistered producers) = %d, want 1", len(mockSvc.unprod))
	}
	if mockSvc.unprod[0] != producerName {
		t.Fatalf("unregistered producer = %q, want %q", mockSvc.unprod[0], producerName)
	}
	if len(mockSvc.unregistered) != 1 {
		t.Fatalf("len(unregistered) = %d, want 1", len(mockSvc.unregistered))
	}
	if mockSvc.unregistered[0] != executionRuntimeComponentName {
		t.Fatalf("unregistered component = %q, want %q", mockSvc.unregistered[0], executionRuntimeComponentName)
	}
}

func TestExecutionServiceSnapshotUpdatesRuntimeMetrics(t *testing.T) {
	svc, err := NewExecutionService(ExecutionServiceConfig{})
	if err != nil {
		t.Fatalf("NewExecutionService() error = %v", err)
	}
	t.Cleanup(svc.close)

	mockSvc := &runtimeServiceMock{}
	if err := svc.BindRuntimeService(mockSvc); err != nil {
		t.Fatalf("BindRuntimeService() error = %v", err)
	}

	handle, err := svc.Attach(Registration{
		Spec: spec.JobSpec{
			Name:             "job-1",
			Plugin:           "/bin/true",
			CheckInterval:    10 * time.Millisecond,
			RetryInterval:    10 * time.Millisecond,
			Timeout:          time.Second,
			MaxCheckAttempts: 1,
		},
		Runner: func(context.Context, runtime.JobRuntime, time.Duration) ([]byte, string, ndexec.ResourceUsage, error) {
			return []byte("OK"), "", ndexec.ResourceUsage{}, nil
		},
		Emitter: runtime.NewNoopEmitter(),
	})
	if err != nil {
		t.Fatalf("Attach() error = %v", err)
	}
	t.Cleanup(func() { svc.Detach(handle) })

	waitForSnapshot(t, time.Second, func(snapshot Snapshot) bool {
		return snapshot.Started > 0 && snapshot.Finished > 0
	}, svc)

	producer := mockSvc.producers[executionRuntimeComponentName+".producer"]
	if producer == nil {
		t.Fatalf("producer not registered")
	}
	waitForProducer(t, time.Second, producer, func() bool {
		reader := svc.runtimeStore.Read(metrix.ReadRaw())
		started, okStarted := reader.Value(executionRuntimeMetricPrefix+".started_total", nil)
		finished, okFinished := reader.Value(executionRuntimeMetricPrefix+".finished_total", nil)
		return okStarted && okFinished && started >= 1 && finished >= 1
	})

	reader := svc.runtimeStore.Read(metrix.ReadRaw())
	assertRuntimeMetricValue(t, reader, executionRuntimeMetricPrefix+".scheduled", nil, 1)
	if started, ok := reader.Value(executionRuntimeMetricPrefix+".started_total", nil); !ok || started < 1 {
		t.Fatalf("started_total = %f, ok=%v, want >=1", started, ok)
	}
	if finished, ok := reader.Value(executionRuntimeMetricPrefix+".finished_total", nil); !ok || finished < 1 {
		t.Fatalf("finished_total = %f, ok=%v, want >=1", finished, ok)
	}
}

func waitForSnapshot(t *testing.T, timeout time.Duration, fn func(Snapshot) bool, svc *Service) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		snapshot := svc.Snapshot()
		if fn(snapshot) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("condition not met before timeout")
}

func waitForProducer(t *testing.T, timeout time.Duration, tick func() error, fn func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := tick(); err != nil {
			t.Fatalf("producer tick error = %v", err)
		}
		if fn() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("producer condition not met before timeout")
}

func assertRuntimeMetricValue(t *testing.T, reader metrix.Reader, name string, labels metrix.Labels, want float64) {
	t.Helper()

	got, ok := reader.Value(name, labels)
	if !ok {
		t.Fatalf("metric %q not found (labels=%v)", name, labels)
	}
	if got != want {
		t.Fatalf("metric %q = %f, want %f", name, got, want)
	}
}
