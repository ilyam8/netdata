// SPDX-License-Identifier: GPL-3.0-or-later

package execution

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	ndexec "github.com/netdata/netdata/go/plugins/plugin/go.d/pkg/ndexec"
	"github.com/netdata/netdata/go/plugins/plugin/scripts.d/collector/nagios/internal/runtime"
	"github.com/netdata/netdata/go/plugins/plugin/scripts.d/collector/nagios/internal/spec"
)

func TestV2Gate_G4_ExecutionService(t *testing.T) {
	tests := map[string]struct {
		cfg ExecutionServiceConfig
		run func(*testing.T, *Service)
	}{
		"concurrent attach detach": {
			cfg: ExecutionServiceConfig{Workers: 4, QueueSize: 16},
			run: func(t *testing.T, svc *Service) {
				const workers = 8
				const loops = 12

				errCh := make(chan error, 1)
				reportErr := func(err error) {
					if err == nil {
						return
					}
					select {
					case errCh <- err:
					default:
					}
				}

				stopSnapshots := make(chan struct{})
				var snapshotsWG sync.WaitGroup
				snapshotsWG.Add(1)
				go func() {
					defer snapshotsWG.Done()
					for {
						select {
						case <-stopSnapshots:
							return
						default:
						}
						snapshot := svc.Snapshot()
						if int(snapshot.Scheduled) != len(snapshot.Jobs) {
							reportErr(fmt.Errorf("snapshot mismatch scheduled=%d jobs=%d", snapshot.Scheduled, len(snapshot.Jobs)))
							return
						}
						time.Sleep(2 * time.Millisecond)
					}
				}()

				var wg sync.WaitGroup
				for i := 0; i < workers; i++ {
					wg.Add(1)
					go func(worker int) {
						defer wg.Done()
						for j := 0; j < loops; j++ {
							handle, err := svc.Attach(gateJobRegistration(fmt.Sprintf("concurrent-%d-%d", worker, j), time.Hour, gateQuickRunner))
							if err != nil {
								reportErr(err)
								return
							}
							svc.Detach(handle)
							svc.Detach(handle)
						}
					}(i)
				}
				wg.Wait()
				close(stopSnapshots)
				snapshotsWG.Wait()

				select {
				case err := <-errCh:
					t.Fatal(err)
				default:
				}

				snapshot := svc.Snapshot()
				if snapshot.Scheduled != 0 || len(snapshot.Jobs) != 0 {
					t.Fatalf("expected no remaining jobs after churn, got %+v", snapshot)
				}
			},
		},
		"destructive update interaction": {
			cfg: ExecutionServiceConfig{Workers: 4, QueueSize: 16},
			run: func(t *testing.T, svc *Service) {
				handle, err := svc.Attach(gateJobRegistration("update-job", time.Hour, gateQuickRunner))
				if err != nil {
					t.Fatalf("initial attach: %v", err)
				}

				errCh := make(chan error, 1)
				reportErr := func(err error) {
					if err == nil {
						return
					}
					select {
					case errCh <- err:
					default:
					}
				}

				var noiseWG sync.WaitGroup
				for i := 0; i < 4; i++ {
					noiseWG.Add(1)
					go func(worker int) {
						defer noiseWG.Done()
						for j := 0; j < 10; j++ {
							noiseHandle, err := svc.Attach(gateJobRegistration(fmt.Sprintf("noise-%d-%d", worker, j), time.Hour, gateQuickRunner))
							if err != nil {
								reportErr(err)
								return
							}
							svc.Detach(noiseHandle)
						}
					}(i)
				}

				for i := 0; i < 12; i++ {
					svc.Detach(handle)
					handle, err = svc.Attach(gateJobRegistration("update-job", time.Hour, gateQuickRunner))
					if err != nil {
						t.Fatalf("reattach update-job: %v", err)
					}
				}
				noiseWG.Wait()

				select {
				case err := <-errCh:
					t.Fatal(err)
				default:
				}

				snapshot := svc.Snapshot()
				if got := gateCountJobsByName(snapshot, "update-job"); got != 1 {
					t.Fatalf("expected exactly one active update-job, got %d", got)
				}

				svc.Detach(handle)
				snapshot = svc.Snapshot()
				if got := gateCountJobsByName(snapshot, "update-job"); got != 0 {
					t.Fatalf("expected update-job to be fully detached, got %d", got)
				}
				if snapshot.Scheduled != 0 || len(snapshot.Jobs) != 0 {
					t.Fatalf("expected no orphan jobs after cleanup, got %+v", snapshot)
				}
			},
		},
		"shared service backlog behavior": {
			cfg: ExecutionServiceConfig{Workers: 1, QueueSize: 4},
			run: func(t *testing.T, svc *Service) {
				slowHandle, err := svc.Attach(gateJobRegistration("slow-job", 10*time.Millisecond, gateSlowRunner(80*time.Millisecond)))
				if err != nil {
					t.Fatalf("attach slow job: %v", err)
				}
				defer svc.Detach(slowHandle)

				fastHandle, err := svc.Attach(gateJobRegistration("fast-job", 10*time.Millisecond, gateQuickRunner))
				if err != nil {
					t.Fatalf("attach fast job: %v", err)
				}
				defer svc.Detach(fastHandle)

				waitForSnapshot(t, 3*time.Second, func(snapshot Snapshot) bool {
					return snapshot.Scheduled == 2 && snapshot.Started > 0 && snapshot.Finished > 0 && snapshot.Skipped > 0
				}, svc)

				snapshot := svc.Snapshot()
				if got := gateCountJobsByName(snapshot, "slow-job"); got != 1 {
					t.Fatalf("expected one slow-job registration, got %d", got)
				}
				if got := gateCountJobsByName(snapshot, "fast-job"); got != 1 {
					t.Fatalf("expected one fast-job registration, got %d", got)
				}

				svc.Detach(slowHandle)
				svc.Detach(fastHandle)
				waitForSnapshot(t, time.Second, func(snapshot Snapshot) bool {
					return snapshot.Scheduled == 0 && len(snapshot.Jobs) == 0
				}, svc)
			},
		},
		"detach cancels queued stale work": {
			cfg: ExecutionServiceConfig{Workers: 1, QueueSize: 4},
			run: func(t *testing.T, svc *Service) {
				var staleRuns atomic.Int64

				blockerStarted := make(chan struct{})
				releaseBlocker := make(chan struct{})

				blockerHandle, err := svc.Attach(gateJobRegistration("blocker", 10*time.Millisecond, gateBlockingRunner(blockerStarted, releaseBlocker)))
				if err != nil {
					t.Fatalf("attach blocker: %v", err)
				}
				defer svc.Detach(blockerHandle)

				waitForSnapshot(t, time.Second, func(snapshot Snapshot) bool {
					return snapshot.Started > 0
				}, svc)

				victimHandle, err := svc.Attach(gateJobRegistration("victim", 10*time.Millisecond, gateCountingRunner(&staleRuns)))
				if err != nil {
					t.Fatalf("attach victim: %v", err)
				}

				waitForSnapshot(t, time.Second, func(snapshot Snapshot) bool {
					return snapshot.Queued > 0 && gateCountJobsByName(snapshot, "victim") == 1
				}, svc)

				svc.Detach(victimHandle)

				close(releaseBlocker)
				waitForSnapshot(t, time.Second, func(snapshot Snapshot) bool {
					return snapshot.Queued == 0 && gateCountJobsByName(snapshot, "victim") == 0
				}, svc)

				deadline := time.Now().Add(250 * time.Millisecond)
				for time.Now().Before(deadline) {
					if staleRuns.Load() != 0 {
						t.Fatalf("expected detached queued job not to execute, got %d runs", staleRuns.Load())
					}
					time.Sleep(10 * time.Millisecond)
				}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc, err := NewExecutionService(test.cfg)
			if err != nil {
				t.Fatalf("NewExecutionService() error = %v", err)
			}
			defer svc.close()
			test.run(t, svc)
		})
	}
}

