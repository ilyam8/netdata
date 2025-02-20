// SPDX-License-Identifier: GPL-3.0-or-later

package snmpsd

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/sourcegraph/conc/pool"

	"github.com/netdata/netdata/go/plugins/logger"
	"github.com/netdata/netdata/go/plugins/plugin/go.d/agent/discovery/sd/model"
	"github.com/netdata/netdata/go/plugins/plugin/go.d/agent/filepersister"
	"github.com/netdata/netdata/go/plugins/plugin/go.d/pkg/iprange"
)

func NewDiscoverer(cfg Config) (*Discoverer, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	d := &Discoverer{
		Logger: logger.New().With(
			slog.String("component", "service discovery"),
			slog.String("discoverer", "snmp"),
		),
		rescanInterval: time.Minute * 1,
		timeout:        time.Second * 1,
		forgetDeadline: time.Hour * 6,
		parallelScans:  40,

		firstDiscovery: true,

		currStatus: newDiscoveryStatus(),
		lastStatus: newDiscoveryStatus(),
	}

	if cfg.RescanInterval > 0 {
		d.rescanInterval = cfg.RescanInterval
	}
	if cfg.Timeout > 0 {
		d.timeout = cfg.Timeout
	}
	if cfg.ParallelScans > 0 {
		d.parallelScans = cfg.ParallelScans
	}

	for _, n := range cfg.Networks {
		cred, ok := cfg.getCredential(n.Credential)
		if !ok {
			return nil, fmt.Errorf("subnet '%s' has no credential", n.Subnet)
		}

		r, err := iprange.ParseRange(n.Subnet)
		if err != nil {
			return nil, fmt.Errorf("invalid subnet range '%s': %v", n.Subnet, err)
		}

		d.subnets = append(d.subnets, subnet{
			str:        n.Subnet,
			ips:        r,
			credential: cred,
		})
	}

	return d, nil
}

type (
	Discoverer struct {
		*logger.Logger
		model.Base

		subnets []subnet

		stateFile string

		firstDiscovery bool
		currStatus     *discoveryStatus
		lastStatus     *discoveryStatus

		parallelScans  int
		rescanInterval time.Duration
		timeout        time.Duration
		forgetDeadline time.Duration
	}
	subnet struct {
		str        string
		ips        iprange.Range
		credential configSnmpCredentials
	}
)

func (d *Discoverer) Discover(ctx context.Context, in chan<- []model.TargetGroup) {
	d.Info("instance is started")
	d.Infof("rescan=%s, timeout=%s, forget=%s, parallel=%v", d.rescanInterval, d.timeout, d.forgetDeadline, d.parallelScans)
	defer func() { d.Info("instance is stopped") }()

	d.loadFileStatus()

	d.discoverNetworks(ctx, in)

	if d.rescanInterval <= 0 {
		filepersister.Save(statusFileName(), d.lastStatus)
		return
	}

	tk := time.NewTicker(d.rescanInterval)
	defer tk.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			d.discoverNetworks(ctx, in)
		}
	}
}

func (d *Discoverer) discoverNetworks(ctx context.Context, in chan<- []model.TargetGroup) {
	defer func() {
		d.lastStatus.replace(d.currStatus)

		if d.firstDiscovery && statusFileName() != "" {
			p := filepersister.New(statusFileName())
			p.FlushEvery = time.Second * 1
			go p.Run(ctx, d.lastStatus)
		}

		d.firstDiscovery = false
		d.lastStatus.setUpdated()
	}()

	d.currStatus.replace(d.lastStatus)

	now := time.Now()

	doProbing := !d.firstDiscovery || now.After(d.currStatus.LastDiscoveryTime.Add(d.rescanInterval))

	d.Infof("discovery mode: %s", map[bool]string{true: "active probing", false: "using cache"}[doProbing])

	var wg sync.WaitGroup

	for _, sub := range d.subnets {
		sub := sub
		wg.Add(1)
		go func() {
			defer wg.Done()
			d.discoverNetwork(ctx, in, sub, doProbing)
		}()
	}

	wg.Wait()

	if doProbing {
		d.currStatus.LastDiscoveryTime = now
	}
}

func (d *Discoverer) discoverNetwork(ctx context.Context, in chan<- []model.TargetGroup, sub subnet, doProbing bool) {
	tgg := newTargetGroup(sub)
	p := pool.New().WithMaxGoroutines(d.parallelScans)

	for ip := range sub.ips.Iterate() {
		ipAddr := ip.String()

		if doProbing {
			p.Go(func() { d.probeIPAddress(ctx, sub, ipAddr, tgg) })
		} else {
			d.useCacheIPAddress(sub, ipAddr, tgg)
		}
	}
	p.Wait()

	send(ctx, in, tgg)
}

func (d *Discoverer) useCacheIPAddress(sub subnet, ip string, tgg *targetGroup) {
	dev := d.currStatus.get(sub.str, ip)
	if dev == nil || !dev.IsSnmp {
		return
	}
	tg := newTarget(ip, sub.credential, dev.SysInfo)
	tgg.addTarget(tg)
}

func (d *Discoverer) probeIPAddress(ctx context.Context, sub subnet, ip string, tgg *targetGroup) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	now := time.Now()

	dev := d.currStatus.get(sub.str, ip)

	const (
		resultCache = iota
		resultErr
		resultOk
	)

	res, si := func() (int, SysInfo) {
		if dev != nil && dev.IsSnmp && now.After(dev.DiscoverTime.Add(d.forgetDeadline)) {
			return resultCache, dev.SysInfo
		}

		client := gosnmp.NewHandler()
		defer client.Close()

		client.SetTarget(ip)
		client.SetTimeout(d.timeout)
		client.SetRetries(0)
		setCredential(client, sub.credential)

		if err := client.Connect(); err != nil {
			d.Debugf("failed to connect to '%s': %v", ip, err)
			return resultErr, SysInfo{}
		}

		si, err := GetSysInfo(client)
		if err != nil {
			d.Debugf("failed to get SysInfo from '%s': %v", ip, err)
			return resultErr, SysInfo{}
		}

		d.Infof("discovered SNMP device at '%s'", ip)
		return resultOk, *si
	}()

	addTarget := true

	switch res {
	case resultCache:
	case resultErr:
		addTarget = false
		if dev != nil && now.Sub(dev.DiscoverTime) > d.forgetDeadline {
			d.currStatus.del(sub.str, ip)
		}
	case resultOk:
		d.currStatus.put(sub.str, ip, &discoveredDevice{
			IsSnmp:       true,
			DiscoverTime: now,
			SysInfo:      si,
		})
	default:
		return
	}

	if addTarget {
		tg := newTarget(ip, sub.credential, si)
		tgg.addTarget(tg)
	}
}

func send(ctx context.Context, in chan<- []model.TargetGroup, tgg model.TargetGroup) {
	if tgg == nil {
		return
	}

	for _, tg := range tgg.Targets() {
		logger.Infof("QQQ send snmp target '%s'", tg.TUID())
	}

	select {
	case <-ctx.Done():
	default:
		select {
		case <-ctx.Done():
		case in <- []model.TargetGroup{tgg}:
		}
	}
}
