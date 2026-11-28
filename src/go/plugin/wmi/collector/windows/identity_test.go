// SPDX-License-Identifier: GPL-3.0-or-later

package windows

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netdata/netdata/go/plugins/pkg/confopt"
	"github.com/netdata/netdata/go/plugins/pkg/metrix"
	"github.com/netdata/netdata/go/plugins/plugin/framework/dyncfg"
	"github.com/netdata/netdata/go/plugins/plugin/framework/vnodes"
	"github.com/netdata/netdata/go/plugins/plugin/go.d/pkg/collecttest"
)

// Node GUIDs are the RFC 9562 name-based (SHA-1) UUIDs, in the DNS namespace,
// of "wmi:" plus the lowercase system UUID, or plus the canonical endpoint for
// a host without a usable UUID. The values come from Python's uuid.uuid5, not
// from this package.
const (
	// fixtureNodeGUID is the node of fixtureSystemUUID.
	fixtureNodeGUID = "732346d5-907c-58b9-af58-7cabd4cbfc54"
	// otherSystemUUID is another machine's UUID, and otherNodeGUID its node.
	otherSystemUUID = "3F2504E0-4F89-11D3-9A0C-0305E82C3301"
	otherNodeGUID   = "12a59bc4-25a5-5b7f-b20e-bd5158f57a90"
	// endpointNodeGUID is the node of the fixture host windows.example.com
	// when it reports no usable UUID: the GUID that jobs got from their
	// endpoint alone before the system UUID was read.
	endpointNodeGUID = "9057e2bf-0a3a-5279-bf53-ae1bea9c7abd"
	// addressNodeGUID is the node of 192.0.2.14 without a usable UUID.
	addressNodeGUID = "7c380ba9-6613-5830-b578-9fd607ec1f6b"
)

func TestParseSystemUUID(t *testing.T) {
	const canonical = "4c4c4544-0042-3510-8051-b7c04f4e4d32"
	tests := map[string]struct {
		row  map[string]any
		want string
	}{
		"as Windows reports it": {row: map[string]any{"uuid": fixtureSystemUUID}, want: canonical},
		"lowercase":             {row: map[string]any{"uuid": canonical}, want: canonical},
		"braces and spaces": {
			row:  map[string]any{"uuid": " {" + fixtureSystemUUID + "} "},
			want: canonical,
		},
		"missing, as WinRM omits it empty": {row: map[string]any{}, want: ""},
		"empty":                            {row: map[string]any{"uuid": ""}, want: ""},
		"not text":                         {row: map[string]any{"uuid": uint64(1)}, want: ""},
		"not a UUID":                       {row: map[string]any{"uuid": "To Be Filled By O.E.M."}, want: ""},
		"all zeros": {
			row:  map[string]any{"uuid": "00000000-0000-0000-0000-000000000000"},
			want: "",
		},
		"all F's": {
			row:  map[string]any{"uuid": "FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF"},
			want: "",
		},
		"firmware default": {
			row:  map[string]any{"uuid": "03000200-0400-0500-0006-000700080009"},
			want: "",
		},
		"firmware default, unswapped": {
			row:  map[string]any{"uuid": "00020003-0004-0005-0006-000700080009"},
			want: "",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, parseSystemUUID(test.row))
		})
	}
}

func TestNodeGUID(t *testing.T) {
	tests := map[string]struct {
		systemUUID string
		host       string
		want       string
	}{
		"system UUID": {
			systemUUID: "4c4c4544-0042-3510-8051-b7c04f4e4d32",
			host:       "192.0.2.14",
			want:       fixtureNodeGUID,
		},
		"no UUID: DNS name":                 {host: "windows.example.com", want: endpointNodeGUID},
		"no UUID: another spelling":         {host: "Windows.Example.com.", want: endpointNodeGUID},
		"no UUID: IP address":               {host: "192.0.2.14", want: addressNodeGUID},
		"no UUID: IPv4-mapped IPv6 address": {host: "::ffff:192.0.2.14", want: addressNodeGUID},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, nodeGUID(test.systemUUID, test.host))
		})
	}
}

