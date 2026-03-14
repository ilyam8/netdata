// SPDX-License-Identifier: GPL-3.0-or-later

package jobmgr

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/netdata/netdata/go/plugins/pkg/netdataapi"
	"github.com/netdata/netdata/go/plugins/pkg/safewriter"
	"github.com/netdata/netdata/go/plugins/plugin/agent/secrets/secretstore"
	"github.com/netdata/netdata/go/plugins/plugin/framework/confgroup"
	"github.com/netdata/netdata/go/plugins/plugin/framework/dyncfg"
	"github.com/netdata/netdata/go/plugins/plugin/framework/functions"
)

func TestDyncfgSecretStoreSeqExec(t *testing.T) {
	tests := map[string]struct {
		run func(t *testing.T, mgr *Manager, out *bytes.Buffer)
	}{
		"add and get": {
			run: func(t *testing.T, mgr *Manager, out *bytes.Buffer) {
				addFn := dyncfg.NewFunction(functions.Function{
					UID:         "ss-add",
					ContentType: "application/json",
					Payload:     mustJSON(t, testVaultConfig()),
					Args: []string{
						mgr.dyncfgSecretStoreTemplateID(secretstore.KindVault),
						string(dyncfg.CommandAdd),
						"vault_prod",
					},
				})
				mgr.dyncfgSecretStoreSeqExec(addFn)

				var addResp map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-add", &addResp)
				assert.Equal(t, float64(202), addResp["status"])

				getFn := dyncfg.NewFunction(functions.Function{
					UID:  "ss-get",
					Args: []string{mgr.dyncfgSecretStoreID(secretstore.StoreKey(secretstore.KindVault, "vault_prod")), string(dyncfg.CommandGet)},
				})
				mgr.dyncfgSecretStoreSeqExec(getFn)

				var cfg map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-get", &cfg)
				assert.Equal(t, "vault_prod", cfg["name"])
				assert.Equal(t, string(secretstore.KindVault), cfg["kind"])
			},
		},
		"duplicate add is rejected": {
			run: func(t *testing.T, mgr *Manager, out *bytes.Buffer) {
				seedSecretStore(t, mgr, secretstore.KindVault, "vault_prod", testVaultConfig(), dyncfg.StatusRunning)

				addFn := dyncfg.NewFunction(functions.Function{
					UID:         "ss-add-duplicate",
					ContentType: "application/json",
					Payload:     mustJSON(t, testVaultConfigTokenFile()),
					Args: []string{
						mgr.dyncfgSecretStoreTemplateID(secretstore.KindVault),
						string(dyncfg.CommandAdd),
						"vault_prod",
					},
				})
				mgr.dyncfgSecretStoreSeqExec(addFn)

				var addResp map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-add-duplicate", &addResp)
				assert.Equal(t, float64(409), addResp["status"])
				assert.Contains(t, addResp["errorMessage"], "already exists")

				getFn := dyncfg.NewFunction(functions.Function{
					UID:  "ss-get-after-duplicate",
					Args: []string{mgr.dyncfgSecretStoreID(secretstore.StoreKey(secretstore.KindVault, "vault_prod")), string(dyncfg.CommandGet)},
				})
				mgr.dyncfgSecretStoreSeqExec(getFn)

				var cfg map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-get-after-duplicate", &cfg)
				assert.Equal(t, "token", cfg["mode"])

				_, ok := mgr.secretStoreSvc.GetStatus(secretstore.StoreKey(secretstore.KindVault, "vault_prod"))
				assert.True(t, ok)
			},
		},
		"runtime-affecting update succeeds for running store": {
			run: func(t *testing.T, mgr *Manager, out *bytes.Buffer) {
				seedSecretStore(t, mgr, secretstore.KindVault, "vault_prod", testVaultConfig(), dyncfg.StatusRunning)

				cfg := prepareDyncfgCfg("success", "mysql")
				mgr.exposed.Add(&dyncfg.Entry[confgroup.Config]{
					Cfg:    cfg,
					Status: dyncfg.StatusRunning,
				})
				mgr.secretStoreDeps.SetActiveJobStores(cfg.FullName(), "success:mysql", []string{secretstore.StoreKey(secretstore.KindVault, "vault_prod")})
				mgr.secretStoreDeps.setRunning(cfg.FullName(), true)

				updateFn := dyncfg.NewFunction(functions.Function{
					UID:         "ss-update",
					ContentType: "application/json",
					Payload:     mustJSON(t, testVaultConfigTokenFile()),
					Args: []string{
						mgr.dyncfgSecretStoreID(secretstore.StoreKey(secretstore.KindVault, "vault_prod")),
						string(dyncfg.CommandUpdate),
					},
				})
				mgr.dyncfgSecretStoreSeqExec(updateFn)

				var resp map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-update", &resp)
				assert.Equal(t, float64(200), resp["status"])
				assert.Equal(t, "", resp["message"])

				getFn := dyncfg.NewFunction(functions.Function{
					UID:  "ss-get-updated",
					Args: []string{mgr.dyncfgSecretStoreID(secretstore.StoreKey(secretstore.KindVault, "vault_prod")), string(dyncfg.CommandGet)},
				})
				mgr.dyncfgSecretStoreSeqExec(getFn)

				var got map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-get-updated", &got)
				assert.Equal(t, "token_file", got["mode"])
			},
		},
		"unknown-field update is ignored for running store": {
			run: func(t *testing.T, mgr *Manager, out *bytes.Buffer) {
				seedSecretStore(t, mgr, secretstore.KindVault, "vault_prod", testVaultConfig(), dyncfg.StatusRunning)

				cfg := prepareDyncfgCfg("success", "mysql")
				mgr.exposed.Add(&dyncfg.Entry[confgroup.Config]{
					Cfg:    cfg,
					Status: dyncfg.StatusRunning,
				})
				mgr.secretStoreDeps.SetActiveJobStores(cfg.FullName(), "success:mysql", []string{secretstore.StoreKey(secretstore.KindVault, "vault_prod")})
				mgr.secretStoreDeps.setRunning(cfg.FullName(), true)

				updateCfg := testVaultConfig()
				updateCfg["ui_note"] = "updated description"

				updateFn := dyncfg.NewFunction(functions.Function{
					UID:         "ss-update-metadata",
					ContentType: "application/json",
					Payload:     mustJSON(t, updateCfg),
					Args: []string{
						mgr.dyncfgSecretStoreID(secretstore.StoreKey(secretstore.KindVault, "vault_prod")),
						string(dyncfg.CommandUpdate),
					},
				})
				mgr.dyncfgSecretStoreSeqExec(updateFn)

				var resp map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-update-metadata", &resp)
				assert.Equal(t, float64(200), resp["status"])

				getFn := dyncfg.NewFunction(functions.Function{
					UID:  "ss-get-metadata",
					Args: []string{mgr.dyncfgSecretStoreID(secretstore.StoreKey(secretstore.KindVault, "vault_prod")), string(dyncfg.CommandGet)},
				})
				mgr.dyncfgSecretStoreSeqExec(getFn)

				var got map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-get-metadata", &got)
				_, ok := got["ui_note"]
				assert.False(t, ok)
			},
		},
		"test command reports affected jobs": {
			run: func(t *testing.T, mgr *Manager, out *bytes.Buffer) {
				seedSecretStore(t, mgr, secretstore.KindVault, "vault_prod", testVaultConfig(), dyncfg.StatusRunning)

				cfg := prepareDyncfgCfg("success", "mysql")
				mgr.exposed.Add(&dyncfg.Entry[confgroup.Config]{
					Cfg:    cfg,
					Status: dyncfg.StatusRunning,
				})
				mgr.secretStoreDeps.SetActiveJobStores(cfg.FullName(), "success:mysql", []string{secretstore.StoreKey(secretstore.KindVault, "vault_prod")})
				testFn := dyncfg.NewFunction(functions.Function{
					UID:         "ss-test-affected",
					ContentType: "application/json",
					Payload:     mustJSON(t, testVaultConfigTokenFile()),
					Args: []string{
						mgr.dyncfgSecretStoreID(secretstore.StoreKey(secretstore.KindVault, "vault_prod")),
						string(dyncfg.CommandTest),
					},
				})
				mgr.dyncfgSecretStoreSeqExec(testFn)

				var resp map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-test-affected", &resp)
				assert.Equal(t, float64(202), resp["status"])
				assert.Contains(t, resp["message"], "Updated configuration will affect")
				assert.Contains(t, resp["message"], "success:mysql")
			},
		},
		"test command reports no-op for unchanged payload": {
			run: func(t *testing.T, mgr *Manager, out *bytes.Buffer) {
				seedSecretStore(t, mgr, secretstore.KindVault, "vault_prod", testVaultConfig(), dyncfg.StatusRunning)

				cfg := prepareDyncfgCfg("success", "mysql")
				mgr.exposed.Add(&dyncfg.Entry[confgroup.Config]{
					Cfg:    cfg,
					Status: dyncfg.StatusRunning,
				})
				mgr.secretStoreDeps.SetActiveJobStores(cfg.FullName(), "success:mysql", []string{secretstore.StoreKey(secretstore.KindVault, "vault_prod")})

				testFn := dyncfg.NewFunction(functions.Function{
					UID:         "ss-test-noop",
					ContentType: "application/json",
					Payload:     mustJSON(t, testVaultConfig()),
					Args: []string{
						mgr.dyncfgSecretStoreID(secretstore.StoreKey(secretstore.KindVault, "vault_prod")),
						string(dyncfg.CommandTest),
					},
				})
				mgr.dyncfgSecretStoreSeqExec(testFn)

				var resp map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-test-noop", &resp)
				assert.Equal(t, float64(202), resp["status"])
				assert.Equal(t, "Submitted configuration does not change the active secretstore.", resp["message"])
			},
		},
		"test command does not mutate generation": {
			run: func(t *testing.T, mgr *Manager, out *bytes.Buffer) {
				seedSecretStore(t, mgr, secretstore.KindVault, "vault_prod", testVaultConfig(), dyncfg.StatusRunning)
				before := mgr.secretStoreSvc.Capture().Generation()

				testCfg := testVaultConfig()
				testCfg["mode_token"].(map[string]any)["extra"] = "ignored"

				testFn := dyncfg.NewFunction(functions.Function{
					UID:         "ss-test",
					ContentType: "application/json",
					Payload:     mustJSON(t, testCfg),
					Args: []string{
						mgr.dyncfgSecretStoreID(secretstore.StoreKey(secretstore.KindVault, "vault_prod")),
						string(dyncfg.CommandTest),
					},
				})
				mgr.dyncfgSecretStoreSeqExec(testFn)

				var resp map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-test", &resp)
				assert.Equal(t, float64(202), resp["status"])
				assert.Equal(t, before, mgr.secretStoreSvc.Capture().Generation())
			},
		},
		"test command with empty payload validates stored config": {
			run: func(t *testing.T, mgr *Manager, out *bytes.Buffer) {
				seedSecretStore(t, mgr, secretstore.KindVault, "vault_prod", testVaultConfig(), dyncfg.StatusRunning)
				before := mgr.secretStoreSvc.Capture().Generation()

				testFn := dyncfg.NewFunction(functions.Function{
					UID:  "ss-test-empty",
					Args: []string{mgr.dyncfgSecretStoreID(secretstore.StoreKey(secretstore.KindVault, "vault_prod")), string(dyncfg.CommandTest)},
				})
				mgr.dyncfgSecretStoreSeqExec(testFn)

				var resp map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-test-empty", &resp)
				assert.Equal(t, float64(202), resp["status"])
				assert.Equal(t, "Stored configuration is valid. No jobs are currently using this secretstore.", resp["message"])

				status, ok := mgr.secretStoreSvc.GetStatus(secretstore.StoreKey(secretstore.KindVault, "vault_prod"))
				require.True(t, ok)
				require.NotNil(t, status.LastValidation)
				assert.True(t, status.LastValidation.OK)
				assert.Equal(t, before, mgr.secretStoreSvc.Capture().Generation())
			},
		},
		"disable rejects unexpected payload": {
			run: func(t *testing.T, mgr *Manager, out *bytes.Buffer) {
				seedSecretStore(t, mgr, secretstore.KindVault, "vault_prod", testVaultConfig(), dyncfg.StatusAccepted)

				disableFn := dyncfg.NewFunction(functions.Function{
					UID:         "ss-disable-unexpected-payload",
					ContentType: "application/json",
					Payload:     mustJSON(t, testVaultConfig()),
					Args: []string{
						mgr.dyncfgSecretStoreID(secretstore.StoreKey(secretstore.KindVault, "vault_prod")),
						string(dyncfg.CommandDisable),
					},
				})
				mgr.dyncfgSecretStoreSeqExec(disableFn)

				var resp map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-disable-unexpected-payload", &resp)
				assert.Equal(t, float64(400), resp["status"])
				assert.Contains(t, resp["errorMessage"], "payload is not supported")
			},
		},
		"disable accepted store does not restart dependents": {
			run: func(t *testing.T, mgr *Manager, out *bytes.Buffer) {
				seedSecretStore(t, mgr, secretstore.KindVault, "vault_prod", testVaultConfig(), dyncfg.StatusAccepted)

				cfg := prepareDyncfgCfg("success", "mysql")
				mgr.exposed.Add(&dyncfg.Entry[confgroup.Config]{
					Cfg:    cfg,
					Status: dyncfg.StatusRunning,
				})
				mgr.secretStoreDeps.SetActiveJobStores(cfg.FullName(), "success:mysql", []string{secretstore.StoreKey(secretstore.KindVault, "vault_prod")})
				mgr.secretStoreDeps.setRunning(cfg.FullName(), true)

				disableFn := dyncfg.NewFunction(functions.Function{
					UID:  "ss-disable-no-restart",
					Args: []string{mgr.dyncfgSecretStoreID(secretstore.StoreKey(secretstore.KindVault, "vault_prod")), string(dyncfg.CommandDisable)},
				})
				mgr.dyncfgSecretStoreSeqExec(disableFn)

				var resp map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-disable-no-restart", &resp)
				assert.Equal(t, float64(200), resp["status"])
				assert.NotContains(t, out.String(), "CONFIG test:collector:success:mysql status running")
			},
		},
		"remove deletes store": {
			run: func(t *testing.T, mgr *Manager, out *bytes.Buffer) {
				seedSecretStore(t, mgr, secretstore.KindVault, "vault_prod", testVaultConfig(), dyncfg.StatusAccepted)

				removeFn := dyncfg.NewFunction(functions.Function{
					UID:  "ss-remove",
					Args: []string{mgr.dyncfgSecretStoreID(secretstore.StoreKey(secretstore.KindVault, "vault_prod")), string(dyncfg.CommandRemove)},
				})
				mgr.dyncfgSecretStoreSeqExec(removeFn)

				var resp map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-remove", &resp)
				assert.Equal(t, float64(200), resp["status"])
				_, ok := mgr.lookupSecretStoreEntry(secretstore.StoreKey(secretstore.KindVault, "vault_prod"))
				assert.False(t, ok)
			},
		},
		"remove accepted store does not restart dependents": {
			run: func(t *testing.T, mgr *Manager, out *bytes.Buffer) {
				seedSecretStore(t, mgr, secretstore.KindVault, "vault_prod", testVaultConfig(), dyncfg.StatusAccepted)

				cfg := prepareDyncfgCfg("success", "mysql")
				mgr.exposed.Add(&dyncfg.Entry[confgroup.Config]{
					Cfg:    cfg,
					Status: dyncfg.StatusRunning,
				})
				mgr.secretStoreDeps.SetActiveJobStores(cfg.FullName(), "success:mysql", []string{secretstore.StoreKey(secretstore.KindVault, "vault_prod")})
				mgr.secretStoreDeps.setRunning(cfg.FullName(), true)

				removeFn := dyncfg.NewFunction(functions.Function{
					UID:  "ss-remove-no-restart",
					Args: []string{mgr.dyncfgSecretStoreID(secretstore.StoreKey(secretstore.KindVault, "vault_prod")), string(dyncfg.CommandRemove)},
				})
				mgr.dyncfgSecretStoreSeqExec(removeFn)

				var resp map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-remove-no-restart", &resp)
				assert.Equal(t, float64(200), resp["status"])
				assert.NotContains(t, out.String(), "CONFIG test:collector:success:mysql status running")
			},
		},
		"userconfig returns yaml from payload": {
			run: func(t *testing.T, mgr *Manager, out *bytes.Buffer) {
				userconfigFn := dyncfg.NewFunction(functions.Function{
					UID:         "ss-userconfig",
					ContentType: "application/json",
					Payload:     mustJSON(t, testVaultConfig()),
					Args: []string{
						mgr.dyncfgSecretStoreTemplateID(secretstore.KindVault),
						string(dyncfg.CommandUserconfig),
					},
				})
				mgr.dyncfgSecretStoreSeqExec(userconfigFn)

				re := regexp.MustCompile(`(?s)FUNCTION_RESULT_BEGIN ss-userconfig [^\n]+\n(.*?)\nFUNCTION_RESULT_END`)
				match := re.FindStringSubmatch(out.String())
				require.Len(t, match, 2)
				var parsed map[string]any
				require.NoError(t, yaml.Unmarshal([]byte(match[1]), &parsed))
				_, ok := parsed["name"]
				assert.False(t, ok)
				_, ok = parsed["kind"]
				assert.False(t, ok)
			},
		},
		"test rejects wrapped config payload": {
			run: func(t *testing.T, mgr *Manager, out *bytes.Buffer) {
				seedSecretStore(t, mgr, secretstore.KindVault, "vault_prod", testVaultConfig(), dyncfg.StatusAccepted)

				testFn := dyncfg.NewFunction(functions.Function{
					UID:         "ss-test-wrapped-config-payload",
					ContentType: "application/json",
					Payload: mustJSON(t, map[string]any{
						"config": testVaultConfigTokenFile(),
					}),
					Args: []string{
						mgr.dyncfgSecretStoreID(secretstore.StoreKey(secretstore.KindVault, "vault_prod")),
						string(dyncfg.CommandTest),
					},
				})
				mgr.dyncfgSecretStoreSeqExec(testFn)

				var resp map[string]any
				mustDecodeFunctionPayload(t, out.String(), "ss-test-wrapped-config-payload", &resp)
				assert.Equal(t, float64(400), resp["status"])
				assert.Contains(t, resp["errorMessage"], "mode")
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mgr, out := newDyncfgSecretStoreTestManager()
			tc.run(t, mgr, out)
		})
	}
}

