package instance

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Client wraps the AWS EC2 API for instances.
type Client struct {
	ec2Client *ec2.Client
}

// NewClient creates a new EC2 instance client from an AWS config.
func NewClient(cfg aws.Config) *Client {
	return &Client{
		ec2Client: ec2.NewFromConfig(cfg),
	}
}

// ListInstances lists all EC2 instances in the configured region using pagination.
func (c *Client) ListInstances(ctx context.Context) ([]types.Instance, error) {
	var instances []types.Instance

	paginator := ec2.NewDescribeInstancesPaginator(c.ec2Client, &ec2.DescribeInstancesInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("describe instances: %w", err)
		}

		for _, reservation := range output.Reservations {
			instances = append(instances, reservation.Instances...)
		}
	}

	return instances, nil
}
