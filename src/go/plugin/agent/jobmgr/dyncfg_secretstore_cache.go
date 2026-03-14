// SPDX-License-Identifier: GPL-3.0-or-later

package jobmgr

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/netdata/netdata/go/plugins/plugin/agent/secrets/secretstore"
	"github.com/netdata/netdata/go/plugins/plugin/framework/dyncfg"
)

type secretStoreCallbacks struct {
	mgr            *Manager
	commandMessage string
}

func (cb *secretStoreCallbacks) ExtractKey(fn dyncfg.Function) (key, name string, ok bool) {
	if fn.Command() == dyncfg.CommandAdd {
		kind, kindOK := cb.mgr.extractSecretStoreKindFromTemplateID(fn.ID())
		name = fn.JobName()
		if !kindOK || name == "" {
			return "", "", false
		}
		return secretstore.StoreKey(kind, name), name, true
	}

	key, ok = cb.mgr.extractSecretStoreKey(fn.ID())
	if !ok {
		return "", "", false
	}
	_, name, err := secretstore.ParseStoreKey(key)
	if err != nil {
		return "", "", false
	}
	return key, name, true
}

func (cb *secretStoreCallbacks) ParseAndValidate(fn dyncfg.Function, name string) (secretstore.Config, error) {
	var kind secretstore.StoreKind
	if fn.Command() == dyncfg.CommandAdd {
		var ok bool
		kind, ok = cb.mgr.extractSecretStoreKindFromTemplateID(fn.ID())
		if !ok {
			return nil, fmt.Errorf("invalid template ID for secretstore add: %s", fn.ID())
		}
	} else {
		key, ok := cb.mgr.extractSecretStoreKey(fn.ID())
		if !ok {
			return nil, fmt.Errorf("invalid secretstore ID: %s", fn.ID())
		}
		var err error
		kind, name, err = secretstore.ParseStoreKey(key)
		if err != nil {
			return nil, err
		}
	}

	cfg, err := cb.mgr.secretStoreConfigFromPayload(fn, name, kind)
	if err != nil {
		return nil, err
	}
	return cb.mgr.normalizeSecretStoreConfig(cfg)
}

func (cb *secretStoreCallbacks) Start(cfg secretstore.Config) error {
	cb.commandMessage = ""
	key := cfg.ExposedKey()

	if _, ok := cb.mgr.secretStoreSvc.GetStatus(key); ok {
		if err := cb.mgr.secretStoreSvc.Update(key, cfg); err != nil {
			return &codedError{err: err, code: secretStoreErrorCode(err)}
		}
	} else if err := cb.mgr.secretStoreSvc.Add(cfg); err != nil {
		return &codedError{err: err, code: secretStoreErrorCode(err)}
	}

	cb.commandMessage = cb.mgr.restartSecretStoreDependentsMessage(key)
	return nil
}

func (cb *secretStoreCallbacks) Update(oldCfg, newCfg secretstore.Config) error {
	cb.commandMessage = ""
	key := oldCfg.ExposedKey()
	if _, ok := cb.mgr.secretStoreSvc.GetStatus(key); ok {
		if err := cb.mgr.secretStoreSvc.Update(key, newCfg); err != nil {
			return &codedError{err: err, code: secretStoreErrorCode(err)}
		}
	} else if err := cb.mgr.secretStoreSvc.Add(newCfg); err != nil {
		return &codedError{err: err, code: secretStoreErrorCode(err)}
	}

	cb.commandMessage = cb.mgr.restartSecretStoreDependentsMessage(key)
	return nil
}

func (cb *secretStoreCallbacks) Stop(cfg secretstore.Config) {
	cb.commandMessage = ""
	key := cfg.ExposedKey()
	if err := cb.mgr.secretStoreSvc.Remove(key); err != nil {
		if errors.Is(err, secretstore.ErrStoreNotFound) {
			return
		}
		return
	}
	cb.commandMessage = cb.mgr.restartSecretStoreDependentsMessage(key)
}

func (*secretStoreCallbacks) OnStatusChange(*dyncfg.Entry[secretstore.Config], dyncfg.Status, dyncfg.Function) {
}

func (cb *secretStoreCallbacks) TakeCommandMessage() string {
	msg := strings.TrimSpace(cb.commandMessage)
	cb.commandMessage = ""
	return msg
}

func (cb *secretStoreCallbacks) ConfigID(cfg secretstore.Config) string {
	return cb.mgr.dyncfgSecretStoreID(cfg.ExposedKey())
}

func (m *Manager) lookupSecretStoreEntry(key string) (*dyncfg.Entry[secretstore.Config], bool) {
	if strings.TrimSpace(key) == "" {
		return nil, false
	}
	if m.secretStoreExposed == nil {
		return nil, false
	}
	return m.secretStoreExposed.LookupByKey(key)
}