func TestSecretStoreConfigFromPayload(t *testing.T) {
	tests := map[string]struct {
		fn              functions.Function
		name            string
		kind            secretstore.StoreKind
		wantErrContains string
		assertConfig    func(t *testing.T, cfg secretstore.Config)
	}{
		"json direct config payload": {
			fn: functions.Function{
				ContentType: "application/json",
				Payload:     mustJSON(t, testVaultConfig()),
			},
			name: "vault_prod",
			kind: secretstore.KindVault,
			assertConfig: func(t *testing.T, cfg secretstore.Config) {
				require.NotNil(t, cfg)
				assert.Equal(t, "vault_prod", cfg.Name())
				assert.Equal(t, secretstore.KindVault, cfg.Kind())
			},
		},
		"yaml direct config payload": {
			fn: functions.Function{
				Payload: []byte("mode: token\nmode_token:\n  token: vault-token\naddr: https://vault.example\n"),
			},
			name: "vault_prod",
			kind: secretstore.KindVault,
			assertConfig: func(t *testing.T, cfg secretstore.Config) {
				require.NotNil(t, cfg)
				assert.Equal(t, "vault_prod", cfg.Name())
				assert.Equal(t, secretstore.KindVault, cfg.Kind())
			},
		},
		"missing payload": {
			fn: functions.Function{
				ContentType: "application/json",
			},
			name:            "vault_prod",
			kind:            secretstore.KindVault,
			wantErrContains: "missing configuration payload",
		},
		"wrapped config payload becomes invalid raw config": {
			fn: functions.Function{
				ContentType: "application/json",
				Payload: mustJSON(t, map[string]any{
					"config": testVaultConfig(),
				}),
			},
			name: "vault_prod",
			kind: secretstore.KindVault,
			assertConfig: func(t *testing.T, cfg secretstore.Config) {
				require.NotNil(t, cfg)
				assert.Equal(t, "vault_prod", cfg.Name())
				assert.Equal(t, secretstore.KindVault, cfg.Kind())
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mgr, _ := newDyncfgSecretStoreTestManager()
			cfg, err := mgr.secretStoreConfigFromPayload(dyncfg.NewFunction(tc.fn), tc.name, tc.kind)
			if tc.wantErrContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrContains)
				return
			}

			require.NoError(t, err)
			if tc.assertConfig != nil {
				tc.assertConfig(t, cfg)
			}
		})
	}
}

