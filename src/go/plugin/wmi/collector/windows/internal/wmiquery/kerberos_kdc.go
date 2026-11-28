// SPDX-License-Identifier: GPL-3.0-or-later

package wmiquery

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"slices"
	"strconv"
	"time"

	"github.com/oiweiwei/gokrb5.fork/v9/iana/errorcode"
	"github.com/oiweiwei/gokrb5.fork/v9/iana/patype"
	"github.com/oiweiwei/gokrb5.fork/v9/messages"
	"github.com/oiweiwei/gokrb5.fork/v9/types"
)

// maxKDCReply bounds the length a KDC reply may declare. gokrb5 allocates the
// declared length before reading the reply, so a rogue KDC, such as one named
// by a spoofed SRV record, could make the plugin allocate up to 4 GiB. Windows
// limits a Kerberos token, ticket included, to MaxTokenSize (at most 65,535
// bytes), so legitimate replies stay far below this.
const maxKDCReply = 1 << 20

// maxS2KIterations bounds the PBKDF2 iteration count a KDC may name for
// deriving the account's AES key from its password. gokrb5 derives the key
// with the count the reply names, up to 2^32, synchronously and past any
// deadline (about 175 ns per iteration on a laptop: over 12 minutes at the
// maximum). Active Directory sends no count, which means the RFC 3962 default of
// 4096; this bound allows 256 times that, about 0.2 s.
const maxS2KIterations = 1 << 20

var (
	errKDCReplyTooLarge = errors.New("KDC reply exceeds the size limit")
	errKeyParameters    = errors.New("KDC reply proposes unacceptable key parameters")
)

type kdcPhase string

const (
	phaseLogin         kdcPhase = "Kerberos login"
	phaseServiceTicket kdcPhase = "Kerberos service ticket request"
)

// kdcExchange is the KDC dialer of one ticket request. It binds every KDC
// connection to the session's context and keeps the last KDC reply, because
// gokrb5 reports a KDC's error only as text that also carries the KDC's
// free-form message.
type kdcExchange struct {
	ctx context.Context
	// dialTimeout bounds each connection attempt, so a KDC that drops
	// connections leaves time to try the next one.
	dialTimeout time.Duration
	dial        func(ctx context.Context, network, address string) (net.Conn, error)

	// The state of the current send. gokrb5 sends each request to the KDCs in
	// turn until one delivers a reply, so a KDC that sends no data leaves the
	// state of an earlier one in the same send. answered reports that a KDC
	// sent data; reply is the last such KDC's message when it was complete and
	// within maxKDCReply, answeredBy its address, and badKeyParameters reports
	// a reply that checkKeyParameters rejected. delivered reports that gokrb5
	// got that reply, which ends the send.
	answered         bool
	reply            []byte
	answeredBy       string
	badKeyParameters bool
	delivered        bool
}

// Dial implements gokrb5's client.KDCDialer. gokrb5 dials once per KDC
// attempt and returns after the first KDC that delivers a reply.
func (e *kdcExchange) Dial(network, address string) (net.Conn, error) {
	if e.delivered {
		// A new request: the login's second try or the service ticket request.
		e.answered, e.reply, e.answeredBy, e.badKeyParameters, e.delivered = false, nil, "", false, false
	}
	ctx, cancel := context.WithTimeout(e.ctx, e.dialTimeout)
	conn, err := e.dial(ctx, network, address)
	cancel()
	if err != nil {
		return nil, err
	}
	stop := context.AfterFunc(e.ctx, func() { _ = conn.Close() })
	return &kdcConn{
		Conn:     conn,
		exchange: e,
		address:  address,
		stop:     stop,
	}, nil
}

// failure classifies a failed exchange with fixed text and the Kerberos
// error name. The KDC's free-form message and the principal names stay out.
func (e *kdcExchange) failure(phase kdcPhase) error {
	if err := e.ctx.Err(); err != nil {
		return fmt.Errorf("%s: %w", phase, err)
	}
	if e.badKeyParameters {
		return diagnostic(string(phase) +
			" failed: the KDC proposed key parameters the poller rejects (not AES, or excessive key derivation)")
	}
	if !e.answered {
		return diagnostic(string(phase) + " failed: could not reach a KDC")
	}
	if e.reply == nil {
		// Not a Kerberos reply: the connection broke, or a service other than a
		// KDC answered, or the reply exceeded maxKDCReply.
		return diagnostic(string(phase) + " failed: the KDC's reply was incomplete or too large")
	}
	var krbErr messages.KRBError
	if krbErr.Unmarshal(e.reply) == nil {
		return diagnostic(string(phase) + " failed: " + kdcErrorText(krbErr.ErrorCode))
	}
	if phase == phaseLogin {
		// A KDC that does not require preauthentication encrypts its reply
		// with the account's key; gokrb5 also checks the clocks here.
		return diagnostic(string(phase) +
			" failed: the KDC's reply could not be verified (wrong password, or the clocks differ by over five minutes)")
	}
	// After a valid reply, a referral to another domain fails without
	// contacting it: the poller does not follow referrals.
	return diagnostic(string(phase) +
		" failed: the KDC's reply could not be verified, or its referral to another domain could not be followed")
}

