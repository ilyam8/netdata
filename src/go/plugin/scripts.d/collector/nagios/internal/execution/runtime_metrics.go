// SPDX-License-Identifier: GPL-3.0-or-later

package execution

import (
	"sync"

	"github.com/netdata/netdata/go/plugins/pkg/metrix"
)

const (
	executionRuntimeMetricPrefix  = "netdata.go.plugin.scriptsd.nagios.execution"
	executionRuntimeComponentName = "nagios.execution"
)

type runtimeMetrics struct {
	mu sync.Mutex

	running   metrix.StatefulGauge
	queued    metrix.StatefulGauge
	scheduled metrix.StatefulGauge

	started  metrix.StatefulCounter
	finished metrix.StatefulCounter
	skipped  metrix.StatefulCounter

	lastStarted  uint64
	lastFinished uint64
	lastSkipped  uint64
}

func newRuntimeMetrics(store metrix.RuntimeStore) *runtimeMetrics {
	if store == nil {
		return nil
	}

	meter := store.Write().StatefulMeter(executionRuntimeMetricPrefix)
	m := &runtimeMetrics{
		running: metrix.SeededGauge(meter,
			"running",
			metrix.WithDescription("Current number of running Nagios jobs in the shared execution service"),
			metrix.WithChartFamily("Scripts.d/Nagios/Execution"),
			metrix.WithUnit("jobs"),
		),
		queued: metrix.SeededGauge(meter,
			"queued",
			metrix.WithDescription("Current number of queued Nagios jobs in the shared execution service"),
			metrix.WithChartFamily("Scripts.d/Nagios/Execution"),
			metrix.WithUnit("jobs"),
		),
		scheduled: metrix.SeededGauge(meter,
			"scheduled",
			metrix.WithDescription("Current number of scheduled Nagios jobs in the shared execution service"),
			metrix.WithChartFamily("Scripts.d/Nagios/Execution"),
			metrix.WithUnit("jobs"),
		),
		started: metrix.SeededCounter(meter,
			"started_total",
			metrix.WithDescription("Total number of Nagios job executions admitted to the shared execution service"),
			metrix.WithChartFamily("Scripts.d/Nagios/Execution Totals"),
			metrix.WithUnit("jobs"),
		),
		finished: metrix.SeededCounter(meter,
			"finished_total",
			metrix.WithDescription("Total number of Nagios job executions finished by the shared execution service"),
			metrix.WithChartFamily("Scripts.d/Nagios/Execution Totals"),
			metrix.WithUnit("jobs"),
		),
		skipped: metrix.SeededCounter(meter,
			"skipped_total",
			metrix.WithDescription("Total number of Nagios job executions skipped by single-flight admission in the shared execution service"),
			metrix.WithChartFamily("Scripts.d/Nagios/Execution Totals"),
			metrix.WithUnit("jobs"),
		),
	}
	return m
}

func (m *runtimeMetrics) observeSnapshot(snapshot Snapshot) {
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.running.Set(float64(snapshot.Running))
	m.queued.Set(float64(snapshot.Queued))
	m.scheduled.Set(float64(snapshot.Scheduled))

	m.started.Add(float64(monotonicDelta(snapshot.Started, m.lastStarted)))
	m.finished.Add(float64(monotonicDelta(snapshot.Finished, m.lastFinished)))
	m.skipped.Add(float64(monotonicDelta(snapshot.Skipped, m.lastSkipped)))

	m.lastStarted = snapshot.Started
	m.lastFinished = snapshot.Finished
	m.lastSkipped = snapshot.Skipped
}

func monotonicDelta(cur, prev uint64) uint64 {
	if cur < prev {
		return cur
	}
	return cur - prev
}