func TestRememberSecretStoreConfig_InvalidDoesNotEnterCaches(t *testing.T) {
	mgr, _ := newDyncfgSecretStoreTestManager()

	raw := newSecretStoreConfigWithSource(t, secretstore.KindVault, "vault_prod", map[string]any{}, "/etc/netdata/secretstores.yaml", confgroup.TypeUser)

	entry, changed, err := mgr.rememberSecretStoreConfig(raw)
	require.Error(t, err)
	assert.False(t, changed)
	assert.Nil(t, entry)
	assert.Zero(t, mgr.secretStoreSeen.Count())
	assert.Zero(t, mgr.secretStoreExposed.Count())
}

func TestRememberSecretStoreConfig_CanonicalizesUnknownFields(t *testing.T) {
	mgr, _ := newDyncfgSecretStoreTestManager()

	cfg := testVaultConfig()
	cfg["ui_note"] = "ignored"
	cfg["mode_token"].(map[string]any)["extra"] = "ignored"

	raw := newSecretStoreConfigWithSource(t, secretstore.KindVault, "vault_prod", cfg, "/etc/netdata/secretstores.yaml", confgroup.TypeUser)
	entry, changed, err := mgr.rememberSecretStoreConfig(raw)
	require.NoError(t, err)
	require.True(t, changed)
	require.NotNil(t, entry)

	data := map[string]any{}
	require.NoError(t, json.Unmarshal(entry.Cfg.DataJSON(), &data))
	_, ok := data["ui_note"]
	assert.False(t, ok)

	modeToken := data["mode_token"].(map[string]any)
	_, ok = modeToken["extra"]
	assert.False(t, ok)
}

