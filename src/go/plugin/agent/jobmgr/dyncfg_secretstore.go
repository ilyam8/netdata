// SPDX-License-Identifier: GPL-3.0-or-later

package jobmgr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/netdata/netdata/go/plugins/pkg/netdataapi"
	"github.com/netdata/netdata/go/plugins/plugin/agent/secrets/secretstore"
	"github.com/netdata/netdata/go/plugins/plugin/framework/confgroup"
	"github.com/netdata/netdata/go/plugins/plugin/framework/dyncfg"
	"github.com/netdata/netdata/go/plugins/plugin/framework/functions"
	"gopkg.in/yaml.v2"
)

const (
	dyncfgSecretStorePrefixf = "%s:secretstore:"
	dyncfgSecretStorePath    = "/collectors/%s/SecretStores"
)

func (m *Manager) dyncfgSecretStorePrefixValue() string {
	return fmt.Sprintf(dyncfgSecretStorePrefixf, m.pluginName)
}

func (m *Manager) dyncfgSecretStoreTemplateID(kind secretstore.StoreKind) string {
	return fmt.Sprintf("%s%s", m.dyncfgSecretStorePrefixValue(), kind)
}

func (m *Manager) dyncfgSecretStoreID(id string) string {
	return fmt.Sprintf("%s%s", m.dyncfgSecretStorePrefixValue(), id)
}

func dyncfgSecretStoreTemplateCmds() string {
	return dyncfg.JoinCommands(dyncfg.CommandAdd, dyncfg.CommandSchema, dyncfg.CommandUserconfig)
}

func (m *Manager) dyncfgSecretStoreTemplatesCreate() {
	if m.secretStoreSvc == nil {
		return
	}
	for _, kind := range m.secretStoreSvc.Kinds() {
		m.dyncfgApi.ConfigCreate(netdataapi.ConfigOpts{
			ID:                m.dyncfgSecretStoreTemplateID(kind),
			Status:            dyncfg.StatusAccepted.String(),
			ConfigType:        dyncfg.ConfigTypeTemplate.String(),
			Path:              fmt.Sprintf(dyncfgSecretStorePath, m.pluginName),
			SourceType:        "internal",
			Source:            "internal",
			SupportedCommands: dyncfgSecretStoreTemplateCmds(),
		})
	}
}

func (m *Manager) dyncfgSecretStoreExec(fn dyncfg.Function) {
	if fn.Command() == dyncfg.CommandSchema {
		m.dyncfgSecretStoreSchema(fn)
		return
	}

	m.enqueueDyncfgFunction(fn)
}

func (m *Manager) dyncfgSecretStoreSeqExec(fn dyncfg.Function) {
	switch fn.Command() {
	case dyncfg.CommandSchema:
		m.dyncfgSecretStoreSchema(fn)
	case dyncfg.CommandGet:
		m.dyncfgSecretStoreGet(fn)
	case dyncfg.CommandAdd:
		if key, _, ok := m.secretStoreCb.ExtractKey(fn); ok {
			if _, exists := m.lookupSecretStoreEntry(key); exists {
				m.dyncfgApi.SendCodef(fn, 409, "The specified secretstore '%s' already exists.", key)
				return
			}
		}
		m.secretStoreHandler.CmdAdd(fn)
	case dyncfg.CommandUpdate:
		m.secretStoreHandler.CmdUpdate(fn)
	case dyncfg.CommandTest:
		m.dyncfgSecretStoreTest(fn)
	case dyncfg.CommandUserconfig:
		m.dyncfgSecretStoreUserconfig(fn)
	case dyncfg.CommandEnable:
		if err := validateSecretStoreNoPayload(fn); err != nil {
			m.dyncfgApi.SendCodef(fn, 400, "%v", err)
			return
		}
		m.secretStoreHandler.CmdEnable(fn)
	case dyncfg.CommandDisable:
		if err := validateSecretStoreNoPayload(fn); err != nil {
			m.dyncfgApi.SendCodef(fn, 400, "%v", err)
			return
		}
		m.secretStoreHandler.CmdDisable(fn)
	case dyncfg.CommandRemove:
		if err := validateSecretStoreNoPayload(fn); err != nil {
			m.dyncfgApi.SendCodef(fn, 400, "%v", err)
			return
		}
		m.secretStoreHandler.CmdRemove(fn)
	default:
		m.Warningf("dyncfg: function '%s' command '%s' not implemented", fn.Fn().Name, fn.Command())
		m.dyncfgApi.SendCodef(fn, 501, "Function '%s' command '%s' is not implemented.", fn.Fn().Name, fn.Command())
	}
}

