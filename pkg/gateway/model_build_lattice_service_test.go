package gateway

import (
	"context"
	"testing"

	"github.com/aws/aws-application-networking-k8s/pkg/config"
	"github.com/aws/aws-application-networking-k8s/pkg/k8s"
	"github.com/aws/aws-application-networking-k8s/pkg/model/core"
	model "github.com/aws/aws-application-networking-k8s/pkg/model/lattice"
	"github.com/aws/aws-application-networking-k8s/pkg/utils/gwlog"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	testclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func Test_LatticeServiceModelBuild(t *testing.T) {
	now := metav1.Now()
	var httpSectionName gwv1.SectionName = "http"
	var serviceKind gwv1.Kind = "Service"
	var serviceimportKind gwv1.Kind = "ServiceImport"
	var weight1 = int32(10)
	var weight2 = int32(90)
	var namespace = gwv1.Namespace("default")

	namespacePtr := func(ns string) *gwv1.Namespace {
		p := gwv1.Namespace(ns)
		return &p
	}

	var backendRef1 = gwv1.BackendRef{
		BackendObjectReference: gwv1.BackendObjectReference{
			Name:      "targetgroup1",
			Namespace: &namespace,
			Kind:      &serviceKind,
		},
		Weight: &weight1,
	}
	var backendRef2 = gwv1.BackendRef{
		BackendObjectReference: gwv1.BackendObjectReference{
			Name:      "targetgroup2",
			Namespace: &namespace,
			Kind:      &serviceimportKind,
		},
		Weight: &weight2,
	}

	tlsSectionName := gwv1.SectionName("tls")
	tlsModeTerminate := gwv1.TLSModeTerminate

	tests := []struct {
		name          string
		gw            []gwv1.Gateway
		gwClass       gwv1.GatewayClass
		route         core.Route
		wantErrIsNil  bool
		wantIsDeleted bool
		expected      model.ServiceSpec
	}{
		{
			name:          "Add LatticeService with hostname",
			wantIsDeleted: false,
			wantErrIsNil:  true,
			gwClass: gwv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gwClass1",
				},
				Spec: gwv1.GatewayClassSpec{
					ControllerName: config.LatticeGatewayControllerName,
				},
			},
			gw: []gwv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gateway1",
						Namespace: "default",
					},
					Spec: gwv1.GatewaySpec{
						GatewayClassName: "gwClass1",
					},
				},
			},
			route: core.NewHTTPRoute(gwv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service1",
					Namespace: "test",
				},
				Spec: gwv1.HTTPRouteSpec{
					CommonRouteSpec: gwv1.CommonRouteSpec{
						ParentRefs: []gwv1.ParentReference{
							{
								Name:      "gateway1",
								Namespace: namespacePtr("default"),
							},
						},
					},
					Hostnames: []gwv1.Hostname{
						"test1.test.com",
						"test2.test.com",
					},
				},
			}),
			expected: model.ServiceSpec{
				ServiceTagFields: model.ServiceTagFields{
					RouteName:      "service1",
					RouteNamespace: "test",
					RouteType:      core.HttpRouteType,
				},
				CustomerDomainName:  "test1.test.com",
				ServiceNetworkNames: []string{"gateway1"},
			},
		},
		{
			name:          "Add LatticeService",
			wantIsDeleted: false,
			wantErrIsNil:  true,
			gwClass: gwv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gwClass1",
				},
				Spec: gwv1.GatewayClassSpec{
					ControllerName: config.LatticeGatewayControllerName,
				},
			},
			gw: []gwv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gateway1",
						Namespace: "default",
					},
					Spec: gwv1.GatewaySpec{
						GatewayClassName: "gwClass1",
					},
				},
			},
			route: core.NewHTTPRoute(gwv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service1",
					Namespace: "default",
				},
				Spec: gwv1.HTTPRouteSpec{
					CommonRouteSpec: gwv1.CommonRouteSpec{
						ParentRefs: []gwv1.ParentReference{
							{
								Name:      "gateway1",
								Namespace: namespacePtr("default"),
							},
						},
					},
				},
			}),
			expected: model.ServiceSpec{
				ServiceTagFields: model.ServiceTagFields{
					RouteName:      "service1",
					RouteNamespace: "default",
					RouteType:      core.HttpRouteType,
				},
				ServiceNetworkNames: []string{"gateway1"},
			},
		},
		{
			name:          "Add LatticeService with GRPCRoute",
			wantIsDeleted: false,
			wantErrIsNil:  true,
			gwClass: gwv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gwClass1",
				},
				Spec: gwv1.GatewayClassSpec{
					ControllerName: config.LatticeGatewayControllerName,
				},
			},
			gw: []gwv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gateway1",
						Namespace: "test",
					},
					Spec: gwv1.GatewaySpec{
						GatewayClassName: "gwClass1",
					},
				},
			},
			route: core.NewGRPCRoute(gwv1.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service1",
					Namespace: "test",
				},
				Spec: gwv1.GRPCRouteSpec{
					CommonRouteSpec: gwv1.CommonRouteSpec{
						ParentRefs: []gwv1.ParentReference{
							{
								Name: "gateway1",
							},
						},
					},
				},
			}),
			expected: model.ServiceSpec{
				ServiceTagFields: model.ServiceTagFields{
					RouteName:      "service1",
					RouteNamespace: "test",
					RouteType:      core.GrpcRouteType,
				},
				ServiceNetworkNames: []string{"gateway1"},
			},
		},
		{
			name:          "Delete LatticeService",
			wantIsDeleted: true,
			wantErrIsNil:  true,
			gwClass: gwv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gwClass1",
				},
				Spec: gwv1.GatewayClassSpec{
					ControllerName: config.LatticeGatewayControllerName,
				},
			},
			gw: []gwv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gateway2",
						Namespace: "ns1",
					},
					Spec: gwv1.GatewaySpec{
						GatewayClassName: "gwClass1",
						Listeners: []gwv1.Listener{
							{
								Name:     httpSectionName,
								Port:     80,
								Protocol: "HTTP",
							},
						},
					},
				},
			},
			route: core.NewHTTPRoute(gwv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "service2",
					Namespace:         "ns1",
					Finalizers:        []string{"gateway.k8s.aws/resources"},
					DeletionTimestamp: &now, // <- the important bit
				},
				Spec: gwv1.HTTPRouteSpec{
					CommonRouteSpec: gwv1.CommonRouteSpec{
						ParentRefs: []gwv1.ParentReference{
							{
								Name:        "gateway2",
								SectionName: &httpSectionName,
							},
						},
					},
					Rules: []gwv1.HTTPRouteRule{
						{
							BackendRefs: []gwv1.HTTPBackendRef{
								{
									BackendRef: backendRef1,
								},
								{
									BackendRef: backendRef2,
								},
							},
						},
					},
				},
			}),
			expected: model.ServiceSpec{
				ServiceTagFields: model.ServiceTagFields{
					RouteName:      "service2",
					RouteNamespace: "ns1",
					RouteType:      core.HttpRouteType,
				},
				ServiceNetworkNames: []string{"gateway2"},
			},
		},
		{
			name:          "Service with customer Cert ARN",
			wantIsDeleted: false,
			wantErrIsNil:  true,
			gwClass: gwv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gwClass1",
				},
				Spec: gwv1.GatewayClassSpec{
					ControllerName: config.LatticeGatewayControllerName,
				},
			},
			gw: []gwv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gateway1",
						Namespace: "default",
					},
					Spec: gwv1.GatewaySpec{
						GatewayClassName: "gwClass1",
						Listeners: []gwv1.Listener{
							{
								Name:     "tls",
								Port:     443,
								Protocol: "HTTPS",
								TLS: &gwv1.GatewayTLSConfig{
									Mode:            &tlsModeTerminate,
									CertificateRefs: nil,
									Options: map[gwv1.AnnotationKey]gwv1.AnnotationValue{
										"application-networking.k8s.aws/certificate-arn": "cert-arn",
									},
								},
							},
						},
					},
				},
			},
			route: core.NewHTTPRoute(gwv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service1",
					Namespace: "default",
				},
				Spec: gwv1.HTTPRouteSpec{
					CommonRouteSpec: gwv1.CommonRouteSpec{
						ParentRefs: []gwv1.ParentReference{
							{
								Name:        "gateway1",
								Namespace:   namespacePtr("default"),
								SectionName: &tlsSectionName,
							},
						},
					},
				},
			}),
			expected: model.ServiceSpec{
				ServiceTagFields: model.ServiceTagFields{
					RouteName:      "service1",
					RouteNamespace: "default",
					RouteType:      core.HttpRouteType,
				},
				CustomerCertARN:     "cert-arn",
				ServiceNetworkNames: []string{"gateway1"},
			},
		},
		{
			//TODO: 見直すことGWClassって本当に必要？
			name: "GW does not exist",
			gwClass: gwv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gwClass1",
				},
				Spec: gwv1.GatewayClassSpec{
					ControllerName: config.LatticeGatewayControllerName,
				},
			},
			gw: []gwv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gateway1",
						Namespace: "default",
					},
					Spec: gwv1.GatewaySpec{
						GatewayClassName: "gwClass1",
					},
				},
			},
			route: core.NewHTTPRoute(gwv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service1",
					Namespace: "default",
				},
				Spec: gwv1.HTTPRouteSpec{
					CommonRouteSpec: gwv1.CommonRouteSpec{
						ParentRefs: []gwv1.ParentReference{
							{
								Name:      "not-a-real-gateway",
								Namespace: namespacePtr("default"),
							},
						},
					},
				},
			}),
			wantErrIsNil: false,
		},
		{
			name:          "Service with TLS section but no cert arn",
			wantIsDeleted: false,
			wantErrIsNil:  true,
			gwClass: gwv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gwClass1",
				},
				Spec: gwv1.GatewayClassSpec{
					ControllerName: config.LatticeGatewayControllerName,
				},
			},
			gw: []gwv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gateway1",
						Namespace: "default",
					},
					Spec: gwv1.GatewaySpec{
						GatewayClassName: "gwClass1",
						Listeners: []gwv1.Listener{
							{
								Name:     "tls",
								Port:     443,
								Protocol: "HTTPS",
								TLS: &gwv1.GatewayTLSConfig{
									Mode:            &tlsModeTerminate,
									CertificateRefs: nil,
								},
							},
						},
					},
				},
			},
			route: core.NewHTTPRoute(gwv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service1",
					Namespace: "default",
				},
				Spec: gwv1.HTTPRouteSpec{
					CommonRouteSpec: gwv1.CommonRouteSpec{
						ParentRefs: []gwv1.ParentReference{
							{
								Name:        "gateway1",
								Namespace:   namespacePtr("default"),
								SectionName: &tlsSectionName,
							},
						},
					},
				},
			}),
			expected: model.ServiceSpec{
				ServiceTagFields: model.ServiceTagFields{
					RouteName:      "service1",
					RouteNamespace: "default",
					RouteType:      core.HttpRouteType,
				},
				ServiceNetworkNames: []string{"gateway1"},
			},
		},
		{
			name:          "Multiple service networks",
			wantIsDeleted: false,
			wantErrIsNil:  true,
			gwClass: gwv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gwClass1",
				},
				Spec: gwv1.GatewayClassSpec{
					ControllerName: config.LatticeGatewayControllerName,
				},
			},
			gw: []gwv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gateway1",
						Namespace: "default",
					},
					Spec: gwv1.GatewaySpec{
						GatewayClassName: "gwClass1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gateway2",
						Namespace: "ns2",
					},
					Spec: gwv1.GatewaySpec{
						GatewayClassName: "gwClass1",
					},
				},
			},
			route: core.NewHTTPRoute(gwv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service1",
					Namespace: "default",
				},
				Spec: gwv1.HTTPRouteSpec{
					CommonRouteSpec: gwv1.CommonRouteSpec{
						ParentRefs: []gwv1.ParentReference{
							{
								Name:      "gateway1",
								Namespace: namespacePtr("default"),
							},
							{
								Name:      "gateway2",
								Namespace: namespacePtr("ns2"),
							},
						},
					},
				},
			}),
			expected: model.ServiceSpec{
				ServiceTagFields: model.ServiceTagFields{
					RouteName:      "service1",
					RouteNamespace: "default",
					RouteType:      core.HttpRouteType,
				},
				ServiceNetworkNames: []string{"gateway1", "gateway2"},
			},
		},
		{
			name:          "Route has multiple parents, one of which is not a Lattice gateway",
			wantIsDeleted: false,
			wantErrIsNil:  true,
			gwClass: gwv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gwClass1",
				},
				Spec: gwv1.GatewayClassSpec{
					ControllerName: config.LatticeGatewayControllerName,
				},
			},
			gw: []gwv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gateway1",
						Namespace: "default",
					},
					Spec: gwv1.GatewaySpec{
						GatewayClassName: "gwClass1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "not-lattice-gateway",
						Namespace: "ns2",
					},
					Spec: gwv1.GatewaySpec{
						GatewayClassName: "not-lattice",
					},
				},
			},
			route: core.NewHTTPRoute(gwv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service1",
					Namespace: "default",
				},
				Spec: gwv1.HTTPRouteSpec{
					CommonRouteSpec: gwv1.CommonRouteSpec{
						ParentRefs: []gwv1.ParentReference{
							{
								Name:      "gateway1",
								Namespace: namespacePtr("default"),
							},
							{
								Name:      "not-lattice-gateway",
								Namespace: namespacePtr("ns2"),
							},
						},
					},
				},
			}),
			expected: model.ServiceSpec{
				ServiceTagFields: model.ServiceTagFields{
					RouteName:      "service1",
					RouteNamespace: "default",
					RouteType:      core.HttpRouteType,
				},
				ServiceNetworkNames: []string{"gateway1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			ctx := context.TODO()

			k8sSchema := runtime.NewScheme()
			clientgoscheme.AddToScheme(k8sSchema)
			gwv1.Install(k8sSchema)
			k8sClient := testclient.NewClientBuilder().WithScheme(k8sSchema).Build()

			assert.NoError(t, k8sClient.Create(ctx, tt.gwClass.DeepCopy()))
			for _, gw := range tt.gw {
				assert.NoError(t, k8sClient.Create(ctx, gw.DeepCopy()))
			}
			stack := core.NewDefaultStack(core.StackID(k8s.NamespacedName(tt.route.K8sObject())))

			task := &latticeServiceModelBuildTask{
				log:    gwlog.FallbackLogger,
				route:  tt.route,
				stack:  stack,
				client: k8sClient,
			}

			svc, err := task.buildLatticeService(ctx)
			if !tt.wantErrIsNil {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)

			assert.Equal(t, tt.wantIsDeleted, svc.IsDeleted)

			assert.Equal(t, tt.expected.RouteName, svc.Spec.RouteName)
			assert.Equal(t, tt.expected.RouteNamespace, svc.Spec.RouteNamespace)
			assert.Equal(t, tt.expected.CustomerCertARN, svc.Spec.CustomerCertARN)
			assert.Equal(t, tt.expected.CustomerDomainName, svc.Spec.CustomerDomainName)
			assert.Equal(t, tt.expected.RouteType, svc.Spec.RouteType)
			assert.Equal(t, tt.expected.ServiceNetworkNames, svc.Spec.ServiceNetworkNames)
		})
	}
}
