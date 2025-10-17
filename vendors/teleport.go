package vendors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
	teleport "github.com/gravitational/teleport/api/client"
	"google.golang.org/grpc"
)

type TeleportResourceOptions struct {
	Kind string
}

type TeleportResourceListOptions struct {
	TeleportResourceOptions
}

type TeleportOptions struct {
	Address  string
	Identity string
	Timeout  int
	Insecure bool
}

type Teleport struct {
	options TeleportOptions
	logger  common.Logger
	client  *teleport.Client
}

const (
	TeleportResourceKubernetes = "kubernetes"
)

func newTeleportClient(options TeleportOptions, ctx context.Context) (*teleport.Client, error) {

	var creds teleport.Credentials

	if utils.FileExists(options.Identity) {
		creds = teleport.LoadIdentityFile(options.Identity)
	} else {
		creds = teleport.LoadIdentityFileFromString(options.Identity)
	}

	client, err := teleport.New(ctx, teleport.Config{
		Addrs:                    []string{options.Address},
		Credentials:              []teleport.Credentials{creds},
		DialOpts:                 []grpc.DialOption{},
		InsecureAddressDiscovery: options.Insecure,
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (t *Teleport) getClientCtx(options TeleportOptions) (*teleport.Client, context.Context, context.CancelFunc, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(options.Timeout)*time.Second)

	client := t.client
	if client == nil || options != t.options {
		cs, err := newTeleportClient(options, ctx)
		if err != nil {
			return nil, ctx, cancel, err
		}
		t.client = cs
		client = cs
	}

	return client, ctx, cancel, nil
}

func (t *Teleport) CustomPing(options TeleportOptions) ([]byte, error) {

	client, ctx, cancel, err := t.getClientCtx(options)
	if err != nil {
		return nil, err
	}
	defer cancel()

	resp, err := client.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(resp)
}

func (t *Teleport) Ping() ([]byte, error) {
	return t.CustomPing(t.options)
}

func (t *Teleport) CustomResourceList(options TeleportOptions, listOptions TeleportResourceListOptions) ([]byte, error) {

	client, ctx, cancel, err := t.getClientCtx(options)
	if err != nil {
		return nil, err
	}
	defer cancel()

	var data []byte

	switch listOptions.Kind {
	case TeleportResourceKubernetes:

		servers, err := client.GetKubernetesServers(ctx)
		if err != nil {
			return nil, err
		}
		data, err = json.Marshal(servers)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("Teleport has unsupported kind for listing: %s", listOptions.Kind)
	}

	return data, nil
}

func (t *Teleport) ResourceList(options TeleportResourceListOptions) ([]byte, error) {
	return t.CustomResourceList(t.options, options)
}

func NewTeleport(options TeleportOptions, logger common.Logger) *Teleport {

	return &Teleport{
		options: options,
		logger:  logger,
		client:  nil,
	}
}
