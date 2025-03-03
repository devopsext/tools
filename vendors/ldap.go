package vendors

import (
	"crypto/tls"
	"fmt"

	"github.com/devopsext/tools/common"
	"github.com/go-ldap/ldap/v3"
)

type Ldap struct {
	client  *ldap.Conn
	options LdapOptions
	logger  common.Logger
}

type LdapOptions struct {
	URL      string // ldaps://your-server:636
	User     string
	Password string
	BaseDN   string
	Insecure bool
	Timeout  int
}

type GroupMember struct {
	DN    string
	CN    string
	Email string
}

func (l *Ldap) Connect() error {
	if l.client != nil {
		return nil
	}

	dialOpts := []ldap.DialOpt{
		ldap.DialWithTLSConfig(&tls.Config{
			InsecureSkipVerify: l.options.Insecure,
		}),
	}

	conn, err := ldap.DialURL(l.options.URL, dialOpts...)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}

	if err := conn.Bind(l.options.User, l.options.Password); err != nil {
		conn.Close()
		return fmt.Errorf("failed to bind: %v", err)
	}

	l.client = conn
	return nil
}

func (l *Ldap) Close() {
	if l.client != nil {
		l.client.Close()
		l.client = nil
	}
}

func (l *Ldap) GetGroupMembers(groupDN string) ([]GroupMember, error) {
	if err := l.Connect(); err != nil {
		return nil, err
	}

	searchRequest := ldap.NewSearchRequest(
		l.options.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(objectClass=user)(memberOf=%s))", groupDN),
		[]string{"distinguishedName", "cn", "mail"},
		nil,
	)

	result, err := l.client.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %v", err)
	}

	var members []GroupMember
	for _, entry := range result.Entries {
		member := GroupMember{
			DN:    entry.GetAttributeValue("distinguishedName"),
			CN:    entry.GetAttributeValue("cn"),
			Email: entry.GetAttributeValue("mail"),
		}
		members = append(members, member)
	}

	return members, nil
}

func NewLdapClient(options LdapOptions, logger common.Logger) (*Ldap, error) {
	ldap := &Ldap{
		options: options,
		logger:  logger,
	}

	// Test connection on initialization
	if err := ldap.Connect(); err != nil {
		return nil, err
	}

	return ldap, nil
}
