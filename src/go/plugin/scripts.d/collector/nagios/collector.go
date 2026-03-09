// SPDX-License-Identifier: GPL-3.0-or-later

package nagios

import (
	"context"
	_ "embed"
	"strings"

	"github.com/netdata/netdata/go/plugins/pkg/metrix"
	"github.com/netdata/netdata/go/plugins/plugin/framework/collectorapi"
	"github.com/netdata/netdata/go/plugins/plugin/framework/runtimecomp"
	"github.com/netdata/netdata/go/plugins/plugin/scripts.d/collector/nagios/internal/execution"
	"github.com/netdata/netdata/go/plugins/plugin/scripts.d/collector/nagios/internal/spec"
	"github.com/netdata/netdata/go/plugins/plugin/scripts.d/pkg/timeperiod"
)

//go:embed config_schema.json
var configSchema string

//go:embed charts.yaml
var nagiosChartTemplateV2 string

func init() {
	collectorapi.Register("nagios", collectorapi.Creator{
		JobConfigSchema: configSchema,
		Defaults: collectorapi.Defaults{
			UpdateEvery: 10,
		},
		CreateV2: func() collectorapi.CollectorV2 { return NewWithExecutionService(sharedExecutionService) },
		Config:   func() any { return &Config{} },
	})
}

var sharedExecutionService = mustSharedExecutionService()

func mustSharedExecutionService() execution.ExecutionService {
	svc, err := execution.NewExecutionService(execution.ExecutionServiceConfig{})
	if err != nil {
		panic(err)
	}
	return svc
}

// Config is the public v2 config surface.
type Config struct {
	spec.JobConfig `yaml:",inline" json:",inline"`
	TimePeriods    []timeperiod.Config `yaml:"time_periods,omitempty" json:"time_periods,omitempty"`
}

// Collector is the v2 Nagios collector.
type Collector struct {
	collectorapi.Base
	Config `yaml:",inline" json:",inline"`

	store   metrix.CollectorStore
	service execution.ExecutionService
	router  *perfdataRouter

	jobSpec   spec.JobSpec
	periods   *timeperiod.Set
	jobHandle *execution.JobHandle
}

func New() *Collector {
	return NewWithExecutionService(sharedExecutionService)
}

func NewWithExecutionService(service execution.ExecutionService) *Collector {
	if service == nil {
		panic("nagios: execution service must not be nil")
	}
	return &Collector{
		Config:  Config{},
		store:   metrix.NewCollectorStore(),
		service: service,
		router:  newPerfdataRouter(defaultPerfdataMetricKeyBudget),
	}
}

func (c *Collector) Configuration() any { return c.Config }

func (c *Collector) Init(ctx context.Context) error {
	if err := c.compileTimePeriods(); err != nil {
		return err
	}
	sp, err := c.JobConfig.ToSpec()
	if err != nil {
		return err
	}
	c.jobSpec = sp

	c.bindRuntimeService(ctx)

	handle, err := c.service.Attach(execution.Registration{
		Spec:    c.jobSpec,
		Periods: c.periods,
	})
	if err != nil {
		return err
	}
	c.jobHandle = handle
	return nil
}

func (c *Collector) Check(context.Context) error {
	_, err := c.JobConfig.ToSpec()
	return err
}

func (c *Collector) Collect(context.Context) error {
	if c.service == nil || c.jobHandle == nil {
		return nil
	}

	snapshot := c.service.Snapshot()
	sm := c.store.Write().SnapshotMeter("nagios")
	jobStateSet := sm.Vec("nagios_job").StateSet(
		"job.state",
		metrix.WithStateSetMode(metrix.ModeEnum),
		metrix.WithStateSetStates("ok", "warning", "critical", "unknown"),
		metrix.WithUnit("state"),
	)

	for _, job := range snapshot.Jobs {
		if job.JobID != c.jobHandle.JobID() {
			continue
		}

		jobLbl := sm.LabelSet(metrix.Label{Key: "nagios_job", Value: job.JobName})
		observeJob := func(name string, value float64) {
			sm.Gauge(name).Observe(value, jobLbl)
		}

		jobStateSet.WithLabelValues(job.JobName).Enable(normalizeJobStateForMetric(job.State))
		observeJob("job.attempt", float64(job.Attempt))
		observeJob("job.max_attempts", float64(job.MaxAttempt))

		perf := c.router.route(job.JobName, c.jobSpec.Plugin, job.PerfSamples)
		for _, measureSet := range buildPerfMeasureSets(perf) {
			sm.MeasureSetGauge(
				measureSet.name,
				metrix.WithMeasureSetFields(perfMeasureSetFieldSpecs()...),
				metrix.WithChartFamily(measureSet.scriptName),
				metrix.WithUnit(measureSet.unit),
			).ObserveFields(measureSet.values, jobLbl)
		}
	}

	counterLbl := sm.LabelSet(metrix.Label{Key: "nagios_job", Value: c.jobSpec.Name})
	counters := c.router.dropCounters()
	sm.Counter("perfdata_dropped_invalid_total").ObserveTotal(float64(counters.Invalid), counterLbl)
	sm.Counter("perfdata_dropped_collision_total").ObserveTotal(float64(counters.Collision), counterLbl)
	sm.Counter("perfdata_dropped_unit_drift_total").ObserveTotal(float64(counters.UnitDrift), counterLbl)
	sm.Counter("perfdata_dropped_budget_total").ObserveTotal(float64(counters.Budget), counterLbl)

	return nil
}

func (c *Collector) Cleanup(context.Context) {
	if c.service != nil && c.jobHandle != nil {
		c.service.Detach(c.jobHandle)
		c.jobHandle = nil
	}
}

func (c *Collector) MetricStore() metrix.CollectorStore { return c.store }

func (c *Collector) ChartTemplateYAML() string { return nagiosChartTemplateV2 }

func normalizeJobStateForMetric(state string) string {
	switch strings.ToUpper(strings.TrimSpace(state)) {
	case "OK":
		return "ok"
	case "WARNING":
		return "warning"
	case "CRITICAL":
		return "critical"
	default:
		return "unknown"
	}
}

func boolToFloat(v bool) float64 {
	if v {
		return 1
	}
	return 0
}

func (c *Collector) compileTimePeriods() error {
	cfgs := timeperiod.EnsureDefault(append([]timeperiod.Config(nil), c.TimePeriods...))
	set, err := timeperiod.Compile(cfgs)
	if err != nil {
		return err
	}
	c.periods = set
	return nil
}

func (c *Collector) bindRuntimeService(ctx context.Context) {
	if ctx == nil || c.service == nil {
		return
	}
	runtimeService, ok := runtimecomp.ServiceFromContext(ctx)
	if !ok {
		return
	}
	if err := c.service.BindRuntimeService(runtimeService); err != nil {
		c.Warningf("execution runtime telemetry registration failed: %v", err)
	}
}
