package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		s3.NewBucket(ctx, "my-bucket", nil)

		vpc, err := ec2.NewVpc(ctx, "main", &ec2.VpcArgs{
			CidrBlock: pulumi.String("10.0.0.0/16"),
		})

		if err != nil {
			return err
		}

		ec2.NewSubnet(ctx, "subnet-private", &ec2.SubnetArgs{
			VpcId:     vpc.ID(),
			CidrBlock: pulumi.String("10.0.1.0/25"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("subnet-private"),
			},
		})

		pubsub, err := ec2.NewSubnet(ctx, "subnet-public", &ec2.SubnetArgs{
			VpcId:     vpc.ID(),
			CidrBlock: pulumi.String("10.0.0.0/26"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("subnet-public"),
			},
		})

		pubsub2, err := ec2.NewSubnet(ctx, "subnet-public2", &ec2.SubnetArgs{
			VpcId:     vpc.ID(),
			CidrBlock: pulumi.String("10.0.160.0/20"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("subnet-public2"),
			},
		})

		ec2.NewSubnet(ctx, "subnet-private2", &ec2.SubnetArgs{
			VpcId:     vpc.ID(),
			CidrBlock: pulumi.String("10.0.128.0/20"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("subnet-private2"),
			},
		})

		pubsub3, err := ec2.NewSubnet(ctx, "subnet-public3", &ec2.SubnetArgs{
			VpcId:              vpc.ID(),
			CidrBlock:          pulumi.String("10.0.112.0/20"),
			AvailabilityZoneId: pulumi.String("use1-az2"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("subnet-public3"),
			},
		})

		ec2.NewSubnet(ctx, "subnet-private3", &ec2.SubnetArgs{
			VpcId:     vpc.ID(),
			CidrBlock: pulumi.String("10.0.144.0/20"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("subnet-private3"),
			},
		})

		gateway, err := ec2.NewInternetGateway(ctx, "gw", &ec2.InternetGatewayArgs{
			VpcId: vpc.ID(),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("internet"),
			},
		})

		routetable, err := ec2.NewRouteTable(ctx, "example", &ec2.RouteTableArgs{
			VpcId: vpc.ID(),
			Routes: ec2.RouteTableRouteArray{
				&ec2.RouteTableRouteArgs{
					CidrBlock: pulumi.String("0.0.0.0/0"),
					GatewayId: gateway.ID(),
				},
			},

			Tags: pulumi.StringMap{
				"Name": pulumi.String("example"),
			},
		})

		ec2.NewRouteTableAssociation(ctx, "routeTableAssociation", &ec2.RouteTableAssociationArgs{
			SubnetId:     pubsub.ID(),
			RouteTableId: routetable.ID(),
		})

		ec2.NewRouteTableAssociation(ctx, "routeTableAssociation2", &ec2.RouteTableAssociationArgs{
			SubnetId:     pubsub2.ID(),
			RouteTableId: routetable.ID(),
		})

		ec2.NewRouteTableAssociation(ctx, "routeTableAssociation3", &ec2.RouteTableAssociationArgs{
			SubnetId:     pubsub3.ID(),
			RouteTableId: routetable.ID(),
		})

		example, err := eks.NewCluster(ctx, "example", &eks.ClusterArgs{
			RoleArn: pulumi.String("arn:aws:iam::153743130237:role/org-admin"),
			VpcConfig: &eks.ClusterVpcConfigArgs{
				PublicAccessCidrs: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
				SubnetIds: pulumi.StringArray{
					pubsub.ID(),
					pubsub2.ID(),
					pubsub3.ID(),
				},
			},
		})

		eks.NewNodeGroup(ctx, "Istio_nodegroup", &eks.NodeGroupArgs{
			ClusterName: pulumi.String("example-5b13529"),
			//eks.NodeGroupTaintArrayInput
			Taints: eks.NodeGroupTaintArgs{
				Effect: pulumi.StringInput("NO_SCHEDULE"),
				// Effect: pulumi.String("NO_SCHEDULE"),
				// Key:    pulumi.String("dedicated"),
			},

			SubnetIds: pulumi.StringArray{
				pubsub.ID(),
				pubsub2.ID(),
				pubsub3.ID(),
			},
			NodeRoleArn: pulumi.String("arn:aws:iam::153743130237:role/org-admin"),
			ScalingConfig: &eks.NodeGroupScalingConfigArgs{
				DesiredSize: pulumi.Int(2),
				MaxSize:     pulumi.Int(2),
				MinSize:     pulumi.Int(1),
			},
		})

		eks.NewNodeGroup(ctx, "workload_nodegroup", &eks.NodeGroupArgs{
			ClusterName: pulumi.String("example-5b13529"),
			SubnetIds: pulumi.StringArray{
				pubsub.ID(),
				pubsub2.ID(),
				pubsub3.ID(),
			},
			NodeRoleArn: pulumi.String("arn:aws:iam::153743130237:role/org-admin"),
			ScalingConfig: &eks.NodeGroupScalingConfigArgs{
				DesiredSize: pulumi.Int(2),
				MaxSize:     pulumi.Int(2),
				MinSize:     pulumi.Int(1),
			},
		})

		// if nodegroup == nil && nodegroup2 == nil {
		// 	return nil
		// }

		if err != nil {
			return err
		}
		ctx.Export("endpoint", example.Endpoint)
		// ctx.Export("kubeconfig-certificate-authority-data", example.CertificateAuthority.ApplyT(func(certificateAuthority eks.ClusterCertificateAuthority) (string, error) {
		// 	return certificateAuthority.Data, nil
		// }).(pulumi.StringOutput))

		if routetable == nil {
			return nil
		}

		if err != nil {
			return err
		}

		return nil

	})

}
