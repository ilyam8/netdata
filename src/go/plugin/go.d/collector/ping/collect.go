// SPDX-License-Identifier: GPL-3.0-or-later

package ping

import (
	"context"
	"sync"

	"github.com/netdata/netdata/go/plugins/plugin/go.d/pkg/pinger"
)

func (c *Collector) collect(ctx context.Context) error {
	mx := c.metrics

	for _, sample := range c.collectSamples(ctx, true) {
		host := sample.Host

		mx.packetsRecv.WithLabelValues(host).Observe(float64(sample.PacketsRecv))
		mx.packetsSent.WithLabelValues(host).Observe(float64(sample.PacketsSent))
		mx.packetLoss.WithLabelValues(host).Observe(sample.PacketLossPct * 1000)

		if sample.RTT.Valid {
			mx.minRTT.WithLabelValues(host).Observe(float64(sample.RTT.Min.Microseconds()))
			mx.maxRTT.WithLabelValues(host).Observe(float64(sample.RTT.Max.Microseconds()))
			mx.avgRTT.WithLabelValues(host).Observe(float64(sample.RTT.Avg.Microseconds()))
			mx.stdDevRTT.WithLabelValues(host).Observe(float64(sample.RTT.StdDev.Microseconds()))
			mx.rttVariance.WithLabelValues(host).Observe(float64(sample.RTT.VarianceMicrosecondsSquared()))
		}

		// A host without replies is not observed, so its distribution chart gets a gap.
		for _, rtt := range sample.RTTs {
			mx.rtt.WithLabelValues(host).Observe(durationToMilliseconds(rtt))
		}

		if sample.Jitter.InstantValid {
			mx.meanJitter.WithLabelValues(host).Observe(float64(sample.Jitter.Mean.Microseconds()))
		}
		if sample.Jitter.SmoothedValid {
			mx.ewmaJitter.WithLabelValues(host).Observe(float64(sample.Jitter.EWMA.Microseconds()))
			mx.smaJitter.WithLabelValues(host).Observe(float64(sample.Jitter.SMA.Microseconds()))
		}
	}

	return nil
}

func (c *Collector) collectSamples(ctx context.Context, track bool) []pinger.Sample {
	if c.client == nil {
		return nil
	}

	var (
		mu      sync.Mutex
		samples = make([]pinger.Sample, 0, len(c.Hosts))
		wg      sync.WaitGroup
	)

	for _, host := range c.Hosts {
		wg.Go(func() {
			sample, err := c.probeHost(ctx, host, track)
			if err != nil {
				c.Error(err)
				return
			}

			mu.Lock()
			samples = append(samples, sample)
			mu.Unlock()
		})
	}

	wg.Wait()

	return samples
}

func (c *Collector) probeHost(ctx context.Context, host string, track bool) (pinger.Sample, error) {
	if track {
		return c.client.ProbeAndTrack(ctx, host)
	}
	return c.client.Probe(ctx, host)
}
