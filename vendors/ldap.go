package vendors

import (
	"crypto/tls"
	"fmt"

	"encoding/json"

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

func (l *Ldap) SearchWithScope(baseDN string, scope int, filter string, attributes []string) (*ldap.SearchResult, error) {
	if err := l.Connect(); err != nil {
		return nil, err
	}

	searchRequest := ldap.NewSearchRequest(
		baseDN,
		scope, // 0 to specify ScopeBaseObject, 1 ScopeOneLevel, or 2 ScopeWholeSubtree
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		attributes,
		nil,
	)

	result, err := l.client.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %v", err)
	}

	return result, nil
}

func (l *Ldap) GetGroupMembers(filter string) ([]byte, error) {
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
		filter,
		[]string{"distinguishedName", "cn", "memberUid", "*"},
		nil,
	)

	result, err := l.client.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %v", err)
	}

	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("object not found, using this filter: %s", filter)
	}

	memberUIDs := result.Entries[0].GetAttributeValues("memberUid")
	if len(memberUIDs) == 0 {
		return nil, fmt.Errorf("no members found for this group")
	}

	membersJson, err := json.Marshal(memberUIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal members: %v", err)
	}

	return membersJson, nil

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
