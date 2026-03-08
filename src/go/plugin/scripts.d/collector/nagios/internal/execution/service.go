// SPDX-License-Identifier: GPL-3.0-or-later

package execution

import (
	"context"
	"errors"
	"sync"

	"github.com/netdata/netdata/go/plugins/logger"
	"github.com/netdata/netdata/go/plugins/pkg/metrix"
	"github.com/netdata/netdata/go/plugins/plugin/framework/runtimecomp"
	"github.com/netdata/netdata/go/plugins/plugin/scripts.d/collector/nagios/internal/runtime"
)

const (
	defaultWorkers   = 50
	defaultQueueSize = 128
)

var errExecutionServiceClosed = errors.New("execution service is closed")

type ExecutionService interface {
	Attach(reg Registration) (*JobHandle, error)
	Detach(handle *JobHandle)
	Snapshot() Snapshot
	BindRuntimeService(service runtimecomp.Service) error
}

type ExecutionServiceConfig struct {
	Logger    *logger.Logger
	Workers   int
	QueueSize int
}

type JobHandle struct {
	jobID string
}

func (h *JobHandle) JobID() string {
	if h == nil {
		return ""
	}
	return h.jobID
}

type Service struct {
	mu             sync.RWMutex
	schedMu        sync.Mutex
	bindMu         sync.Mutex
	scheduler      *runtime.Scheduler
	cancel         context.CancelFunc
	jobs           map[string]Registration
	runtimeService runtimecomp.Service
	runtimeStore   metrix.RuntimeStore
	runtimeMetrics *runtimeMetrics
	componentName  string
	producerName   string
	componentRegs  bool
	producerRegs   bool
	workers        int
	queueSize      int
	closed         bool
}

func NewExecutionService(cfg ExecutionServiceConfig) (*Service, error) {
	workers := cfg.Workers
	if workers <= 0 {
		workers = defaultWorkers
	}
	queueSize := cfg.QueueSize
	if queueSize <= 0 {
		queueSize = defaultQueueSize
	}

	scheduler, err := runtime.NewScheduler(runtime.SchedulerConfig{
		Logger:        cfg.Logger,
		Workers:       workers,
		QueueCapacity: queueSize,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	if err := scheduler.Start(ctx); err != nil {
		cancel()
		scheduler.Stop()
		return nil, err
	}

	return &Service{
		scheduler:    scheduler,
		cancel:       cancel,
		jobs:         make(map[string]Registration),
		runtimeStore: metrix.NewRuntimeStore(),
		workers:      workers,
		queueSize:    queueSize,
	}, nil
}

func (s *Service) Attach(reg Registration) (*JobHandle, error) {
	s.schedMu.Lock()
	defer s.schedMu.Unlock()

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, errExecutionServiceClosed
	}

	jobID, err := s.scheduler.RegisterJob(registrationToRuntime(reg))
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}

	s.jobs[jobID] = reg
	s.mu.Unlock()

	s.observeRuntimeSnapshotLocked()

	return &JobHandle{jobID: jobID}, nil
}

func (s *Service) Detach(handle *JobHandle) {
	if handle == nil {
		return
	}

	s.schedMu.Lock()
	defer s.schedMu.Unlock()

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}

	s.scheduler.UnregisterJob(handle.jobID)
	delete(s.jobs, handle.jobID)
	s.mu.Unlock()

	s.observeRuntimeSnapshotLocked()
}

func (s *Service) Snapshot() Snapshot {
	s.schedMu.Lock()
	defer s.schedMu.Unlock()

	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return Snapshot{}
	}
	s.mu.RUnlock()

	return s.snapshotLocked()
}

func (s *Service) BindRuntimeService(service runtimecomp.Service) error {
	if s == nil || service == nil {
		return nil
	}

	s.bindMu.Lock()
	defer s.bindMu.Unlock()

	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return errExecutionServiceClosed
	}
	if s.componentRegs {
		s.mu.RUnlock()
		return nil
	}
	store := s.runtimeStore
	s.mu.RUnlock()

	cfg := runtimecomp.ComponentConfig{
		Name:        executionRuntimeComponentName,
		Store:       store,
		UpdateEvery: 1,
		Autogen: runtimecomp.AutogenPolicy{
			Enabled: true,
		},
		Plugin:  "go.d",
		Module:  "nagios_execution",
		JobName: "service",
		JobLabels: map[string]string{
			"component": "nagios_execution_service",
		},
	}
	if err := service.RegisterComponent(cfg); err != nil {
		return err
	}
	producerName := cfg.Name + ".producer"
	if err := service.RegisterProducer(producerName, s.producerTick); err != nil {
		service.UnregisterComponent(cfg.Name)
		return err
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		service.UnregisterProducer(producerName)
		service.UnregisterComponent(cfg.Name)
		return errExecutionServiceClosed
	}
	if s.componentRegs {
		s.mu.Unlock()
		service.UnregisterProducer(producerName)
		service.UnregisterComponent(cfg.Name)
		return nil
	}
	s.runtimeService = service
	s.componentName = cfg.Name
	s.producerName = producerName
	s.componentRegs = true
	s.producerRegs = true
	if s.runtimeMetrics == nil {
		s.runtimeMetrics = newRuntimeMetrics(s.runtimeStore)
	}
	s.mu.Unlock()

	s.schedMu.Lock()
	s.observeRuntimeSnapshotLocked()
	s.schedMu.Unlock()
	return nil
}

func (s *Service) Close() {
	s.close()
}

func (s *Service) close() {
	s.schedMu.Lock()
	defer s.schedMu.Unlock()

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	cancel := s.cancel
	scheduler := s.scheduler
	runtimeService := s.runtimeService
	componentName := s.componentName
	producerName := s.producerName
	componentRegs := s.componentRegs
	producerRegs := s.producerRegs
	s.jobs = nil
	s.mu.Unlock()

	cancel()
	scheduler.Stop()
	if producerRegs && runtimeService != nil {
		runtimeService.UnregisterProducer(producerName)
	}
	if componentRegs && runtimeService != nil {
		runtimeService.UnregisterComponent(componentName)
	}
}

func (s *Service) workerCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.workers
}

func (s *Service) queueCapacity() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.queueSize
}

func (s *Service) attachedJobCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.jobs)
}

func (s *Service) observeSnapshot(snapshot Snapshot) {
	s.mu.RLock()
	metrics := s.runtimeMetrics
	s.mu.RUnlock()
	if metrics == nil {
		return
	}
	metrics.observeSnapshot(snapshot)
}

func (s *Service) producerTick() error {
	s.schedMu.Lock()
	defer s.schedMu.Unlock()

	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return nil
	}
	s.mu.RUnlock()

	s.observeRuntimeSnapshotLocked()
	return nil
}

func (s *Service) snapshotLocked() Snapshot {
	return snapshotFromRuntime(s.scheduler.Snapshot())
}

func (s *Service) observeRuntimeSnapshotLocked() {
	s.observeSnapshot(s.snapshotLocked())
}
