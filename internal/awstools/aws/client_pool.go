package aws

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
)

// clientPoolThreadSafe is a concurrent map implementation to store multiple AWS clients.
type clientPoolThreadSafe struct {
	sync.Mutex

	clients map[ClientKey]Client
}

type ClientKey struct {
	Profile, Region string
}

func (p *clientPoolThreadSafe) set(key ClientKey, client Client) {
	p.Lock()
	p.clients[key] = client
	p.Unlock()
}

func startClient(
	wg *sync.WaitGroup,
	ctx context.Context,
	recordErr func(error),
	store func(*Client),
	configs ...func(*config.LoadOptions) error,
) {
	wg.Go(func() {
		err := ctx.Err()
		if err != nil {
			recordErr(err)
			return
		}

		client, err := NewClient(ctx, configs...)
		if err != nil {
			recordErr(err)
			return
		}

		store(client)
	})
}

// NewClientPool creates an AWS client for each permutation of the given profiles and regions.
// If profiles, regions, or both are empty, credentials and regions are picked up via
// the usual default provider chain, respectively. For example, if regions are empty,
// the region is first looked for via the according region environment variable or
// second the default region for each profile is used from `~/.aws/config`.
func NewClientPool(ctx context.Context, profiles []string, regions []string) (map[ClientKey]Client, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	var errOnce sync.Once
	var firstErr error

	recordErr := func(err error) {
		if err == nil {
			return
		}

		errOnce.Do(func() {
			firstErr = err
			cancel()
		})
	}

	clientPool := &clientPoolThreadSafe{
		clients: make(map[ClientKey]Client),
	}

	switch {
	case len(profiles) > 0 && len(regions) > 0:
		for _, profile := range profiles {
			for _, region := range regions {
				p := profile
				r := region
				startClient(
					&wg,
					ctx,
					recordErr,
					func(client *Client) {
						client.Profile = p
						clientPool.set(ClientKey{p, client.Region}, *client)
					},
					config.WithSharedConfigProfile(p),
					config.WithRegion(r),
				)
			}
		}
	case len(profiles) > 0:
		for _, profile := range profiles {
			p := profile
			startClient(
				&wg,
				ctx,
				recordErr,
				func(client *Client) {
					client.Profile = p
					clientPool.set(ClientKey{p, client.Region}, *client)
				},
				config.WithSharedConfigProfile(p),
			)
		}
	case len(regions) > 0:
		for _, region := range regions {
			r := region
			startClient(
				&wg,
				ctx,
				recordErr,
				func(client *Client) {
					clientPool.set(ClientKey{"", client.Region}, *client)
				},
				config.WithRegion(r),
			)
		}
	default:
		client, err := NewClient(ctx)
		if err != nil {
			return nil, err
		}

		return map[ClientKey]Client{{"", client.Region}: *client}, nil
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	return clientPool.clients, nil
}