// The startup check fixes the node's identity from the host's reply, over
// either transport and from any address of the host.
func TestCollector_JobV2NodeIdentity(t *testing.T) {
	tests := map[string]struct {
		host      string
		transport confopt.Enum[transportSpec]
		reply     queryResult
		wantGUID  string
	}{
		"DCOM, DNS name": {
			host:     "windows.example.com",
			reply:    systemUUIDResult(),
			wantGUID: fixtureNodeGUID,
		},
		"WinRM, IP address of the same machine": {
			host:      "192.0.2.14",
			transport: transportWinRM,
			reply:     systemUUIDResult(),
			wantGUID:  fixtureNodeGUID,
		},
		"another machine": {
			host:     "windows.example.com",
			reply:    fixtureRows(map[string]any{"UUID": otherSystemUUID}),
			wantGUID: otherNodeGUID,
		},
		"firmware default": {
			host:     "windows.example.com",
			reply:    fixtureRows(map[string]any{"UUID": "03000200-0400-0500-0006-000700080009"}),
			wantGUID: endpointNodeGUID,
		},
		"DCOM, empty UUID": {
			host:     "windows.example.com",
			reply:    fixtureRows(map[string]any{"UUID": ""}),
			wantGUID: endpointNodeGUID,
		},
		"WinRM, empty UUID omitted": {
			host:      "192.0.2.14",
			transport: transportWinRM,
			reply:     fixtureRows(map[string]any{"Name": "Standard PC"}),
			wantGUID:  addressNodeGUID,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				c := configuredCollector()
				c.Host = test.host
				c.Transport = test.transport
				c.newSession = func(context.Context) (querySession, error) {
					results := timedQueryResults()
					results[querySystemUUID] = test.reply
					return &fakeSession{
						results: results,
					}, nil
				}
				collect := startJob(t, c, vnodes.VirtualNode{})

				output := collect(c.UpdateEvery)
				assert.Contains(t, output, "HOST_DEFINE '"+test.wantGUID+"'")
				assert.Contains(t, output, "HOST_LABEL 'address' '"+test.host+"'")
				assert.Equal(t, test.wantGUID, c.VirtualNode().GUID)
			})
		})
	}
}

// A collection that finds another machine at the address publishes nothing on
// the node and queries nothing more; the original machine's return resumes it.
func TestCollector_JobV2IdentityChange(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := configuredCollector()
		reply := systemUUIDResult()
		var session *fakeSession
		c.newSession = func(context.Context) (querySession, error) {
			results := timedQueryResults()
			results[querySystemUUID] = reply
			session = &fakeSession{
				results: results,
			}
			return session, nil
		}
		collect := startJob(t, c, vnodes.VirtualNode{})

		up := collect(c.UpdateEvery)
		assert.Contains(t, up, "HOST '"+fixtureNodeGUID+"'")

		reply = fixtureRows(map[string]any{"UUID": otherSystemUUID})
		other := collect(2 * c.UpdateEvery)
		assert.NotContains(t, other, fixtureNodeGUID)
		assert.NotContains(t, other, otherNodeGUID)
		assert.Contains(t, other, "SET 'failed' = 1")
		assert.Equal(t, []string{querySystemUUID}, session.queries)
		assert.Equal(t, fixtureNodeGUID, c.VirtualNode().GUID)

		reply = systemUUIDResult()
		back := collect(3 * c.UpdateEvery)
		assert.Contains(t, back, "HOST '"+fixtureNodeGUID+"'")
		assert.Contains(t, back, "SET 'used' = 6442450944")
		assert.Contains(t, back, "SET 'success' = 1")
	})
}

func TestCollector_IdentityChangeError(t *testing.T) {
	const changed = "host system UUID differs from the one read when the job started"
	tests := map[string]struct {
		atStart   queryResult
		later     queryResult
		wantError string
	}{
		"same UUID in another spelling": {
			atStart: systemUUIDResult(),
			later:   fixtureRows(map[string]any{"UUID": "{4c4c4544-0042-3510-8051-b7c04f4e4d32}"}),
		},
		"another UUID": {
			atStart:   systemUUIDResult(),
			later:     fixtureRows(map[string]any{"UUID": otherSystemUUID}),
			wantError: changed,
		},
		"UUID replaced by a placeholder": {
			atStart:   systemUUIDResult(),
			later:     fixtureRows(map[string]any{"UUID": "00000000-0000-0000-0000-000000000000"}),
			wantError: changed,
		},
		"UUID where the host reported none": {
			atStart:   fixtureRows(map[string]any{"UUID": ""}),
			later:     systemUUIDResult(),
			wantError: changed,
		},
		"still no UUID": {
			atStart: fixtureRows(map[string]any{"UUID": "FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF"}),
			later:   fixtureRows(map[string]any{"UUID": ""}),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			c := configuredCollector()
			reply := test.atStart
			c.newSession = func(context.Context) (querySession, error) {
				results := validQueryResults()
				results[querySystemUUID] = reply
				return &fakeSession{
					results: results,
				}, nil
			}
			require.NoError(t, c.Init(t.Context()))
			require.NoError(t, c.Check(t.Context()))
			startGUID := c.VirtualNode().GUID

			reply = test.later
			_, err := collecttest.CollectScalarSeries(c)
			if test.wantError == "" {
				require.NoError(t, err)
			} else {
				message, ok := dyncfg.PublicMessage(err)
				require.True(t, ok, "%v", err)
				assert.Equal(t, test.wantError, message)
			}
			assert.Equal(t, startGUID, c.VirtualNode().GUID)
		})
	}
}