func (m *Manager) rememberSecretStoreConfig(cfg secretstore.Config) (*dyncfg.Entry[secretstore.Config], bool, error) {
	cfg, err := m.normalizeSecretStoreConfig(cfg)
	if err != nil {
		return nil, false, err
	}

	m.secretStoreHandler.RememberDiscoveredConfig(cfg)

	entry, ok := m.lookupSecretStoreEntry(cfg.ExposedKey())
	if !ok {
		entry = m.secretStoreHandler.AddDiscoveredConfig(cfg, dyncfg.StatusAccepted)
		m.secretStoreHandler.NotifyJobCreate(cfg, dyncfg.StatusAccepted)
		return entry, true, nil
	}

	sp, ep := cfg.SourceTypePriority(), entry.Cfg.SourceTypePriority()
	if ep > sp || (ep == sp && entry.Status == dyncfg.StatusRunning) {
		return entry, false, nil
	}

	if entry.Status == dyncfg.StatusRunning || entry.Status == dyncfg.StatusFailed {
		m.secretStoreCb.Stop(entry.Cfg)
		m.secretStoreCb.TakeCommandMessage()
	}

	entry = m.secretStoreHandler.AddDiscoveredConfig(cfg, dyncfg.StatusAccepted)
	m.secretStoreHandler.NotifyJobCreate(cfg, dyncfg.StatusAccepted)
	return entry, true, nil
}

func (m *Manager) removeSecretStoreConfig(cfg secretstore.Config) (*dyncfg.Entry[secretstore.Config], bool) {
	entry, ok := m.secretStoreHandler.RemoveDiscoveredConfig(cfg)
	if !ok {
		return nil, false
	}

	m.secretStoreCb.Stop(entry.Cfg)
	m.secretStoreCb.TakeCommandMessage()
	m.secretStoreHandler.NotifyJobRemove(entry.Cfg)
	return entry, true
}

func (m *Manager) validateSecretStoreConfig(cfg secretstore.Config) error {
	if m.secretStoreSvc == nil {
		return fmt.Errorf("secretstore service is not available")
	}
	return m.secretStoreSvc.Validate(cfg)
}

func (m *Manager) normalizeSecretStoreConfig(cfg secretstore.Config) (secretstore.Config, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if m.secretStoreSvc == nil {
		return nil, fmt.Errorf("secretstore service is not available")
	}
	return m.secretStoreSvc.Normalize(cfg)
}

func (m *Manager) validateSecretStoreStored(key string) error {
	entry, ok := m.lookupSecretStoreEntry(key)
	if !ok {
		return secretstore.ErrStoreNotFound
	}

	if _, ok := m.secretStoreSvc.GetStatus(key); ok {
		return m.secretStoreSvc.ValidateStored(key)
	}

	_, err := m.normalizeSecretStoreConfig(entry.Cfg)
	return err
}

func (m *Manager) secretStoreContext() context.Context {
	if m == nil || m.ctx == nil {
		return context.Background()
	}
	return m.ctx
}

func (m *Manager) secretStoreAffectedJobs(key string) []secretstore.JobRef {
	if m == nil || m.secretStoreDeps == nil {
		return nil
	}
	exposed, _ := m.secretStoreDeps.Impacted(key)
	return exposed
}

func (m *Manager) secretStoreAffectedJobsString(key string) string {
	refs := m.secretStoreAffectedJobs(key)
	if len(refs) == 0 {
		return ""
	}

	var b strings.Builder
	for i, ref := range refs {
		if i > 0 {
			b.WriteString(", ")
		}
		if ref.Display != "" {
			b.WriteString(ref.Display)
		} else {
			b.WriteString(ref.ID)
		}
	}
	return b.String()
}

type secretStoreRestartFailure struct {
	ref secretstore.JobRef
	err error
}

func (m *Manager) restartSecretStoreDependentsMessage(key string) string {
	failures := m.restartSecretStoreDependentsBestEffort(key)
	if len(failures) == 0 {
		return ""
	}

	parts := make([]string, 0, len(failures))
	for _, failure := range failures {
		name := failure.ref.Display
		if name == "" {
			name = failure.ref.ID
		}
		parts = append(parts, fmt.Sprintf("%s (%v)", name, failure.err))
	}

	return fmt.Sprintf("Dependent collector restart failures: %s.", strings.Join(parts, "; "))
}

func (m *Manager) restartSecretStoreDependentsBestEffort(key string) []secretStoreRestartFailure {
	if m == nil || m.secretStoreDeps == nil {
		return nil
	}

	exposed, _ := m.secretStoreDeps.Impacted(key)
	var failures []secretStoreRestartFailure
	for _, job := range exposed {
		entry, ok := m.lookupExposedByFullName(job.ID)
		if !ok {
			continue
		}
		switch entry.Status {
		case dyncfg.StatusRunning, dyncfg.StatusFailed:
		default:
			continue
		}
		if err := m.restartDependentCollectorJob(job.ID); err != nil {
			m.Warningf("dyncfg: secretstore: failed to restart dependent job '%s' after store '%s' change: %v", job.ID, key, err)
			failures = append(failures, secretStoreRestartFailure{ref: job, err: err})
		}
	}
	return failures
}