// kdcErrorText explains the Kerberos errors an operator can act on.
func kdcErrorText(code int32) string {
	switch code {
	case errorcode.KDC_ERR_C_PRINCIPAL_UNKNOWN:
		return "the account does not exist in the realm (KDC_ERR_C_PRINCIPAL_UNKNOWN)"
	case errorcode.KDC_ERR_S_PRINCIPAL_UNKNOWN:
		return "the realm has no service principal for the host name (KDC_ERR_S_PRINCIPAL_UNKNOWN)"
	case errorcode.KDC_ERR_PREAUTH_FAILED:
		return "wrong password (KDC_ERR_PREAUTH_FAILED)"
	case errorcode.KRB_AP_ERR_SKEW:
		return "the clocks of the poller and the KDC differ by more than the domain allows (KRB_AP_ERR_SKEW)"
	case errorcode.KDC_ERR_ETYPE_NOSUPP:
		return "no AES encryption type in common (KDC_ERR_ETYPE_NOSUPP)"
	case errorcode.KDC_ERR_CLIENT_REVOKED:
		return "the account is disabled, locked out or expired (KDC_ERR_CLIENT_REVOKED)"
	case errorcode.KDC_ERR_KEY_EXPIRED:
		return "the account's password has expired (KDC_ERR_KEY_EXPIRED)"
	case errorcode.KDC_ERR_POLICY:
		return "a domain policy refused the request (KDC_ERR_POLICY)"
	case errorcode.KDC_ERR_WRONG_REALM:
		return "the realm does not hold the account (KDC_ERR_WRONG_REALM)"
	default:
		return "the KDC returned Kerberos error " + strconv.Itoa(int(code))
	}
}

// replied records what a KDC sent over TCP, a 4-byte length and then the
// message, when it sent anything.
func (e *kdcExchange) replied(c *kdcConn) {
	if len(c.received) == 0 {
		return
	}
	e.answered, e.reply, e.answeredBy = true, nil, c.address
	e.badKeyParameters, e.delivered = c.badKeyParameters, c.delivered
	if len(c.received) < 4 {
		return
	}
	length := binary.BigEndian.Uint32(c.received)
	if length <= maxKDCReply && uint64(len(c.received)-4) >= uint64(length) {
		e.reply = c.received[4 : 4+length]
	}
}

// checkKeyParameters rejects what a KDC reply proposes for deriving the
// account's key from its password when gokrb5 must not act on it. gokrb5 uses
// the first entry of each PA-ETYPE-INFO and PA-ETYPE-INFO2 list as given, in a
// KDC error asking for preauthentication and in an AS reply: it panics on an
// empty list, accepts any encryption type, so a reply could switch the poller
// to RC4, and derives the key with any iteration count.
func checkKeyParameters(message []byte) error {
	var pas types.PADataSequence
	var krbErr messages.KRBError
	var asRep messages.ASRep
	switch {
	case krbErr.Unmarshal(message) == nil:
		if len(krbErr.EData) == 0 || pas.Unmarshal(krbErr.EData) != nil {
			return nil
		}
	case asRep.Unmarshal(message) == nil:
		pas = asRep.PAData
	default:
		return nil
	}
	for _, pa := range pas {
		switch pa.PADataType {
		case patype.PA_ETYPE_INFO2:
			// gokrb5 reports a list it cannot decode as an error.
			if info, err := pa.GetETypeInfo2(); err == nil &&
				(len(info) == 0 || !slices.Contains(kerberosEncTypes, info[0].EType) ||
					!acceptableIterations(info[0].S2KParams)) {
				return errKeyParameters
			}
		case patype.PA_ETYPE_INFO:
			if info, err := pa.GetETypeInfo(); err == nil &&
				(len(info) == 0 || !slices.Contains(kerberosEncTypes, info[0].EType)) {
				return errKeyParameters
			}
		}
	}
	return nil
}

// acceptableIterations checks AES string-to-key parameters: a 4-byte
// big-endian iteration count, which gokrb5 reads only at that length.
func acceptableIterations(params []byte) bool {
	if len(params) != 4 {
		return true
	}
	iterations := binary.BigEndian.Uint32(params)
	return iterations >= 1 && iterations <= maxS2KIterations
}

type kdcConn struct {
	net.Conn
	exchange *kdcExchange
	address  string
	stop     func() bool
	received []byte
	// badKeyParameters and delivered describe the reply, as in kdcExchange.
	badKeyParameters bool
	delivered        bool
}

// SetDeadline keeps gokrb5's fixed per-KDC deadline within the session's.
func (c *kdcConn) SetDeadline(t time.Time) error {
	if deadline, ok := c.exchange.ctx.Deadline(); ok && deadline.Before(t) {
		t = deadline
	}
	return c.Conn.SetDeadline(t)
}

func (c *kdcConn) Read(b []byte) (int, error) {
	var n int
	var err error
	if missing := 4 - len(c.received); missing > 0 && len(b) >= missing {
		// gokrb5 reads the length prefix with one Read and trusts whatever
		// arrived, so the prefix is read whole before it is checked.
		n, err = io.ReadFull(c.Conn, b[:missing])
	} else {
		n, err = c.Conn.Read(b)
	}
	c.received = append(c.received, b[:n]...)
	if len(c.received) < 4 {
		return n, err
	}
	length := binary.BigEndian.Uint32(c.received)
	if length > maxKDCReply {
		return n, errKDCReplyTooLarge
	}
	if uint64(len(c.received)-4) == uint64(length) {
		if checkKeyParameters(c.received[4:]) != nil {
			c.badKeyParameters = true
			// gokrb5 reads the message with io.ReadFull, which drops an error
			// returned with the last bytes; withhold them instead.
			return 0, errKeyParameters
		}
		c.delivered = true
	}
	return n, err
}

func (c *kdcConn) Close() error {
	c.stop()
	c.exchange.replied(c)
	return c.Conn.Close()
}
