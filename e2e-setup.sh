export CLASTER_VPC_ID=vpc-0ed040bc35eac4ae6
export CLASTER_NAME=eks-auto-mode-sample
export AWS_ACCOUNT_ID=111815285043
export REGION=ap-northeast-1


kubectl kustomize "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v1.2.0" | kubectl apply -f -
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: aws-application-networking-system
  labels:
    control-plane: gateway-api-controller
EOF
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: gateway-api-controller
  namespace: aws-application-networking-system
EOF

cd helm
helm install gateway-api-controller . -f values.yaml -n aws-application-networking-system

cd ../

kubectl apply -f config/crds/bases/externaldns.k8s.io_dnsendpoints.yaml
kubectl apply -f config/crds/bases/gateway.networking.k8s.io_tlsroutes.yaml
kubectl apply -f config/crds/bases/application-networking.k8s.aws_serviceexports.yaml
kubectl apply -f config/crds/bases/application-networking.k8s.aws_serviceimports.yaml
kubectl apply -f config/crds/bases/application-networking.k8s.aws_targetgrouppolicies.yaml
kubectl apply -f config/crds/bases/application-networking.k8s.aws_vpcassociationpolicies.yaml
kubectl apply -f config/crds/bases/application-networking.k8s.aws_accesslogpolicies.yaml
kubectl apply -f config/crds/bases/application-networking.k8s.aws_iamauthpolicies.yaml
