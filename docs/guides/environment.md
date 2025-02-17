### Environment Variables
AWS Gateway API Controller for VPC Lattice supports a number of configuration options, which are set through environment variables.
The following environment variables are available, and all of them are optional.

---

#### `CLUSTER_NAME`

**Type:** *string*

**Default:** *Inferred from IMDS metadata*

A unique name to identify a cluster. This will be used in AWS resource tags to record ownership.
This variable is required except for EKS cluster. This needs to be specified if IMDS is not available.

---

#### `CLUSTER_VPC_ID`

**Type:** *string*

**Default:** *Inferred from IMDS metadata or CLUSTER_NAME*

When running AWS Gateway API Controller outside the Kubernetes Cluster, this specifies the VPC of the cluster. 

---

#### `AWS_ACCOUNT_ID`

**Type:** *string*

**Default:** *Inferred from IMDS metadata or AWS STS GetCallerIdentity API*

When running AWS Gateway API Controller outside the Kubernetes Cluster, this specifies the AWS account.

---

#### `REGION` or `AWS_REGION`

**Type:** *string*

**Default:** *Inferred from IMDS metadata*

When running AWS Gateway API Controller outside the Kubernetes Cluster, this specifies the AWS Region of VPC Lattice Service endpoint. This needs to be specified if IMDS is not available.

---

#### `LOG_LEVEL`

**Type:** *string*

**Default:** *"info"*

When set as "debug", the AWS Gateway API Controller will emit debug level logs.


---

#### `DEFAULT_SERVICE_NETWORK`

**Type:** *string*

**Default:** ""

When set as a non-empty value, creates a service network with that name.
The created service network will be also associated with cluster VPC.

---

#### `ENABLE_SERVICE_NETWORK_OVERRIDE`

**Type:** *string*

**Default:** ""

When set as "true", the controller will run in "single service network" mode that will override all gateways to point to default service network, instead of searching for service network with the same name. Can be used for small setups and conformance tests.

---

#### `WEBHOOK_ENABLED`

**Type:** *string*

**Default:** ""

When set as "true", the controller will start the webhook listener responsible for pod readiness gate injection 
(see `pod-readiness-gates.md`). This is disabled by default for `deploy.yaml` because the controller will not start 
successfully without the TLS certificate for the webhook in place. While this can be fixed by running 
`scripts/gen-webhook-cert.sh`, it requires manual action. The webhook is enabled by default for the Helm install
as the Helm install will also generate the necessary certificate.

---

#### `DISABLE_TAGGING_SERVICE_API`

**Type:** *string*

**Default:** ""

When set as "true", the controller will not use the [AWS Resource Groups Tagging API](https://docs.aws.amazon.com/resourcegroupstagging/latest/APIReference/overview.html). 

The Resource Groups Tagging API is only available on the public internet and customers using private clusters will need to enable this feature. When enabled, the controller will use VPC Lattice APIs to lookup tags which are not as performant and requires more API calls.

The Helm chart sets this value to "false" by default.

---

#### `ROUTE_MAX_CONCURRENT_RECONCILES`

**Type:** *int*

**Default:** 1

Maximum number of concurrently running reconcile loops per route type (HTTP, GRPC, TLS)