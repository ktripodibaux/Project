package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ebs"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/rds"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi-postgresql/sdk/v3/go/postgresql"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		helm.NewChart(ctx, "istio-system", helm.ChartArgs{
			Path: pulumi.String("./istio-system"),
		})

		// helm.NewChart(ctx, "wordpress", helm.ChartArgs{
		// 	Path: pulumi.String("./wordpress"),
		// })

		// wordpress, err := helmv2.NewChart(ctx, "wpdev", helmv2.ChartArgs{
		// 	Version: pulumi.String("15.2.14"),
		// 	Chart:   pulumi.String("wordpress"),
		// 	FetchArgs: &helmv2.FetchArgs{
		// 		Repo: pulumi.String("https://charts.bitnami.com/bitnami"),
		// 	},
		// })

		// frontendIP := wordpress.GetResource("v1/Service", "wpdev-wordpress", "default").ApplyT(func(r interface{}) (pulumi.StringPtrOutput, error) {
		// 	svc := r.(*corev1.Service)
		// 	return svc.Status.LoadBalancer().Ingress().Index(pulumi.Int(0)).Ip(), nil
		// })
		// ctx.Export("frontendIp", frontendIP)

		// s3.NewBucket(ctx, "my-bucket", nil)

		selected, err := s3.LookupBucket(ctx, &s3.LookupBucketArgs{
			Bucket: "kurt-boundlessbucket",
		}, nil)

		if selected == nil {
			return nil
		}

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
			VpcId:               vpc.ID(),
			MapPublicIpOnLaunch: pulumi.Bool(true),
			CidrBlock:           pulumi.String("10.0.0.0/26"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("subnet-public"),
			},
		})

		pubsub2, err := ec2.NewSubnet(ctx, "subnet-public2", &ec2.SubnetArgs{
			VpcId:               vpc.ID(),
			MapPublicIpOnLaunch: pulumi.Bool(true),
			CidrBlock:           pulumi.String("10.0.160.0/20"),
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
			VpcId:               vpc.ID(),
			MapPublicIpOnLaunch: pulumi.Bool(true),
			CidrBlock:           pulumi.String("10.0.112.0/20"),
			AvailabilityZoneId:  pulumi.String("use1-az2"),
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
			ClusterName: pulumi.String("example-b992118"),
			//eks.NodeGroupTaintArrayInput
			Taints: eks.NodeGroupTaintArray{
				// eks.NodeGroupTaintArgs{
				// 	Effect: pulumi.String("NO_SCHEDULE"),
				// 	Key:    pulumi.String("istio"),
				// 	Value:  pulumi.String("istio"),
				// },
				eks.NodeGroupTaintArgs{
					Effect: pulumi.String("NO_SCHEDULE"),
					Key:    pulumi.String("dedicated"),
					Value:  pulumi.String("istiod"),
				},
			},

			SubnetIds: pulumi.StringArray{
				pubsub.ID(),
				pubsub2.ID(),
				pubsub3.ID(),
			},
			NodeRoleArn: pulumi.String("arn:aws:iam::153743130237:role/org-admin"),
			ScalingConfig: &eks.NodeGroupScalingConfigArgs{
				DesiredSize: pulumi.Int(3),
				MaxSize:     pulumi.Int(4),
				MinSize:     pulumi.Int(1),
			},
		})

		eks.NewNodeGroup(ctx, "workload_nodegroup", &eks.NodeGroupArgs{
			ClusterName: pulumi.String("example-b992118"),
			SubnetIds: pulumi.StringArray{
				pubsub.ID(),
				pubsub2.ID(),
				pubsub3.ID(),
			},
			NodeRoleArn: pulumi.String("arn:aws:iam::153743130237:role/org-admin"),
			ScalingConfig: &eks.NodeGroupScalingConfigArgs{
				DesiredSize: pulumi.Int(3),
				MaxSize:     pulumi.Int(4),
				MinSize:     pulumi.Int(1),
			},
		})

		ec2.NewDefaultVpc(ctx, "default", &ec2.DefaultVpcArgs{
			Tags: pulumi.StringMap{
				"Name": pulumi.String("Default VPC"),
			},
		})

		rds.NewCluster(ctx, "testpostgresql", &rds.ClusterArgs{
			AvailabilityZones: pulumi.StringArray{
				pulumi.String("us-east-1c"),
				pulumi.String("us-east-1d"),
				pulumi.String("us-east-1a"),
			},
			BackupRetentionPeriod: pulumi.Int(5),
			ClusterIdentifier:     pulumi.String("aurora-cluster-test"),
			DatabaseName:          pulumi.String("mytestdb"),
			Engine:                pulumi.String("aurora-postgresql"),
			MasterPassword:        pulumi.String("Testing123"),
			MasterUsername:        pulumi.String("testUser"),
			PreferredBackupWindow: pulumi.String("07:00-09:00"),
			// FinalSnapshotIdentifier: pulumi.String("title"),
		})

		// rds.NewClusterInstance(ctx, "test1", &rds.ClusterInstanceArgs{
		// 	ApplyImmediately:  pulumi.Bool(true),
		// 	ClusterIdentifier: cluster.ID(),
		// 	Identifier:        pulumi.String("test1"),
		// 	InstanceClass:     pulumi.String("db.t2.small"),
		// 	Engine:            cluster.Engine,
		// 	EngineVersion:     cluster.EngineVersion,
		// })

		postgresql.NewProvider(ctx, "db", &postgresql.ProviderArgs{
			Database:         pulumi.String("mytestdb"),
			DatabaseUsername: pulumi.String("testUser"),
			Password:         pulumi.String("Testing123"),
		})

		rds.NewInstance(ctx, "default", &rds.InstanceArgs{
			AllocatedStorage: pulumi.Int(10),
			DbName:           pulumi.String("mydb"),
			Engine:           pulumi.String("postgres"),
			// EngineVersion:      pulumi.String("5.7"),
			InstanceClass: pulumi.String("db.t3.micro"),
			// ParameterGroupName: pulumi.String("default.mysql5.7"),
			Password:          pulumi.String("foobarbaz"),
			SkipFinalSnapshot: pulumi.Bool(true),
			Username:          pulumi.String("foo"),
		})

		// postgresql.Database()

		// rds.NewClusterInstance(ctx, fmt.Sprintf("clusterInstances-%v", 1), &rds.ClusterInstanceArgs{
		// 	Identifier:        pulumi.String(fmt.Sprintf("aurora-cluster-test-%v", 1)),
		// 	ClusterIdentifier: db.ID(),
		// 	InstanceClass:     pulumi.String("db.r4.large"),
		// 	Engine:            db.Engine,
		// 	EngineVersion:     db.EngineVersion,
		// })
		// if err != nil {
		// 	return err
		// }

		netinterface, err := ec2.NewNetworkInterface(ctx, "test", &ec2.NetworkInterfaceArgs{
			SubnetId: pubsub.ID(),
		})

		eip, err := ec2.NewEip(ctx, "lb", &ec2.EipArgs{
			// Instance:         pulumi.String("i-07592843f75ba47bc"),
			NetworkInterface: netinterface.ID(),
			Vpc:              pulumi.Bool(true),
		})

		ec2.NewNatGateway(ctx, "example", &ec2.NatGatewayArgs{
			AllocationId: eip.ID(),
			SubnetId:     pubsub.ID(),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("gw NAT"),
			},
		})

		ebs.NewVolume(ctx, "example", &ebs.VolumeArgs{
			AvailabilityZone: pulumi.String("us-east-1a"),
			Size:             pulumi.Int(8),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("HelloWorld"),
			},
		})

		// if nodegroup == nil && nodegroup2 == nil {
		// 	return nil
		// }

		// if err != nil {
		// 	return err
		// }
		ctx.Export("endpoint", example.Endpoint)

		// // ctx.Export("kubeconfig-certificate-authority-data", example.CertificateAuthority.ApplyT(func(certificateAuthority eks.ClusterCertificateAuthority) (string, error) {
		// // 	return certificateAuthority.Data, nil
		// // }).(pulumi.StringOutput))

		// if routetable == nil {
		// 	return nil
		// }

		// if err != nil {
		// 	return err
		// }

		return nil

	})

}
