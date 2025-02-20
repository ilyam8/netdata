// SPDX-License-Identifier: GPL-3.0-or-later

package snmpsd

import (
	"fmt"
	"time"

	"github.com/gosnmp/gosnmp"
)

type (
	Config struct {
		RescanInterval time.Duration           `yaml:"rescan_interval"`
		Timeout        time.Duration           `yaml:"timeout"`
		ForgetDeadline time.Duration           `yaml:"forget_deadline"`
		ParallelScans  int                     `yaml:"parallel_scans"`
		Credentials    []configSnmpCredentials `yaml:"credentials"`
		Networks       []configNetwork         `yaml:"networks"`
	}

	configNetwork struct {
		Subnet     string `yaml:"subnet"`
		Credential string `yaml:"credential"`
	}
	configSnmpCredentials struct {
		Name              string `yaml:"name"`
		Version           string `yaml:"version"`
		Community         string `yaml:"community"`
		UserName          string `yaml:"username"`
		SecurityLevel     string `yaml:"security_level"`
		AuthProtocol      string `yaml:"auth_protocol"`
		AuthPassphrase    string `yaml:"auth_passphrase"`
		PrivacyProtocol   string `yaml:"privacy_protocol"`
		PrivacyPassphrase string `yaml:"privacy_passphrase"`
	}
)

func (c *Config) validate() error {
	if len(c.Credentials) == 0 {
		return fmt.Errorf("no credentials provided")
	}
	if len(c.Networks) == 0 {
		return fmt.Errorf("no networks provided")
	}
	return nil
}

func (c *Config) getCredential(name string) (configSnmpCredentials, bool) {
	for _, cred := range c.Credentials {
		if cred.Name == name {
			return cred, true
		}
	}
	return configSnmpCredentials{}, false
}

func setCredential(client gosnmp.Handler, conf configSnmpCredentials) {
	switch parseSNMPVersion(conf.Version) {
	case gosnmp.Version1:
		client.SetVersion(gosnmp.Version1)
		client.SetCommunity(conf.Community)
	case gosnmp.Version2c:
		client.SetVersion(gosnmp.Version2c)
		client.SetCommunity(conf.Community)
	case gosnmp.Version3:
		client.SetVersion(gosnmp.Version3)
		client.SetSecurityModel(gosnmp.UserSecurityModel)
		client.SetMsgFlags(parseSNMPv3SecurityLevel(conf.SecurityLevel))
		client.SetSecurityParameters(&gosnmp.UsmSecurityParameters{
			UserName:                 conf.UserName,
			AuthenticationProtocol:   parseSNMPv3AuthProtocol(conf.AuthProtocol),
			AuthenticationPassphrase: conf.AuthPassphrase,
			PrivacyProtocol:          parseSNMPv3PrivProtocol(conf.PrivacyProtocol),
			PrivacyPassphrase:        conf.PrivacyPassphrase,
		})
	}
}

func parseSNMPVersion(version string) gosnmp.SnmpVersion {
	switch version {
	case "0", "1":
		return gosnmp.Version1
	case "2", "2c", "":
		return gosnmp.Version2c
	case "3":
		return gosnmp.Version3
	default:
		return gosnmp.Version2c
	}
}

func parseSNMPv3SecurityLevel(level string) gosnmp.SnmpV3MsgFlags {
	switch level {
	case "1", "none", "noAuthNoPriv", "":
		return gosnmp.NoAuthNoPriv
	case "2", "authNoPriv":
		return gosnmp.AuthNoPriv
	case "3", "authPriv":
		return gosnmp.AuthPriv
	default:
		return gosnmp.NoAuthNoPriv
	}
}

func parseSNMPv3AuthProtocol(protocol string) gosnmp.SnmpV3AuthProtocol {
	switch protocol {
	case "1", "none", "noAuth", "":
		return gosnmp.NoAuth
	case "2", "md5":
		return gosnmp.MD5
	case "3", "sha":
		return gosnmp.SHA
	case "4", "sha224":
		return gosnmp.SHA224
	case "5", "sha256":
		return gosnmp.SHA256
	case "6", "sha384":
		return gosnmp.SHA384
	case "7", "sha512":
		return gosnmp.SHA512
	default:
		return gosnmp.NoAuth
	}
}

func parseSNMPv3PrivProtocol(protocol string) gosnmp.SnmpV3PrivProtocol {
	switch protocol {
	case "1", "none", "noPriv", "":
		return gosnmp.NoPriv
	case "2", "des":
		return gosnmp.DES
	case "3", "aes":
		return gosnmp.AES
	case "4", "aes192":
		return gosnmp.AES192
	case "5", "aes256":
		return gosnmp.AES256
	case "6", "aes192c":
		return gosnmp.AES192C
	case "7", "aes256c":
		return gosnmp.AES256C
	default:
		return gosnmp.NoPriv
	}
}
