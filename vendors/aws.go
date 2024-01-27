package vendors

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type EC2Options struct {
	Region    string
	AccessKey string
	SecretKey string
}

type EC2 struct {
	clients []*ec2.Client // By AWS design, one client is needed per region
	options EC2Options
}

type EC2Instance struct {
	Host    string
	IP      string
	Region  string
	OS      string
	Server  string
	Vendor  string
	Cluster string
}

func NewEC2(options EC2Options) (*EC2, error) {
	credentials := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(options.AccessKey, options.SecretKey, ""))
	clients := make([]*ec2.Client, 0)
	regions, err := GetAvailableAWSRegions(options)
	if err != nil {
		return nil, err
	}

	for _, region := range regions {
		clients = append(clients, ec2.New(ec2.Options{
			Region:      region,
			Credentials: credentials,
		}))
	}

	return &EC2{
		clients: clients,
		options: options,
	}, nil
}

func GetAvailableAWSRegions(options EC2Options) ([]string, error) {
	var regions []string
	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(options.AccessKey, options.SecretKey, ""))

	client := ec2.New(ec2.Options{
		Region:      options.Region,
		Credentials: creds,
	})
	rawRegions, err := client.DescribeRegions(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	for _, region := range rawRegions.Regions {
		regions = append(regions, *region.RegionName)
	}
	return regions, nil
}

func (e *EC2) GetAllEC2Instances() ([]EC2Instance, error) {
	instances := make([]EC2Instance, 0)
	for _, client := range e.clients {
		input := ec2.DescribeInstancesInput{}
		rawInstances, err := client.DescribeInstances(context.TODO(), &input)
		if err != nil {
			return nil, err
		}

		for _, reservation := range rawInstances.Reservations {
			for _, instance := range reservation.Instances {
				instances = append(instances, EC2Instance{
					Host:    *instance.InstanceId,
					IP:      *instance.PublicIpAddress,
					Region:  client.Options().Region,
					OS:      *instance.PlatformDetails,
					Vendor:  "aws",
					Server:  *instance.InstanceId,
					Cluster: "aws",
				})
			}
		}
	}
	return instances, nil
}