func TestRememberSecretStoreConfig_ReplayOrderingMatchesDyncfgAdd(t *testing.T) {
	key := secretstore.StoreKey(secretstore.KindVault, "vault_prod")

	liveMgr, _ := newDyncfgSecretStoreTestManager()
	replayMgr, _ := newDyncfgSecretStoreTestManager()

	addFn := dyncfg.NewFunction(functions.Function{
		UID:         "ss-replay-add",
		ContentType: "application/json",
		Payload:     mustJSON(t, testVaultConfig()),
		Args: []string{
			liveMgr.dyncfgSecretStoreTemplateID(secretstore.KindVault),
			string(dyncfg.CommandAdd),
			"vault_prod",
		},
	})
	liveMgr.dyncfgSecretStoreSeqExec(addFn)

	liveEntry, ok := liveMgr.lookupSecretStoreEntry(key)
	require.True(t, ok)
	assert.Equal(t, dyncfg.StatusAccepted, liveEntry.Status)
	_, ok = liveMgr.secretStoreSvc.GetStatus(key)
	assert.False(t, ok)

	raw := newSecretStoreConfigWithSource(t, secretstore.KindVault, "vault_prod", testVaultConfig(), confgroup.TypeDyncfg, confgroup.TypeDyncfg)
	replayEntry, changed, err := replayMgr.rememberSecretStoreConfig(raw)
	require.NoError(t, err)
	require.True(t, changed)
	require.NotNil(t, replayEntry)
	assert.Equal(t, dyncfg.StatusAccepted, replayEntry.Status)
	assert.JSONEq(t, string(liveEntry.Cfg.DataJSON()), string(replayEntry.Cfg.DataJSON()))
	_, ok = replayMgr.secretStoreSvc.GetStatus(key)
	assert.False(t, ok)

	liveMgr.dyncfgSecretStoreSeqExec(dyncfg.NewFunction(functions.Function{
		UID:  "ss-replay-enable-live",
		Args: []string{liveMgr.dyncfgSecretStoreID(key), string(dyncfg.CommandEnable)},
	}))
	replayMgr.dyncfgSecretStoreSeqExec(dyncfg.NewFunction(functions.Function{
		UID:  "ss-replay-enable-restored",
		Args: []string{replayMgr.dyncfgSecretStoreID(key), string(dyncfg.CommandEnable)},
	}))

	liveEntry, ok = liveMgr.lookupSecretStoreEntry(key)
	require.True(t, ok)
	assert.Equal(t, dyncfg.StatusRunning, liveEntry.Status)

	replayEntry, ok = replayMgr.lookupSecretStoreEntry(key)
	require.True(t, ok)
	assert.Equal(t, dyncfg.StatusRunning, replayEntry.Status)

	liveStatus, ok := liveMgr.secretStoreSvc.GetStatus(key)
	require.True(t, ok)
	replayStatus, ok := replayMgr.secretStoreSvc.GetStatus(key)
	require.True(t, ok)
	assert.Equal(t, liveStatus.Name, replayStatus.Name)
	assert.Equal(t, liveStatus.Kind, replayStatus.Kind)
}

