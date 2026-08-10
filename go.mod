module github.com/chenrui333/terradozer

go 1.26.2

require (
	github.com/apex/log v1.9.0
	github.com/aws/aws-sdk-go v1.55.8
	github.com/aws/aws-sdk-go-v2 v1.41.7
	github.com/aws/aws-sdk-go-v2/config v1.32.17
	github.com/aws/aws-sdk-go-v2/service/accessanalyzer v1.47.2
	github.com/aws/aws-sdk-go-v2/service/acm v1.38.3
	github.com/aws/aws-sdk-go-v2/service/acmpca v1.46.14
	github.com/aws/aws-sdk-go-v2/service/apigateway v1.39.3
	github.com/aws/aws-sdk-go-v2/service/apigatewayv2 v1.34.3
	github.com/aws/aws-sdk-go-v2/service/applicationautoscaling v1.41.16
	github.com/aws/aws-sdk-go-v2/service/appmesh v1.35.14
	github.com/aws/aws-sdk-go-v2/service/apprunner v1.39.16
	github.com/aws/aws-sdk-go-v2/service/appsync v1.53.7
	github.com/aws/aws-sdk-go-v2/service/athena v1.57.6
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.66.2
	github.com/aws/aws-sdk-go-v2/service/autoscalingplans v1.30.16
	github.com/aws/aws-sdk-go-v2/service/backup v1.55.2
	github.com/aws/aws-sdk-go-v2/service/batch v1.64.1
	github.com/aws/aws-sdk-go-v2/service/budgets v1.43.6
	github.com/aws/aws-sdk-go-v2/service/cloud9 v1.33.22
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.71.11
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.62.0
	github.com/aws/aws-sdk-go-v2/service/cloudhsmv2 v1.34.25
	github.com/aws/aws-sdk-go-v2/service/cloudtrail v1.55.11
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.57.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatchevents v1.32.25
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.72.0
	github.com/aws/aws-sdk-go-v2/service/codeartifact v1.38.23
	github.com/aws/aws-sdk-go-v2/service/codebuild v1.68.15
	github.com/aws/aws-sdk-go-v2/service/codecommit v1.33.14
	github.com/aws/aws-sdk-go-v2/service/codedeploy v1.35.15
	github.com/aws/aws-sdk-go-v2/service/codepipeline v1.46.23
	github.com/aws/aws-sdk-go-v2/service/codestarconnections v1.35.15
	github.com/aws/aws-sdk-go-v2/service/codestarnotifications v1.31.23
	github.com/aws/aws-sdk-go-v2/service/cognitoidentity v1.33.24
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.60.2
	github.com/aws/aws-sdk-go-v2/service/configservice v1.62.3
	github.com/aws/aws-sdk-go-v2/service/costandusagereportservice v1.34.15
	github.com/aws/aws-sdk-go-v2/service/databasemigrationservice v1.62.2
	github.com/aws/aws-sdk-go-v2/service/datapipeline v1.30.22
	github.com/aws/aws-sdk-go-v2/service/datasync v1.58.4
	github.com/aws/aws-sdk-go-v2/service/dax v1.29.18
	github.com/aws/aws-sdk-go-v2/service/devicefarm v1.38.10
	github.com/aws/aws-sdk-go-v2/service/directconnect v1.38.17
	github.com/aws/aws-sdk-go-v2/service/directoryservice v1.38.18
	github.com/aws/aws-sdk-go-v2/service/dlm v1.36.2
	github.com/aws/aws-sdk-go-v2/service/docdb v1.48.15
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.57.3
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.299.1
	github.com/aws/aws-sdk-go-v2/service/ecr v1.57.2
	github.com/aws/aws-sdk-go-v2/service/ecrpublic v1.38.15
	github.com/aws/aws-sdk-go-v2/service/ecs v1.79.1
	github.com/aws/aws-sdk-go-v2/service/efs v1.41.16
	github.com/aws/aws-sdk-go-v2/service/eks v1.83.0
	github.com/aws/aws-sdk-go-v2/service/elasticache v1.52.2
	github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk v1.34.4
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.33.25
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.54.12
	github.com/aws/aws-sdk-go-v2/service/elasticsearchservice v1.40.2
	github.com/aws/aws-sdk-go-v2/service/elastictranscoder v1.33.0
	github.com/aws/aws-sdk-go-v2/service/emr v1.59.2
	github.com/aws/aws-sdk-go-v2/service/firehose v1.42.16
	github.com/aws/aws-sdk-go-v2/service/fms v1.44.24
	github.com/aws/aws-sdk-go-v2/service/fsx v1.65.9
	github.com/aws/aws-sdk-go-v2/service/gamelift v1.54.0
	github.com/aws/aws-sdk-go-v2/service/glacier v1.32.8
	github.com/aws/aws-sdk-go-v2/service/globalaccelerator v1.35.18
	github.com/aws/aws-sdk-go-v2/service/glue v1.140.1
	github.com/aws/aws-sdk-go-v2/service/guardduty v1.75.3
	github.com/aws/aws-sdk-go-v2/service/iam v1.53.10
	github.com/aws/aws-sdk-go-v2/service/imagebuilder v1.53.1
	github.com/aws/aws-sdk-go-v2/service/inspector v1.30.22
	github.com/aws/aws-sdk-go-v2/service/iot v1.73.0
	github.com/aws/aws-sdk-go-v2/service/kafka v1.51.0
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.43.7
	github.com/aws/aws-sdk-go-v2/service/kinesisanalytics v1.30.25
	github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2 v1.37.2
	github.com/aws/aws-sdk-go-v2/service/kinesisvideo v1.33.10
	github.com/aws/aws-sdk-go-v2/service/kms v1.51.1
	github.com/aws/aws-sdk-go-v2/service/lakeformation v1.47.8
	github.com/aws/aws-sdk-go-v2/service/lambda v1.90.1
	github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice v1.34.16
	github.com/aws/aws-sdk-go-v2/service/licensemanager v1.37.12
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.53.2
	github.com/aws/aws-sdk-go-v2/service/macie v1.19.2
	github.com/aws/aws-sdk-go-v2/service/macie2 v1.51.2
	github.com/aws/aws-sdk-go-v2/service/mediaconvert v1.91.2
	github.com/aws/aws-sdk-go-v2/service/mediapackage v1.39.23
	github.com/aws/aws-sdk-go-v2/service/mediastore v1.29.23
	github.com/aws/aws-sdk-go-v2/service/mq v1.34.22
	github.com/aws/aws-sdk-go-v2/service/mwaa v1.39.24
	github.com/aws/aws-sdk-go-v2/service/neptune v1.44.5
	github.com/aws/aws-sdk-go-v2/service/networkfirewall v1.60.1
	github.com/aws/aws-sdk-go-v2/service/opsworks v1.31.0
	github.com/aws/aws-sdk-go-v2/service/organizations v1.51.3
	github.com/aws/aws-sdk-go-v2/service/pinpoint v1.39.23
	github.com/aws/aws-sdk-go-v2/service/qldb v1.32.2
	github.com/aws/aws-sdk-go-v2/service/quicksight v1.108.0
	github.com/aws/aws-sdk-go-v2/service/ram v1.36.5
	github.com/aws/aws-sdk-go-v2/service/rds v1.118.2
	github.com/aws/aws-sdk-go-v2/service/redshift v1.62.7
	github.com/aws/aws-sdk-go-v2/service/resourcegroups v1.33.26
	github.com/aws/aws-sdk-go-v2/service/route53 v1.62.7
	github.com/aws/aws-sdk-go-v2/service/route53resolver v1.42.7
	github.com/aws/aws-sdk-go-v2/service/s3 v1.100.1
	github.com/aws/aws-sdk-go-v2/service/s3control v1.70.1
	github.com/aws/aws-sdk-go-v2/service/s3outposts v1.34.14
	github.com/aws/aws-sdk-go-v2/service/sagemaker v1.244.0
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.41.7
	github.com/aws/aws-sdk-go-v2/service/securityhub v1.69.2
	github.com/aws/aws-sdk-go-v2/service/servicecatalog v1.39.14
	github.com/aws/aws-sdk-go-v2/service/servicediscovery v1.39.28
	github.com/aws/aws-sdk-go-v2/service/servicequotas v1.34.7
	github.com/aws/aws-sdk-go-v2/service/ses v1.34.24
	github.com/aws/aws-sdk-go-v2/service/sfn v1.40.12
	github.com/aws/aws-sdk-go-v2/service/shield v1.34.23
	github.com/aws/aws-sdk-go-v2/service/signer v1.32.7
	github.com/aws/aws-sdk-go-v2/service/sns v1.39.17
	github.com/aws/aws-sdk-go-v2/service/sqs v1.42.27
	github.com/aws/aws-sdk-go-v2/service/ssm v1.68.6
	github.com/aws/aws-sdk-go-v2/service/ssoadmin v1.38.0
	github.com/aws/aws-sdk-go-v2/service/storagegateway v1.43.16
	github.com/aws/aws-sdk-go-v2/service/sts v1.42.1
	github.com/aws/aws-sdk-go-v2/service/swf v1.33.18
	github.com/aws/aws-sdk-go-v2/service/synthetics v1.42.16
	github.com/aws/aws-sdk-go-v2/service/timestreamwrite v1.35.22
	github.com/aws/aws-sdk-go-v2/service/transfer v1.72.0
	github.com/aws/aws-sdk-go-v2/service/waf v1.30.22
	github.com/aws/aws-sdk-go-v2/service/wafregional v1.30.23
	github.com/aws/aws-sdk-go-v2/service/wafv2 v1.71.5
	github.com/aws/aws-sdk-go-v2/service/worklink v1.23.2
	github.com/aws/aws-sdk-go-v2/service/workspaces v1.68.1
	github.com/aws/aws-sdk-go-v2/service/xray v1.36.23
	github.com/fatih/color v1.19.0
	github.com/gruntwork-io/terratest v0.23.0
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/go-plugin v1.3.0
	github.com/hashicorp/terraform v0.12.31
	github.com/mitchellh/cli v1.0.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/gomega v1.40.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.11.1
	github.com/zclconf/go-cty v1.7.1
	go.uber.org/mock v0.6.0
)