func (m *Manager) dyncfgSecretStoreSchema(fn dyncfg.Function) {
	kind, ok := m.extractSecretStoreKindFromTemplateID(fn.ID())
	if !ok {
		storeKey, ok := m.extractSecretStoreKey(fn.ID())
		if !ok {
			m.dyncfgApi.SendCodef(fn, 400, "Invalid ID format for secretstore schema: %s.", fn.ID())
			return
		}
		entry, ok := m.lookupSecretStoreEntry(storeKey)
		if !ok {
			m.dyncfgApi.SendCodef(fn, 404, "The specified secretstore '%s' is not configured.", storeKey)
			return
		}
		kind = entry.Cfg.Kind()
	}

	schema, ok := m.secretStoreSvc.Schema(kind)
	if !ok {
		m.dyncfgApi.SendCodef(fn, 404, "The specified secretstore kind '%s' is not supported.", kind)
		return
	}

	m.dyncfgApi.SendJSON(fn, schema)
}

func (m *Manager) dyncfgSecretStoreGet(fn dyncfg.Function) {
	storeKey, ok := m.extractSecretStoreKey(fn.ID())
	if !ok {
		m.dyncfgApi.SendCodef(fn, 400, "Invalid ID format for secretstore get: %s.", fn.ID())
		return
	}

	entry, ok := m.lookupSecretStoreEntry(storeKey)
	if !ok {
		m.dyncfgApi.SendCodef(fn, 404, "The specified secretstore '%s' is not configured.", storeKey)
		return
	}

	m.dyncfgApi.SendJSON(fn, string(entry.Cfg.DataJSON()))
}

func (m *Manager) dyncfgSecretStoreTest(fn dyncfg.Function) {
	storeKey, ok := m.extractSecretStoreKey(fn.ID())
	if !ok {
		m.dyncfgApi.SendCodef(fn, 400, "Invalid ID format for secretstore test: %s.", fn.ID())
		return
	}

	if !fn.HasPayload() {
		if err := m.validateSecretStoreStored(storeKey); err != nil {
			m.dyncfgApi.SendCodef(fn, secretStoreErrorCode(err), "%v", err)
			return
		}
		m.sendSecretStoreTestImpactMessage(fn, storeKey, true)
		return
	}

	entry, ok := m.lookupSecretStoreEntry(storeKey)
	if !ok {
		m.dyncfgApi.SendCodef(fn, 404, "The specified secretstore '%s' is not configured.", storeKey)
		return
	}

	cfg, err := m.secretStoreConfigFromPayload(fn, entry.Cfg.Name(), entry.Cfg.Kind())
	if err != nil {
		m.dyncfgApi.SendCodef(fn, 400, "%v", err)
		return
	}

	normalized, err := m.normalizeSecretStoreConfig(cfg)
	if err != nil {
		m.dyncfgApi.SendCodef(fn, 400, "%v", err)
		return
	}

	if string(normalized.DataJSON()) == string(entry.Cfg.DataJSON()) {
		m.dyncfgApi.SendCodef(fn, 202, "Submitted configuration does not change the active secretstore.")
		return
	}

	m.sendSecretStoreTestImpactMessage(fn, storeKey, false)
}

func (m *Manager) dyncfgSecretStoreUserconfig(fn dyncfg.Function) {
	kind, ok := m.extractSecretStoreKindFromTemplateID(fn.ID())
	if !ok {
		m.dyncfgApi.SendCodef(fn, 400, "Invalid template ID for secretstore userconfig: %s.", fn.ID())
		return
	}
	if err := fn.ValidateHasPayload(); err != nil {
		m.dyncfgApi.SendCodef(fn, 400, "%v", err)
		return
	}

	store, ok := m.secretStoreSvc.New(kind)
	if !ok {
		m.dyncfgApi.SendCodef(fn, 404, "The specified secretstore kind '%s' is not supported.", kind)
		return
	}

	cfg := store.Configuration()
	if cfg == nil {
		m.dyncfgApi.SendCodef(fn, 500, "Secretstore kind '%s' does not provide configuration.", kind)
		return
	}
	if err := fn.UnmarshalPayload(cfg); err != nil {
		m.dyncfgApi.SendCodef(fn, 400, "Invalid configuration format. Failed to create configuration from payload: %v.", err)
		return
	}

	bs, err := yaml.Marshal(cfg)
	if err != nil {
		m.dyncfgApi.SendCodef(fn, 500, "Failed to convert configuration into YAML: %v.", err)
		return
	}

	m.dyncfgApi.SendYAML(fn, string(bs))
}