func TestRemoveSecretStoreConfig_DoesNotRevealLowerPrioritySeenConfig(t *testing.T) {
	mgr, _ := newDyncfgSecretStoreTestManager()
	key := secretstore.StoreKey(secretstore.KindVault, "vault_prod")

	userCfg := newSecretStoreConfigWithSource(t, secretstore.KindVault, "vault_prod", testVaultConfig(), "/etc/netdata/secretstores.yaml", confgroup.TypeUser)
	entry, changed, err := mgr.rememberSecretStoreConfig(userCfg)
	require.NoError(t, err)
	require.True(t, changed)
	require.NotNil(t, entry)
	assert.Equal(t, userCfg.UID(), entry.Cfg.UID())

	dyncfgCfg := newSecretStoreConfigWithSource(t, secretstore.KindVault, "vault_prod", testVaultConfigTokenFile(), confgroup.TypeDyncfg, confgroup.TypeDyncfg)
	entry, changed, err = mgr.rememberSecretStoreConfig(dyncfgCfg)
	require.NoError(t, err)
	require.True(t, changed)
	require.NotNil(t, entry)
	assert.Equal(t, dyncfgCfg.UID(), entry.Cfg.UID())

	removed, ok := mgr.removeSecretStoreConfig(dyncfgCfg)
	require.True(t, ok)
	require.NotNil(t, removed)
	assert.Equal(t, dyncfgCfg.UID(), removed.Cfg.UID())

	_, ok = mgr.lookupSecretStoreEntry(key)
	assert.False(t, ok)

	seenUser, ok := mgr.secretStoreSeen.LookupByUID(userCfg.UID())
	require.True(t, ok)
	assert.Equal(t, userCfg.UID(), seenUser.UID())
	assert.Equal(t, 1, mgr.secretStoreSeen.Count())
	assert.Zero(t, mgr.secretStoreExposed.Count())
}

