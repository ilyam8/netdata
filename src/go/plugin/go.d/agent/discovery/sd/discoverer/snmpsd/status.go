// SPDX-License-Identifier: GPL-3.0-or-later

package snmpsd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func statusFileName() string {
	v := os.Getenv("NETDATA_LIB_DIR")
	if v == "" {
		return ""
	}
	return filepath.Join(v, "god-sd-snmp-status.json")
}

func (d *Discoverer) loadFileStatus() {
	filename := statusFileName()
	if filename == "" {
		return
	}

	s := newDiscoveryStatus()

	if f, err := os.Open(filename); err != nil {
		d.Warningf("failed to open status file %s: %v", filename, err)
	} else {
		defer func() { _ = f.Close() }()
		if err := json.NewDecoder(f).Decode(s); err != nil {
			d.Warningf("failed to parse status file %s: %v", filename, err)
		}
	}

	d.lastStatus = s

	d.Infof("loaded status file: last discovery=%s", d.lastStatus.LastDiscoveryTime)
}

func newDiscoveryStatus() *discoveryStatus {
	return &discoveryStatus{
		Networks: make(map[string]map[string]*discoveredDevice),
		ch:       make(chan struct{}, 1),
	}
}

type (
	discoveryStatus struct {
		mux               sync.RWMutex
		Networks          map[string]map[string]*discoveredDevice `json:"networks"`
		LastDiscoveryTime time.Time                               `json:"last_discovery_time"`
		ch                chan struct{}
	}
	discoveredDevice struct {
		DiscoverTime time.Time `json:"discover_time"`
		IsSnmp       bool      `json:"is_snmp"`
		SysInfo      SysInfo   `json:"sysinfo"`
	}
)

func (s *discoveryStatus) Bytes() ([]byte, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return json.MarshalIndent(s, "", " ")
}

func (s *discoveryStatus) Updated() <-chan struct{} {
	return s.ch
}

func (s *discoveryStatus) setUpdated() {
	select {
	case s.ch <- struct{}{}:
	default:
	}
}

func (s *discoveryStatus) get(subnet, ip string) *discoveredDevice {
	devices, ok := s.Networks[subnet]
	if !ok {
		return nil
	}
	return devices[ip]
}

func (s *discoveryStatus) put(subnet, ip string, dev *discoveredDevice) {
	devices, ok := s.Networks[subnet]
	if !ok {
		devices = make(map[string]*discoveredDevice)
		s.Networks[subnet] = devices
	}
	devices[ip] = dev
}

func (s *discoveryStatus) del(subnet, ip string) {
	devices, ok := s.Networks[subnet]
	if !ok {
		return
	}
	delete(devices, ip)
}

func (s *discoveryStatus) replace(st *discoveryStatus) {
	networks := make(map[string]map[string]*discoveredDevice)

	for sub, devices := range st.Networks {
		if networks[sub] == nil {
			networks[sub] = make(map[string]*discoveredDevice)
		}
		for ip, dev := range devices {
			networks[sub][ip] = &discoveredDevice{
				DiscoverTime: dev.DiscoverTime,
				IsSnmp:       dev.IsSnmp,
				SysInfo:      dev.SysInfo,
			}
		}
	}
	s.mux.Lock()
	s.Networks = networks
	s.LastDiscoveryTime = st.LastDiscoveryTime
	s.mux.Unlock()
}
