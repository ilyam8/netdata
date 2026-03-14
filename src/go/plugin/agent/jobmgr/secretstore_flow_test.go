// SPDX-License-Identifier: GPL-3.0-or-later

package jobmgr

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/netdata/netdata/go/plugins/plugin/agent/secrets/resolver"
	"github.com/netdata/netdata/go/plugins/plugin/agent/secrets/secretstore"
	"github.com/netdata/netdata/go/plugins/plugin/framework/collectorapi"
	"github.com/netdata/netdata/go/plugins/plugin/framework/confgroup"
	"github.com/netdata/netdata/go/plugins/plugin/framework/dyncfg"
	"github.com/netdata/netdata/go/plugins/plugin/framework/functions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestStoreConfig struct {
	Value string `yaml:"value" json:"value"`
}

type testStore struct {
	TestStoreConfig `yaml:",inline" json:""`
}

func (s *testStore) Configuration() any { return &s.TestStoreConfig }

func (s *testStore) Init(context.Context) error {
	if s.TestStoreConfig.Value == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

func (s *testStore) Publish() secretstore.PublishedStore {
	return &testPublishedStore{value: s.TestStoreConfig.Value}
}

type testPublishedStore struct {
	value string
}

func (s *testPublishedStore) Resolve(_ context.Context, req secretstore.ResolveRequest) (string, error) {
	if req.Operand != "value" {
		return "", fmt.Errorf("unexpected operand %q", req.Operand)
	}
	return s.value, nil
}

type secretAwareCollector struct {
	collectorapi.Base
	Config collectorapi.MockConfiguration `yaml:",inline" json:""`
}

func (c *secretAwareCollector) Configuration() any           { return c.Config }
func (c *secretAwareCollector) Check(context.Context) error  { return nil }
func (c *secretAwareCollector) Cleanup(context.Context)      {}
func (c *secretAwareCollector) Charts() *collectorapi.Charts { return &collectorapi.Charts{} }
func (c *secretAwareCollector) Collect(context.Context) map[string]int64 {
	return map[string]int64{"value": 1}
}

func (c *secretAwareCollector) Init(context.Context) error {
	if c.Config.OptionStr != "good" {
		return fmt.Errorf("secret is not usable: %s", c.Config.OptionStr)
	}
	return nil
}

func newTestSecretStoreService() secretstore.Service {
	return secretstore.NewService(secretstore.Creator{
		Kind:        secretstore.KindVault,
		DisplayName: "Vault",
		Schema:      `{"jsonSchema":{"type":"object","properties":{"value":{"type":"string"}}},"uiSchema":[]}`,
		Create: func() secretstore.Store {
			return &testStore{}
		},
	})
}

func TestApplyConfig_ResolvesStoreReferenceWithKindAndName(t *testing.T) {
	svc := newTestSecretStoreService()
	raw := newSecretStoreConfigWithSource(t, secretstore.KindVault, "vault_prod", map[string]any{"value": "resolved-secret"}, confgroup.TypeDyncfg, confgroup.TypeDyncfg)
	require.NoError(t, svc.Add(raw))

	cfg := prepareDyncfgCfg("success", "secret-job").
		Set("option_str", "${store:vault:vault_prod:value}").
		Set("option_int", 7)

	module := &collectorapi.MockCollectorV1{}
	err := applyConfig(t.Context(), cfg, module, secretresolver.New(), svc, svc.Capture())
	require.NoError(t, err)

	assert.Equal(t, "resolved-secret", module.Config.OptionStr)
	assert.Equal(t, 7, module.Config.OptionInt)
}

func TestDyncfgSecretStoreUpdate_RestartsFailedDependentAfterStoreIsFixed(t *testing.T) {
	mgr, out := newDyncfgSecretStoreTestManager()
	mgr.secretStoreSvc = newTestSecretStoreService()
	mgr.modules["gated"] = collectorapi.Creator{
		Create: func() collectorapi.CollectorV1 { return &secretAwareCollector{} },
	}

	key := secretstore.StoreKey(secretstore.KindVault, "vault_prod")
	seedSecretStore(t, mgr, secretstore.KindVault, "vault_prod", map[string]any{"value": "good"}, dyncfg.StatusRunning)

	cfg := prepareDyncfgCfg("gated", "mysql").
		Set("option_str", "${store:vault:vault_prod:value}").
		Set("option_int", 1)
	mgr.exposed.Add(&dyncfg.Entry[confgroup.Config]{
		Cfg:    cfg,
		Status: dyncfg.StatusRunning,
	})
	mgr.syncSecretStoreDepsForConfig(cfg)
	require.NoError(t, mgr.collectorCb.Start(cfg))

	_, running := mgr.secretStoreDeps.Impacted(key)
	require.Len(t, running, 1)
	assert.Equal(t, cfg.FullName(), running[0].ID)

	badFn := dyncfg.NewFunction(functions.Function{
		UID:         "ss-update-bad",
		ContentType: "application/json",
		Payload:     mustJSON(t, map[string]any{"value": "bad"}),
		Args: []string{
			mgr.dyncfgSecretStoreID(key),
			string(dyncfg.CommandUpdate),
		},
	})
	mgr.dyncfgSecretStoreSeqExec(badFn)

	var badResp map[string]any
	mustDecodeFunctionPayload(t, out.String(), "ss-update-bad", &badResp)
	assert.Equal(t, float64(200), badResp["status"])
	assert.Contains(t, badResp["message"], "Dependent collector restart failures")
	assert.Contains(t, badResp["message"], "gated:mysql")

	entry, ok := mgr.lookupExposedByFullName(cfg.FullName())
	require.True(t, ok)
	assert.Equal(t, dyncfg.StatusFailed, entry.Status)

	_, running = mgr.secretStoreDeps.Impacted(key)
	assert.Empty(t, running)

	goodFn := dyncfg.NewFunction(functions.Function{
		UID:         "ss-update-good",
		ContentType: "application/json",
		Payload:     mustJSON(t, map[string]any{"value": "good"}),
		Args: []string{
			mgr.dyncfgSecretStoreID(key),
			string(dyncfg.CommandUpdate),
		},
	})
	mgr.dyncfgSecretStoreSeqExec(goodFn)

	var goodResp map[string]any
	mustDecodeFunctionPayload(t, out.String(), "ss-update-good", &goodResp)
	assert.Equal(t, float64(200), goodResp["status"])
	assert.Equal(t, "", goodResp["message"])

	entry, ok = mgr.lookupExposedByFullName(cfg.FullName())
	require.True(t, ok)
	assert.Equal(t, dyncfg.StatusRunning, entry.Status)

	_, running = mgr.secretStoreDeps.Impacted(key)
	require.Len(t, running, 1)
	assert.Equal(t, cfg.FullName(), running[0].ID)
}

func TestDyncfgSecretStoreGet_CanonicalJSONDoesNotExposeUnknownFields(t *testing.T) {
	mgr, out := newDyncfgSecretStoreTestManager()
	mgr.secretStoreSvc = newTestSecretStoreService()

	cfg := newSecretStoreConfigWithSource(t, secretstore.KindVault, "vault_prod", map[string]any{
		"value":   "resolved-secret",
		"ignored": "drop-me",
	}, confgroup.TypeDyncfg, confgroup.TypeDyncfg)
	_, changed, err := mgr.rememberSecretStoreConfig(cfg)
	require.NoError(t, err)
	require.True(t, changed)

	getFn := dyncfg.NewFunction(functions.Function{
		UID:  "ss-get-canonical",
		Args: []string{mgr.dyncfgSecretStoreID(secretstore.StoreKey(secretstore.KindVault, "vault_prod")), string(dyncfg.CommandGet)},
	})
	mgr.dyncfgSecretStoreSeqExec(getFn)

	var got map[string]any
	mustDecodeFunctionPayload(t, out.String(), "ss-get-canonical", &got)
	assert.Equal(t, "resolved-secret", got["value"])
	_, ok := got["ignored"]
	assert.False(t, ok)
}

func TestSecretStoreConfigFromPayload_PreservesKindAndNameForStoreSyntax(t *testing.T) {
	mgr, _ := newDyncfgSecretStoreTestManager()
	mgr.secretStoreSvc = newTestSecretStoreService()

	fn := dyncfg.NewFunction(functions.Function{
		ContentType: "application/json",
		Payload:     mustJSON(t, map[string]any{"value": "resolved-secret"}),
	})
	cfg, err := mgr.secretStoreConfigFromPayload(fn, "vault_prod", secretstore.KindVault)
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(cfg.DataJSON(), &payload))
	assert.Equal(t, "vault_prod", cfg.Name())
	assert.Equal(t, secretstore.KindVault, cfg.Kind())
	assert.Equal(t, "resolved-secret", payload["value"])
}
