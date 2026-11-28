// SPDX-License-Identifier: GPL-3.0-or-later

package pinger

import (
	"testing"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/stretchr/testify/assert"
)

func TestDeriveSample(t *testing.T) {
	tests := map[string]struct {
		stats *probing.Statistics
		want  Sample
	}{
		"replies": {
			stats: testStats("host"),
			want: Sample{
				Host:          "host",
				PacketsSent:   5,
				PacketsRecv:   5,
				PacketLossPct: 0,
				RTTs: []time.Duration{
					10 * time.Millisecond,
					12 * time.Millisecond,
					15 * time.Millisecond,
					18 * time.Millisecond,
					20 * time.Millisecond,
				},
				RTT: RTTSummary{
					Valid:  true,
					Min:    10 * time.Millisecond,
					Max:    20 * time.Millisecond,
					Avg:    15 * time.Millisecond,
					StdDev: 5 * time.Millisecond,
				},
				Jitter: JitterSummary{
					InstantValid: true,
					Mean:         2500 * time.Microsecond,
				},
			},
		},
		"no replies": {
			stats: &probing.Statistics{
				PacketsSent: 5,
				PacketsRecv: 0,
				PacketLoss:  100,
			},
			want: Sample{
				Host:          "host",
				PacketsSent:   5,
				PacketsRecv:   0,
				PacketLossPct: 100,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := deriveSample("host", tc.stats, false, newStateStore(), AnalysisConfig{})
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestCalcMeanJitter(t *testing.T) {
	tests := map[string]struct {
		rtts []time.Duration
		want time.Duration
	}{
		"empty": {
			want: 0,
		},
		"single": {
			rtts: []time.Duration{10 * time.Millisecond},
			want: 0,
		},
		"two samples": {
			rtts: []time.Duration{10 * time.Millisecond, 15 * time.Millisecond},
			want: 5 * time.Millisecond,
		},
		"five samples": {
			rtts: []time.Duration{
				10 * time.Millisecond,
				12 * time.Millisecond,
				15 * time.Millisecond,
				18 * time.Millisecond,
				20 * time.Millisecond,
			},
			want: 2500 * time.Microsecond,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, calcMeanJitter(tc.rtts))
		})
	}
}

func TestStateStore_Update(t *testing.T) {
	state := newStateStore()
	cfg := AnalysisConfig{
		JitterEWMASamples: 16,
		JitterSMAWindow:   3,
	}

	ewma, sma := state.update("host", 2500*time.Microsecond, cfg)
	assert.Equal(t, time.Duration(156250), ewma)
	assert.Equal(t, 2500*time.Microsecond, sma)

	ewma, sma = state.update("host", 2500*time.Microsecond, cfg)
	assert.Equal(t, time.Duration(302734), ewma)
	assert.Equal(t, 2500*time.Microsecond, sma)

	_, sma = state.update("host", 4000*time.Microsecond, cfg)
	assert.Equal(t, time.Duration(3000000), sma)
}

func TestRTTSummaryVariance(t *testing.T) {
	rtt := RTTSummary{
		Valid:  true,
		StdDev: 5 * time.Millisecond,
	}

	assert.Equal(t, int64(25000000), rtt.VarianceMicrosecondsSquared())
	assert.InDelta(t, 25.0, rtt.VarianceMillisecondsSquared(), 1e-9)
}
