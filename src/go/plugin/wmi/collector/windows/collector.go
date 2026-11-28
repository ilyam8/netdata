// SPDX-License-Identifier: GPL-3.0-or-later

// Package windows collects host metrics from a remote Windows host over WMI.
package windows

import (
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"time"

	"github.com/netdata/netdata/go/plugins/logger"
	"github.com/netdata/netdata/go/plugins/pkg/metrix"
	"github.com/netdata/netdata/go/plugins/pkg/tlscfg"
	"github.com/netdata/netdata/go/plugins/plugin/framework/collectorapi"
	"github.com/netdata/netdata/go/plugins/plugin/framework/dyncfg"
	"github.com/netdata/netdata/go/plugins/plugin/framework/vnodes"
	"github.com/netdata/netdata/go/plugins/plugin/go.d/pkg/pinger"
	"github.com/netdata/netdata/go/plugins/plugin/wmi/collector/windows/internal/wmiquery"
)

//go:embed config_schema.json
var configSchema string

//go:embed charts.yaml
var chartTemplateYAML string

func init() {
	collectorapi.Register("windows", collectorapi.Creator{
		Defaults: collectorapi.Defaults{
			UpdateEvery: defaultUpdateEvery,
		},
		JobConfigSchema: configSchema,
		CreateV2:        func() collectorapi.CollectorV2 { return New() },
		Config:          func() any { return &Config{} },
	})
}

// querySession is one collection cycle's WMI session, over either transport.
type querySession interface {
	Query(ctx context.Context, wql string) ([]map[string]any, error)
	Close()
}

type Collector struct {
	collectorapi.Base
	Config `yaml:",inline" json:""`

	store   metrix.CollectorStore
	metrics collectorMetrics
	// vnode is the generated node, created by a successful startup check from
	// systemUUID, the host's SMBIOS system UUID it read ("" when the host reports
	// none usable). Both stay unset for a job with a configured vnode.
	vnode         *vnodes.VirtualNode
	systemUUID    string
	nextInventory time.Time
	lastCPU       cpuSample
	// lastDisks holds each physical disk's sample of the previous poll, by
	// device, when that poll read the disks.
	lastDisks map[string]physicalDisk
	// diskDrives caches each numbered physical disk's Win32_DiskDrive entry,
	// by device; an empty entry stands for a disk Windows does not list there.
	diskDrives map[string]diskDrive

	winrmTLS *tls.Config
	// dcomAuth authenticates the DCOM sessions; a Kerberos one keeps the
	// job's service ticket between polls.
	dcomAuth   wmiquery.DCOMAuth
	newSession func(context.Context) (querySession, error)
	// now reads the poller's clock around each query, so the host's clock can
	// be compared with it.
	now func() time.Time

	// pingClient is nil once the poller turned out to lack the privilege to ping.
	pingClient pinger.Client
	newPinger  func(pinger.Config, *logger.Logger) (pinger.Client, error)
}

func New() *Collector {
	store := metrix.NewCollectorStore()
	c := &Collector{
		Config:    defaultConfig(),
		store:     store,
		metrics:   newCollectorMetrics(store),
		now:       time.Now,
		newPinger: pinger.New,
	}
	c.newSession = c.openSession
	return c
}

func (c *Collector) openSession(ctx context.Context) (querySession, error) {
	if c.Transport == transportWinRM {
		return wmiquery.OpenWinRM(c.Host, c.Username, c.Password, c.winrmTLS), nil
	}
	session, err := wmiquery.OpenDCOM(ctx, c.Host, c.dcomAuth)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (c *Collector) Configuration() any { return c.Config }

func (c *Collector) Init(ctx context.Context) error {
	if err := c.Config.validate(); err != nil {
		return err
	}
	switch c.Auth {
	case authKerberos:
		account, realm, _ := kerberosPrincipal(c.Username)
		kdcs, _ := kdcAddresses(c.KDC)
		c.dcomAuth = wmiquery.NewKerberosAuth(account, realm, c.Password, kdcs)
	default:
		c.dcomAuth = wmiquery.NewNTLMAuth(c.Username, c.Password)
	}
	if c.Transport == transportWinRM {
		cfg, err := tlscfg.NewTLSConfig(ctx, tlscfg.TLSConfig{
			TLSCA: c.TLSCA,
		})
		if err != nil {
			// The loader's message names the file, so only the wrapping text carries it.
			return fmt.Errorf("%w: %v", dyncfg.NewPublicError("could not load tls_ca", err), err)
		}
		c.winrmTLS = cfg
	}
	return c.initPinger()
}

// Check polls the enabled metric groups, without the inventory, and reads the
// host's system UUID, which fixes the generated node's identity for the life of
// the job.
func (c *Collector) Check(ctx context.Context) error {
	if _, err := c.poll(ctx, false); err != nil {
		return err
	}
	if c.Vnode != "" {
		return nil
	}
	if c.systemUUID == "" {
		// Logged as public text, which the Agent keeps for jobs with secret references.
		c.Warning(publicError("host reports no usable system UUID: its node is identified by the configured host"))
	}
	c.vnode = newVirtualNode(c.systemUUID, c.Host, nodeStaleAfter(c.UpdateEvery, c.Timeout.Duration()))
	return nil
}

// Collect pings the host alongside the WMI poll. A failed poll fails the
// collection, so the job never publishes ping alone.
func (c *Collector) Collect(ctx context.Context) error {
	// Only a generated node carries the host's inventory.
	inventoryDue := c.vnode != nil && !time.Now().Before(c.nextInventory)
	waitPing := c.startPing(ctx)
	snapshot, err := c.poll(ctx, inventoryDue)
	ping, pinged := waitPing(err != nil)
	if err != nil {
		return err
	}
	c.metrics.write(snapshot)
	if pinged {
		c.metrics.writePing(c.Host, ping)
	}
	return nil
}

// Cleanup forgets the job's Kerberos ticket. Every poll closes its own session.
func (c *Collector) Cleanup(context.Context) {
	if c.dcomAuth != nil {
		c.dcomAuth.Close()
	}
}

func (c *Collector) VirtualNode() *vnodes.VirtualNode { return c.vnode }

func (c *Collector) MetricStore() metrix.CollectorStore { return c.store }

func (c *Collector) ChartTemplateYAML() string { return chartTemplateYAML }