func TestRememberSecretStoreConfig_DrainsRestartFailureMessageFromNonHandlerStop(t *testing.T) {
	mgr, _ := newDyncfgSecretStoreTestManager()
	existing := newSecretStoreConfigWithSource(t, secretstore.KindVault, "vault_prod", testVaultConfig(), "/etc/netdata/secretstores.yaml", confgroup.TypeUser)
	require.NoError(t, mgr.secretStoreSvc.Add(existing))
	mgr.secretStoreSeen.Add(existing)
	mgr.secretStoreExposed.Add(&dyncfg.Entry[secretstore.Config]{
		Cfg:    existing,
		Status: dyncfg.StatusRunning,
	})

	cfg := prepareDyncfgCfg("fail", "mysql")
	mgr.exposed.Add(&dyncfg.Entry[confgroup.Config]{
		Cfg:    cfg,
		Status: dyncfg.StatusFailed,
	})
	mgr.secretStoreDeps.SetActiveJobStores(cfg.FullName(), "fail:mysql", []string{secretstore.StoreKey(secretstore.KindVault, "vault_prod")})

	replacement := newSecretStoreConfigWithSource(t, secretstore.KindVault, "vault_prod", testVaultConfigTokenFile(), confgroup.TypeDyncfg, confgroup.TypeDyncfg)
	_, changed, err := mgr.rememberSecretStoreConfig(replacement)
	require.NoError(t, err)
	require.True(t, changed)

	assert.Equal(t, "", mgr.secretStoreCb.TakeCommandMessage())
}

