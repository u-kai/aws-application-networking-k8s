package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/sts"
)

const (
	LatticeGatewayControllerName = "application-networking.k8s.aws/gateway-api-controller"
	defaultLogLevel              = "Info"
)

const (
	REGION                          = "REGION"
	AWS_REGION                      = "AWS_REGION"
	CLUSTER_VPC_ID                  = "CLUSTER_VPC_ID"
	CLUSTER_NAME                    = "CLUSTER_NAME"
	DEFAULT_SERVICE_NETWORK         = "DEFAULT_SERVICE_NETWORK"
	DISABLE_TAGGING_SERVICE_API     = "DISABLE_TAGGING_SERVICE_API"
	ENABLE_SERVICE_NETWORK_OVERRIDE = "ENABLE_SERVICE_NETWORK_OVERRIDE"
	AWS_ACCOUNT_ID                  = "AWS_ACCOUNT_ID"
	DEV_MODE                        = "DEV_MODE"
	WEBHOOK_ENABLED                 = "WEBHOOK_ENABLED"
	ROUTE_MAX_CONCURRENT_RECONCILES = "ROUTE_MAX_CONCURRENT_RECONCILES"
)

var VpcID = ""
var AccountID = ""
var Region = ""
var DefaultServiceNetwork = ""
var ClusterName = ""
var DevMode = ""
var WebhookEnabled = ""

var DisableTaggingServiceAPI = false
var ServiceNetworkOverrideMode = false
var RouteMaxConcurrentReconciles = 1

func ConfigInit() error {
	sess, _ := session.NewSession()
	metadata := NewEC2Metadata(sess)
	return configInit(sess, metadata)
}

func configInit(sess *session.Session, metadata EC2Metadata) error {
	var err error

	var metadataErr error
	if Region = os.Getenv(REGION); Region == "" {
		if Region, metadataErr = metadata.Region(); metadataErr != nil {
			if Region = os.Getenv(AWS_REGION); Region == "" {
				return fmt.Errorf("region is not specified")
			}
		}
	}

	if ClusterName = os.Getenv(CLUSTER_NAME); ClusterName == "" {
		if sess == nil {
			return fmt.Errorf("cluster name is not specified")
		}
		if ClusterName, err = getClusterName(sess, Region); err != nil {
			return fmt.Errorf("cannot get cluster name: %s", err)
		}
	}

	if VpcID = os.Getenv(CLUSTER_VPC_ID); VpcID == "" {
		if metadataErr != nil {
			if VpcID, err = fromClusterNameToVPCId(sess, ClusterName); err != nil {
				return fmt.Errorf("vpcId is not specified: %s", err)
			}
		} else if VpcID, err = metadata.VpcID(); err != nil {
			return fmt.Errorf("vpcId is not specified: %s", err)
		}
	}

	if AccountID = os.Getenv(AWS_ACCOUNT_ID); AccountID == "" {
		if metadataErr != nil {
			if AccountID, err = fromIdentityToAccountId(sess); err != nil {
				return fmt.Errorf("account is not specified: %s", err)
			}
		} else if AccountID, err = metadata.AccountId(); err != nil {
			return fmt.Errorf("account is not specified: %s", err)
		}
	}

	DevMode = os.Getenv(DEV_MODE)
	WebhookEnabled = os.Getenv(WEBHOOK_ENABLED)

	DefaultServiceNetwork = os.Getenv(DEFAULT_SERVICE_NETWORK)

	overrideFlag := os.Getenv(ENABLE_SERVICE_NETWORK_OVERRIDE)
	if strings.ToLower(overrideFlag) == "true" && DefaultServiceNetwork != "" {
		ServiceNetworkOverrideMode = true
	}

	disableTaggingAPI := os.Getenv(DISABLE_TAGGING_SERVICE_API)

	if strings.ToLower(disableTaggingAPI) == "true" {
		DisableTaggingServiceAPI = true
	}

	routeMaxConcurrentReconciles := os.Getenv(ROUTE_MAX_CONCURRENT_RECONCILES)
	if routeMaxConcurrentReconciles != "" {
		routeMaxConcurrentReconcilesInt, err := strconv.Atoi(routeMaxConcurrentReconciles)
		if err != nil {
			return fmt.Errorf("invalid value for ROUTE_MAX_CONCURRENT_RECONCILES: %s", err)
		}
		RouteMaxConcurrentReconciles = routeMaxConcurrentReconcilesInt
	}

	return nil
}

// try to find cluster name, search in env then in ec2 instance tags
func getClusterName(sess *session.Session, region string) (string, error) {
	meta := ec2metadata.New(sess)
	doc, err := meta.GetInstanceIdentityDocument()
	if err != nil {
		return "", err
	}
	instanceId := doc.InstanceID
	ec2Client := ec2.New(sess, &aws.Config{Region: aws.String(region)})
	tagReq := &ec2.DescribeTagsInput{Filters: []*ec2.Filter{{
		Name:   aws.String("resource-id"),
		Values: []*string{aws.String(instanceId)},
	}}}
	tagRes, err := ec2Client.DescribeTags(tagReq)
	if err != nil {
		return "", err
	}
	for _, tag := range tagRes.Tags {
		if *tag.Key == "aws:eks:cluster-name" {
			return *tag.Value, nil
		}
	}
	return "", errors.New("not found in env and metadata")
}

func fromClusterNameToVPCId(sess *session.Session, clusterName string) (string, error) {
	eksClient := eks.New(sess)
	clusterConf, err := eksClient.DescribeClusterWithContext(context.Background(), &eks.DescribeClusterInput{Name: aws.String(clusterName)})
	if err != nil {
		return "", err
	}
	if clusterConf.Cluster.ResourcesVpcConfig == nil {
		return "", fmt.Errorf("VPC ID is not found in cluster %s", clusterName)
	}
	return *clusterConf.Cluster.ResourcesVpcConfig.VpcId, nil
}

func fromIdentityToAccountId(sess *session.Session) (string, error) {
	stsClient := sts.New(sess)
	identity, err := stsClient.GetCallerIdentityWithContext(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}
	if identity.Account == nil {
		return "", fmt.Errorf("account id is not found")
	}
	return *identity.Account, nil
}