require (
	cel.dev/expr v0.25.1 // indirect
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/auth v0.18.2 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/iam v1.5.3 // indirect
	cloud.google.com/go/monitoring v1.24.3 // indirect
	cloud.google.com/go/storage v1.61.3 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.31.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.55.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.55.0 // indirect
	github.com/agext/levenshtein v1.2.2 // indirect
	github.com/apparentlymart/go-cidr v1.0.1 // indirect
	github.com/apparentlymart/go-textseg v1.0.0 // indirect
	github.com/apparentlymart/go-textseg/v12 v12.0.0 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.10 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.19.16 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.24 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.11.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.21 // indirect
	github.com/aws/smithy-go v1.25.1 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/bgentry/speakeasy v0.1.0 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/bmatcuk/doublestar v1.1.5 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20251210132809-ee656c7534f5 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.36.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.3.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.14 // indirect
	github.com/googleapis/gax-go/v2 v2.17.0 // indirect
	github.com/hashicorp/aws-sdk-go-base/v2 v2.0.0-beta.72 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-getter v1.8.6 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/hashicorp/go-uuid v1.0.1 // indirect
	github.com/hashicorp/go-version v1.8.0 // indirect
	github.com/hashicorp/hcl v0.0.0-20170504190234-a4b07c25de5f // indirect
	github.com/hashicorp/hcl/v2 v2.3.0 // indirect
	github.com/hashicorp/hil v0.0.0-20190212112733-ab17b08d6590 // indirect
	github.com/hashicorp/terraform-config-inspect v0.0.0-20191212124732-c6ae6269b9d7 // indirect
	github.com/hashicorp/terraform-svchost v0.0.0-20191011084731-65d371908596 // indirect
	github.com/hashicorp/yamux v0.0.0-20180604194846-3520598351bb // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/posener/complete v1.2.1 // indirect
	github.com/pquerna/otp v1.2.0 // indirect
	github.com/spf13/afero v1.10.0 // indirect
	github.com/spiffe/go-spiffe/v2 v2.6.0 // indirect
	github.com/ulikunitz/xz v0.5.15 // indirect
	github.com/vmihailenco/msgpack/v4 v4.3.12 // indirect
	github.com/vmihailenco/tagparser v0.1.1 // indirect
	github.com/zclconf/go-cty-yaml v1.0.1 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.39.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/metric v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.51.0 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	google.golang.org/api v0.271.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20260128011058-8636f8732409 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260203192932-546029d2fa20 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260226221140-a57be14db171 // indirect
	google.golang.org/grpc v1.80.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Terraform 0.12's snapshotFS does not implement methods added by newer afero.
replace github.com/spf13/afero => github.com/spf13/afero v1.3.3