func newDyncfgSecretStoreTestManager() (*Manager, *bytes.Buffer) {
	var out bytes.Buffer

	mgr := New(Config{PluginName: testPluginName})
	mgr.ctx = context.Background()
	mgr.modules = prepareMockRegistry()
	mgr.fileStatus = newFileStatus()
	mgr.SetDyncfgResponder(dyncfg.NewResponder(netdataapi.New(safewriter.New(&out))))

	return mgr, &out
}

func testVaultConfig() map[string]any {
	return map[string]any{
		"mode": "token",
		"mode_token": map[string]any{
			"token": "vault-token",
		},
		"addr": "https://vault.example",
	}
}

func testVaultConfigTokenFile() map[string]any {
	return map[string]any{
		"mode": "token_file",
		"mode_token_file": map[string]any{
			"path": "/var/lib/netdata/vault.token",
		},
		"addr": "https://vault.example",
	}
}

func newSecretStoreFromConfig(t *testing.T, svc secretstore.Service, kind secretstore.StoreKind, name string, cfg map[string]any) secretstore.Config {
	t.Helper()
	_ = svc
	return newSecretStoreConfigWithSource(t, kind, name, cfg, confgroup.TypeDyncfg, confgroup.TypeDyncfg)
}

func newSecretStoreConfigWithSource(t *testing.T, kind secretstore.StoreKind, name string, cfg map[string]any, source, sourceType string) secretstore.Config {
	t.Helper()
	bs, err := json.Marshal(cfg)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(bs, &payload))
	out := secretstore.Config(payload)
	out.SetName(name)
	out.SetKind(kind)
	out.SetSource(source)
	out.SetSourceType(sourceType)
	return out
}

