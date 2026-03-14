// SPDX-License-Identifier: GPL-3.0-or-later

package secretstore_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/netdata/netdata/go/plugins/plugin/agent/secrets/secretstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePublished struct {
	blockOnCtx *atomic.Bool
}

func (p *fakePublished) Resolve(ctx context.Context, req secretstore.ResolveRequest) (string, error) {
	if p.blockOnCtx != nil && p.blockOnCtx.Load() {
		<-ctx.Done()
		return "", ctx.Err()
	}
	return req.Operand, nil
}

type fakeConfig struct {
	Auth map[string]any `json:"auth,omitempty"`
}

type fakeStore struct {
	cfg        fakeConfig
	failInit   *atomic.Bool
	blockOnCtx *atomic.Bool
	published  secretstore.PublishedStore
}

func (s *fakeStore) UnmarshalJSON(b []byte) error { return json.Unmarshal(b, &s.cfg) }
func (s *fakeStore) Configuration() any           { return &s.cfg }
func (s *fakeStore) Publish() secretstore.PublishedStore {
	return s.published
}

func (s *fakeStore) Init(context.Context) error {
	if s.failInit != nil && s.failInit.Load() {
		return errors.New("simulated validation error")
	}
	if len(s.cfg.Auth) == 0 {
		return errors.New("auth is required")
	}
	s.published = &fakePublished{blockOnCtx: s.blockOnCtx}
	return nil
}

func newFakeCreator(kind secretstore.StoreKind, failInit, blockOnCtx *atomic.Bool) secretstore.Creator {
	schema := map[string]any{
		"jsonSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"auth": map[string]any{"type": "object"},
			},
			"required": []any{"auth"},
		},
		"uiSchema": map[string]any{},
	}
	bs, err := json.Marshal(schema)
	if err != nil {
		panic(err)
	}

	return secretstore.Creator{
		Kind:        kind,
		DisplayName: "Fake Provider",
		Schema:      string(bs),
		Create: func() secretstore.Store {
			return &fakeStore{
				failInit:   failInit,
				blockOnCtx: blockOnCtx,
			}
		},
	}
}

func newFakeStore(_ *testing.T, _ secretstore.Service, kind secretstore.StoreKind, cfg fakeConfig, name string) secretstore.Config {
	bs, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(bs, &payload); err != nil {
		panic(err)
	}
	out := secretstore.Config(payload)
	out.SetName(name)
	out.SetKind(kind)
	out.SetSource("dyncfg")
	out.SetSourceType("dyncfg")
	return out
}

func TestServiceStatusLifecycle(t *testing.T) {
	var failInit atomic.Bool
	svc := secretstore.NewService(newFakeCreator(secretstore.KindVault, &failInit, nil))

	store := newFakeStore(t, svc, secretstore.KindVault, fakeConfig{
		Auth: map[string]any{"mode": "token_env"},
	}, "vault_prod")

	err := svc.Add(store)
	require.NoError(t, err)

	failInit.Store(true)
	storeKey := secretstore.StoreKey(secretstore.KindVault, "vault_prod")
	err = svc.ValidateStored(storeKey)
	require.Error(t, err)

	status, ok := svc.GetStatus(storeKey)
	require.True(t, ok)
	require.NotNil(t, status.LastValidation)
	assert.False(t, status.LastValidation.OK)
	assert.Equal(t, "simulated validation error", status.LastErrorSummary)

	failInit.Store(false)
	err = svc.ValidateStored(storeKey)
	require.NoError(t, err)

	status, ok = svc.GetStatus(storeKey)
	require.True(t, ok)
	require.NotNil(t, status.LastValidation)
	assert.True(t, status.LastValidation.OK)
	assert.Empty(t, status.LastErrorSummary)
}

func TestServiceResolveHonorsCanceledContext(t *testing.T) {
	var blockOnCtx atomic.Bool
	blockOnCtx.Store(true)

	svc := secretstore.NewService(newFakeCreator(secretstore.KindVault, nil, &blockOnCtx))
	store := newFakeStore(t, svc, secretstore.KindVault, fakeConfig{
		Auth: map[string]any{"mode": "token_env"},
	}, "vault_prod")
	err := svc.Add(store)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = svc.Resolve(ctx, svc.Capture(), "vault:vault_prod:secret", "${store:vault:vault_prod:secret}")
	require.ErrorIs(t, err, context.Canceled)
}

func TestServiceConcurrentResolveAndMutation(t *testing.T) {
	svc := secretstore.NewService(newFakeCreator(secretstore.KindVault, nil, nil))
	baseCfg := fakeConfig{
		Auth: map[string]any{"mode": "token_env"},
	}

	err := svc.Add(newFakeStore(t, svc, secretstore.KindVault, baseCfg, "vault_prod"))
	require.NoError(t, err)

	var wg sync.WaitGroup
	errCh := make(chan error, 32)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			snapshot := svc.Capture()
			val, err := svc.Resolve(context.Background(), snapshot, "vault:vault_prod:secret/data/app#key", "${store:vault:vault_prod:secret/data/app#key}")
			if err != nil {
				errCh <- err
				return
			}
			if val != "secret/data/app#key" {
				errCh <- errors.New("unexpected resolved value")
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			updateCfg := baseCfg
			if i%2 == 0 {
				updateCfg.Auth = map[string]any{
					"mode": "token_env",
					"tag":  "alt",
				}
			}
			if err := svc.Update(secretstore.StoreKey(secretstore.KindVault, "vault_prod"), newFakeStore(t, svc, secretstore.KindVault, updateCfg, "vault_prod")); err != nil {
				errCh <- err
				return
			}
		}
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		require.NoError(t, err)
	}
}
