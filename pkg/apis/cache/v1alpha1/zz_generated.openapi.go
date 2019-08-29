// +build !ignore_autogenerated

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/operator-framework/operator-sdk-samples/rocketmq-operator/pkg/apis/cache/v1alpha1.Broker":       schema_pkg_apis_cache_v1alpha1_Broker(ref),
		"github.com/operator-framework/operator-sdk-samples/rocketmq-operator/pkg/apis/cache/v1alpha1.BrokerSpec":   schema_pkg_apis_cache_v1alpha1_BrokerSpec(ref),
		"github.com/operator-framework/operator-sdk-samples/rocketmq-operator/pkg/apis/cache/v1alpha1.BrokerStatus": schema_pkg_apis_cache_v1alpha1_BrokerStatus(ref),
	}
}

func schema_pkg_apis_cache_v1alpha1_Broker(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "Broker is the Schema for the brokers API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/operator-framework/operator-sdk-samples/rocketmq-operator/pkg/apis/cache/v1alpha1.BrokerSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/operator-framework/operator-sdk-samples/rocketmq-operator/pkg/apis/cache/v1alpha1.BrokerStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/operator-framework/operator-sdk-samples/rocketmq-operator/pkg/apis/cache/v1alpha1.BrokerSpec", "github.com/operator-framework/operator-sdk-samples/rocketmq-operator/pkg/apis/cache/v1alpha1.BrokerStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_cache_v1alpha1_BrokerSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "BrokerSpec defines the desired state of Broker",
				Properties: map[string]spec.Schema{
					"size": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"nameServers": {
						SchemaProps: spec.SchemaProps{
							Description: "NameServers defines the name service list e.g. 192.168.1.1:9876;192.168.1.2:9876",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"replicationMode": {
						SchemaProps: spec.SchemaProps{
							Description: "ReplicationMode defines the replication is sync or async e.g. SYNC",
							Type:        []string{"string"},
							Format:      "",
						},
					},
				},
				Required: []string{"size"},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_cache_v1alpha1_BrokerStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "BrokerStatus defines the observed state of Broker",
				Properties: map[string]spec.Schema{
					"nodes": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type:   []string{"string"},
										Format: "",
									},
								},
							},
						},
					},
				},
				Required: []string{"nodes"},
			},
		},
		Dependencies: []string{},
	}
}
