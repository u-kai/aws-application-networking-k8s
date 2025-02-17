package policyhelper

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	anv1alpha1 "github.com/aws/aws-application-networking-k8s/pkg/apis/applicationnetworking/v1alpha1"
)

type GroupKind struct {
	Group string
	Kind  string
}

func ObjToGroupKind(obj client.Object) GroupKind {
	switch obj.(type) {
	case *gwv1.Gateway:
		return GroupKind{gwv1.GroupName, "Gateway"}
	case *gwv1.HTTPRoute:
		return GroupKind{gwv1.GroupName, "HTTPRoute"}
	case *gwv1.GRPCRoute:
		return GroupKind{gwv1alpha2.GroupName, "GRPCRoute"}
	case *gwv1alpha2.TCPRoute:
		return GroupKind{gwv1alpha2.GroupName, "TCPRoute"}
	case *anv1alpha1.ServiceExport:
		return GroupKind{anv1alpha1.GroupName, "ServiceExport"}
	case *corev1.Service:
		return GroupKind{corev1.GroupName, "Service"}
	default:
		return GroupKind{}
	}
}

func TargetRefGroupKind(tr *TargetRef) GroupKind {
	return GroupKind{
		Group: string(tr.Group),
		Kind:  string(tr.Kind),
	}
}

func GroupKindToObj(gk GroupKind) (client.Object, bool) {
	switch gk {
	case GroupKind{gwv1.GroupName, "Gateway"}:
		return &gwv1.Gateway{}, true
	case GroupKind{gwv1.GroupName, "HTTPRoute"}:
		return &gwv1.HTTPRoute{}, true
	case GroupKind{gwv1alpha2.GroupName, "GRPCRoute"}:
		return &gwv1.GRPCRoute{}, true
	case GroupKind{gwv1alpha2.GroupName, "TCPRoute"}:
		return &gwv1alpha2.TCPRoute{}, true
	case GroupKind{corev1.GroupName, "Service"}:
		return &corev1.Service{}, true
	case GroupKind{anv1alpha1.GroupName, "ServiceExport"}:
		return &anv1alpha1.ServiceExport{}, true
	default:
		return nil, false
	}
}
