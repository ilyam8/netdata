// SPDX-License-Identifier: GPL-3.0-or-later

package ping

import (
	"time"

	"github.com/netdata/netdata/go/plugins/pkg/metrix"
)

// rttBucketBoundsMs are the round-trip time histogram bucket upper bounds in
// milliseconds. They name the ping.host_rtt_distribution dimensions, so
// changing them changes a public chart.
var rttBucketBoundsMs = []float64{0.1, 0.2, 0.5, 1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000}

// hostMetrics holds the per-host instruments, labeled by "host".
type hostMetrics struct {
	minRTT      metrix.SnapshotGaugeVec
	maxRTT      metrix.SnapshotGaugeVec
	avgRTT      metrix.SnapshotGaugeVec
	stdDevRTT   metrix.SnapshotGaugeVec
	rttVariance metrix.SnapshotGaugeVec
	meanJitter  metrix.SnapshotGaugeVec
	ewmaJitter  metrix.SnapshotGaugeVec
	smaJitter   metrix.SnapshotGaugeVec
	packetLoss  metrix.SnapshotGaugeVec
	packetsRecv metrix.SnapshotGaugeVec
	packetsSent metrix.SnapshotGaugeVec

	// rtt is a cumulative count of echo replies by round-trip time in milliseconds.
	rtt metrix.StatefulHistogramVec
}

func newHostMetrics(store metrix.CollectorStore) hostMetrics {
	snapshot := store.Write().SnapshotMeter("").Vec("host")
	stateful := store.Write().StatefulMeter("").Vec("host")

	return hostMetrics{
		minRTT:      snapshot.Gauge("min_rtt"),
		maxRTT:      snapshot.Gauge("max_rtt"),
		avgRTT:      snapshot.Gauge("avg_rtt"),
		stdDevRTT:   snapshot.Gauge("std_dev_rtt"),
		rttVariance: snapshot.Gauge("rtt_variance"),
		meanJitter:  snapshot.Gauge("mean_jitter"),
		ewmaJitter:  snapshot.Gauge("ewma_jitter"),
		smaJitter:   snapshot.Gauge("sma_jitter"),
		packetLoss:  snapshot.Gauge("packet_loss"),
		packetsRecv: snapshot.Gauge("packets_recv"),
		packetsSent: snapshot.Gauge("packets_sent"),
		rtt:         stateful.Histogram("rtt", metrix.WithHistogramBounds(rttBucketBoundsMs...)),
	}
}

func durationToMilliseconds(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}