func gateJobRegistration(name string, interval time.Duration, runner func(context.Context, runtime.JobRuntime, time.Duration) ([]byte, string, ndexec.ResourceUsage, error)) Registration {
	return Registration{
		Spec: spec.JobSpec{
			Name:             name,
			Plugin:           "/bin/true",
			CheckInterval:    interval,
			RetryInterval:    interval,
			Timeout:          time.Second,
			MaxCheckAttempts: 1,
		},
		Runner:  runner,
		Emitter: runtime.NewNoopEmitter(),
	}
}

func gateQuickRunner(context.Context, runtime.JobRuntime, time.Duration) ([]byte, string, ndexec.ResourceUsage, error) {
	return nil, "", ndexec.ResourceUsage{}, nil
}

func gateSlowRunner(delay time.Duration) func(context.Context, runtime.JobRuntime, time.Duration) ([]byte, string, ndexec.ResourceUsage, error) {
	return func(context.Context, runtime.JobRuntime, time.Duration) ([]byte, string, ndexec.ResourceUsage, error) {
		time.Sleep(delay)
		return nil, "", ndexec.ResourceUsage{}, nil
	}
}

func gateBlockingRunner(started chan<- struct{}, release <-chan struct{}) func(context.Context, runtime.JobRuntime, time.Duration) ([]byte, string, ndexec.ResourceUsage, error) {
	var once sync.Once

	return func(context.Context, runtime.JobRuntime, time.Duration) ([]byte, string, ndexec.ResourceUsage, error) {
		once.Do(func() { close(started) })
		<-release
		return nil, "", ndexec.ResourceUsage{}, nil
	}
}

func gateCountingRunner(counter *atomic.Int64) func(context.Context, runtime.JobRuntime, time.Duration) ([]byte, string, ndexec.ResourceUsage, error) {
	return func(context.Context, runtime.JobRuntime, time.Duration) ([]byte, string, ndexec.ResourceUsage, error) {
		counter.Add(1)
		return nil, "", ndexec.ResourceUsage{}, nil
	}
}

func gateCountJobsByName(snapshot Snapshot, jobName string) int {
	count := 0
	for _, job := range snapshot.Jobs {
		if job.JobName == jobName {
			count++
		}
	}
	return count
}
