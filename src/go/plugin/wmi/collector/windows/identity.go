// SPDX-License-Identifier: GPL-3.0-or-later

package windows

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/netdata/netdata/go/plugins/plugin/wmi/collector/windows/internal/wmiquery"
)

const querySystemUUID = "SELECT UUID FROM Win32_ComputerSystemProduct"

// placeholderUUIDs stand in for a unique system UUID: all zeros, which Windows
// reports when the firmware has none, all F's, and a common firmware default,
// in Windows' byte order and in the order of tools that do not swap its first
// three fields.
var placeholderUUIDs = [...]uuid.UUID{
	uuid.Nil,
	uuid.Max,
	uuid.MustParse("03000200-0400-0500-0006-000700080009"),
	uuid.MustParse("00020003-0004-0005-0006-000700080009"),
}

// parseSystemUUID returns the host's SMBIOS system UUID in lowercase canonical
// form, or "" when the reply holds none that identifies the machine: the value
// is missing, is not a UUID, or is a placeholder.
func parseSystemUUID(row map[string]any) string {
	id, err := uuid.Parse(textProperty(row, "uuid"))
	if err != nil || slices.Contains(placeholderUUIDs[:], id) {
		return ""
	}
	return id.String()
}

// identify reads the host's system UUID as the session's first query. The
// startup check records it for the node's identity; a collection compares it
// and fails when the address reaches another machine, so that machine's samples
// never reach the node. A failed read fails the startup check, but only logs a
// warning in a collection, whose host the startup check identified. It reports
// whether the host answered.
func (c *Collector) identify(parent, ctx context.Context, session querySession) (reached bool, err error) {
	rows, err := session.Query(ctx, querySystemUUID)
	if err != nil {
		if ctx.Err() != nil {
			return false, c.interrupted(parent, ctx)
		}
		if wmiquery.SessionFailed(err) {
			// WinRM connects and logs in with the first query.
			return false, safeTransportError("failed to open WMI session", err)
		}
		return true, c.failedIdentification(safeTransportError("system UUID WMI query failed", err))
	}
	if len(rows) != 1 {
		return true, c.failedIdentification(publicError("system UUID WMI query did not return one instance"))
	}

	id := parseSystemUUID(rows[0])
	if c.vnode == nil {
		c.systemUUID = id
		return true, nil
	}
	if id != c.systemUUID {
		return true, publicError("host system UUID differs from the one read when the job started")
	}
	return true, nil
}

// failedIdentification fails the startup check, which has no identity for the
// node yet, and is a warning in a collection.
func (c *Collector) failedIdentification(err error) error {
	if c.vnode == nil {
		return err
	}
	c.Limit("wmi:system UUID", 1, time.Minute).Warning(err)
	return nil
}