func (m *Manager) extractSecretStoreKindFromTemplateID(id string) (secretstore.StoreKind, bool) {
	rest, ok := strings.CutPrefix(id, m.dyncfgSecretStorePrefixValue())
	if !ok || rest == "" || strings.Contains(rest, ":") {
		return "", false
	}
	kind := secretstore.StoreKind(rest)
	if m.secretStoreSvc == nil {
		return "", false
	}
	_, ok = m.secretStoreSvc.DisplayName(kind)
	return kind, ok
}

func (m *Manager) extractSecretStoreKey(id string) (string, bool) {
	rest, ok := strings.CutPrefix(id, m.dyncfgSecretStorePrefixValue())
	if !ok || rest == "" {
		return "", false
	}
	kind, name, err := secretstore.ParseStoreKey(rest)
	if err != nil {
		return "", false
	}
	return secretstore.StoreKey(kind, name), true
}

func (m *Manager) secretStoreConfigFromPayload(fn dyncfg.Function, name string, kind secretstore.StoreKind) (secretstore.Config, error) {
	if err := fn.ValidateHasPayload(); err != nil {
		return nil, err
	}

	var payload secretstore.Config
	if err := fn.UnmarshalPayload(&payload); err != nil {
		return nil, fmt.Errorf("invalid configuration format: %w", err)
	}
	if payload == nil {
		payload = secretstore.Config{}
	}
	payload.SetName(name)
	payload.SetKind(kind)
	payload.SetSource(confgroup.TypeDyncfg)
	payload.SetSourceType(confgroup.TypeDyncfg)
	return payload, nil
}

func validateSecretStoreNoPayload(fn dyncfg.Function) error {
	if fn.HasPayload() {
		return fmt.Errorf("payload is not supported for this command")
	}
	return nil
}

func (m *Manager) sendSecretStoreTestImpactMessage(fn dyncfg.Function, storeKey string, validationOnly bool) {
	if validationOnly {
		if s := m.secretStoreAffectedJobsString(storeKey); s != "" {
			m.dyncfgApi.SendCodef(fn, 202, "Stored configuration is valid. Jobs currently using this secretstore: %s.", s)
			return
		}
		m.dyncfgApi.SendCodef(fn, 202, "Stored configuration is valid. No jobs are currently using this secretstore.")
		return
	}

	if s := m.secretStoreAffectedJobsString(storeKey); s != "" {
		m.dyncfgApi.SendCodef(fn, 202, "Updated configuration will affect: %s.", s)
		return
	}
	m.dyncfgApi.SendCodef(fn, 202, "No jobs will be affected by this change.")
}

func (m *Manager) restartDependentCollectorJob(fullName string) error {
	entry, ok := m.lookupExposedByFullName(fullName)
	if !ok {
		return fmt.Errorf("job '%s' is not exposed", fullName)
	}

	oldStatus := entry.Status
	switch oldStatus {
	case dyncfg.StatusRunning, dyncfg.StatusFailed:
	default:
		return fmt.Errorf("job '%s' restart is not allowed in '%s' state", fullName, oldStatus)
	}

	m.collectorCb.Stop(entry.Cfg)

	if err := m.collectorCb.Start(entry.Cfg); err != nil {
		entry.Status = dyncfg.StatusFailed
		m.handler.NotifyJobStatus(entry.Cfg, dyncfg.StatusFailed)
		m.collectorCb.OnStatusChange(entry, oldStatus, dyncfg.NewFunction(functions.Function{}))
		return fmt.Errorf("job '%s' restart failed: %w", fullName, err)
	}

	entry.Status = dyncfg.StatusRunning
	m.handler.NotifyJobStatus(entry.Cfg, dyncfg.StatusRunning)
	m.collectorCb.OnStatusChange(entry, oldStatus, dyncfg.NewFunction(functions.Function{}))
	return nil
}

func (m *Manager) lookupExposedByFullName(fullName string) (*dyncfg.Entry[confgroup.Config], bool) {
	if strings.TrimSpace(fullName) == "" {
		return nil, false
	}

	var found *dyncfg.Entry[confgroup.Config]
	m.exposed.ForEach(func(_ string, entry *dyncfg.Entry[confgroup.Config]) bool {
		if entry.Cfg.FullName() == fullName {
			found = entry
			return false
		}
		return true
	})

	return found, found != nil
}

func secretStoreErrorCode(err error) int {
	switch {
	case errors.Is(err, secretstore.ErrStoreExists):
		return 409
	case errors.Is(err, secretstore.ErrStoreNotFound):
		return 404
	default:
		return 400
	}
}
