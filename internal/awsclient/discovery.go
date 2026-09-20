package awsclient

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/PYTHON01100100/ec2s/internal/config"
)

const perEnvTimeout = 20 * time.Second

// Result is the outcome of discovering instances in a single Environment.
// Errors are per-Environment data, not fatal: one bad profile or expired SSO
// token must not prevent other accounts/regions from populating.
type Result struct {
	Env       config.Environment
	Instances []Instance
	Err       error
}

type clientFactory func(ctx context.Context, env config.Environment) (ec2.DescribeInstancesAPIClient, error)

// Discover concurrently fetches EC2 instances for every given Environment,
// fanning out one goroutine per (account, region) pair and fanning the
// results back in. A slow or failing Environment cannot block or fail the
// others.
func Discover(ctx context.Context, envs []config.Environment) []Result {
	return discover(ctx, envs, func(ctx context.Context, env config.Environment) (ec2.DescribeInstancesAPIClient, error) {
		return NewClient(ctx, env)
	})
}

func discover(ctx context.Context, envs []config.Environment, newClient clientFactory) []Result {
	results := make(chan Result, len(envs))
	var wg sync.WaitGroup

	for _, env := range envs {
		wg.Add(1)
		go func(env config.Environment) {
			defer wg.Done()

			envCtx, cancel := context.WithTimeout(ctx, perEnvTimeout)
			defer cancel()

			instances, err := fetchOne(envCtx, env, newClient)
			results <- Result{Env: env, Instances: instances, Err: err}
		}(env)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	out := make([]Result, 0, len(envs))
	for r := range results {
		out = append(out, r)
	}
	return out
}

func fetchOne(ctx context.Context, env config.Environment, newClient clientFactory) ([]Instance, error) {
	client, err := newClient(ctx, env)
	if err != nil {
		return nil, err
	}

	var instances []Instance
	paginator := ec2.NewDescribeInstancesPaginator(client, &ec2.DescribeInstancesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, reservation := range page.Reservations {
			for _, raw := range reservation.Instances {
				instances = append(instances, fromSDK(raw, env))
			}
		}
	}

	return instances, nil
}