// The runtime retries a failed startup check on the same collector; the node
// takes the identity of the check that succeeds, here after the address moved
// to another machine between the attempts.
func TestCollector_CheckRetryRereadsIdentity(t *testing.T) {
	c := configuredCollector()
	attempts := []map[string]queryResult{
		{querySystemUUID: systemUUIDResult()},
		withOtherSystemUUID(validQueryResults()),
	}
	c.newSession = func(context.Context) (querySession, error) {
		results := attempts[0]
		attempts = attempts[1:]
		return &fakeSession{
			results: results,
		}, nil
	}
	require.NoError(t, c.Init(t.Context()))

	require.ErrorContains(t, c.Check(t.Context()), "no enabled WMI query group succeeded")
	assert.Nil(t, c.VirtualNode())
	require.NoError(t, c.Check(t.Context()))
	assert.Equal(t, otherNodeGUID, c.VirtualNode().GUID)
}

func withOtherSystemUUID(results map[string]queryResult) map[string]queryResult {
	results[querySystemUUID] = fixtureRows(map[string]any{"UUID": otherSystemUUID})
	return results
}

// The identity query reached the host, so a session lost after it is not a
// failure to open the session: the poll reports its query failures.
func TestCollector_SessionLostAfterIdentity(t *testing.T) {
	lostAfterIdentity := func() map[string]queryResult {
		results := validQueryResults()
		for query := range results {
			if query != querySystemUUID {
				results[query] = queryResult{
					err: sessionFailure{errors.New("connection reset")},
				}
			}
		}
		return results
	}
	tests := map[string]struct {
		checkFirst bool
	}{
		"startup check":       {},
		"collection after it": {checkFirst: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			c := configuredCollector()
			c.Transport = transportWinRM
			results := validQueryResults()
			c.newSession = func(context.Context) (querySession, error) {
				return &fakeSession{
					results: results,
				}, nil
			}
			require.NoError(t, c.Init(t.Context()))
			run := c.Check
			if test.checkFirst {
				require.NoError(t, c.Check(t.Context()))
				run = c.Collect
			}

			results = lostAfterIdentity()
			err := run(t.Context())
			message, ok := dyncfg.PublicMessage(err)
			require.True(t, ok, "%v", err)
			assert.Contains(t, message, "no enabled WMI query group succeeded")
			assert.NotContains(t, message, "failed to open WMI session")
		})
	}
}

// A failed read fails the startup check, before any metric query, but not a
// collection: the startup check identified the host.
func TestCollector_IdentityReadFailure(t *testing.T) {
	failures := map[string]struct {
		reply     queryResult
		wantError string
	}{
		"query rejected": {
			reply: queryResult{
				err: errors.New("access denied for fixture-account"),
			},
			wantError: "system UUID WMI query failed",
		},
		"no instance": {
			reply:     fixtureRows(),
			wantError: "system UUID WMI query did not return one instance",
		},
		"two instances": {
			reply: fixtureRows(
				map[string]any{"UUID": fixtureSystemUUID},
				map[string]any{"UUID": otherSystemUUID},
			),
			wantError: "system UUID WMI query did not return one instance",
		},
	}
	for name, failure := range failures {
		t.Run("startup check: "+name, func(t *testing.T) {
			c := configuredCollector()
			session := &fakeSession{
				results: validQueryResults(),
			}
			session.results[querySystemUUID] = failure.reply
			c.newSession = func(context.Context) (querySession, error) { return session, nil }
			require.NoError(t, c.Init(t.Context()))

			err := c.Check(t.Context())
			message, ok := dyncfg.PublicMessage(err)
			require.True(t, ok, "%v", err)
			assert.Equal(t, failure.wantError, message)
			assert.NotContains(t, err.Error(), "fixture-")
			assert.Equal(t, []string{querySystemUUID}, session.queries)
			assert.Nil(t, c.VirtualNode())
		})
		t.Run("collection: "+name, func(t *testing.T) {
			c := configuredCollector()
			reply := systemUUIDResult()
			c.newSession = func(context.Context) (querySession, error) {
				results := validQueryResults()
				results[querySystemUUID] = reply
				return &fakeSession{
					results: results,
				}, nil
			}
			require.NoError(t, c.Init(t.Context()))
			require.NoError(t, c.Check(t.Context()))

			reply = failure.reply
			metrics, err := collecttest.CollectScalarSeries(c)
			require.NoError(t, err)
			assert.Equal(t, metrix.SampleValue(6*1024*1024*1024), metrics["memory_used_bytes"])
			assert.Equal(t, fixtureNodeGUID, c.VirtualNode().GUID)
		})
	}
}