func seedSecretStore(t *testing.T, mgr *Manager, kind secretstore.StoreKind, name string, cfg map[string]any, status dyncfg.Status) secretstore.Config {
	t.Helper()

	raw := newSecretStoreFromConfig(t, mgr.secretStoreSvc, kind, name, cfg)
	if status == dyncfg.StatusRunning || status == dyncfg.StatusFailed {
		err := mgr.secretStoreSvc.Add(raw)
		require.NoError(t, err)
	}

	mgr.secretStoreSeen.Add(raw)
	mgr.secretStoreExposed.Add(&dyncfg.Entry[secretstore.Config]{
		Cfg:    raw,
		Status: status,
	})

	return raw
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	bs, err := json.Marshal(v)
	require.NoError(t, err)
	return bs
}

func mustDecodeFunctionPayload(t *testing.T, output, uid string, dst any) {
	t.Helper()

	re := regexp.MustCompile(`(?s)FUNCTION_RESULT_BEGIN ` + regexp.QuoteMeta(uid) + ` [^\n]+\n(.*?)\nFUNCTION_RESULT_END`)
	match := re.FindStringSubmatch(output)
	require.Len(t, match, 2, "function result for uid '%s' not found in output:\n%s", uid, output)
	require.NoError(t, json.Unmarshal([]byte(match[1]), dst))
}
