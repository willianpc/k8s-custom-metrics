package provider

import (
	"context"
	"log"
	"time"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/metrics/pkg/apis/custom_metrics"

	// "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/apimachinery/pkg/labels"
	// "k8s.io/apimachinery/pkg/runtime/schema"
	// "k8s.io/apimachinery/pkg/types"
	// "k8s.io/client-go/dynamic"
	// "k8s.io/metrics/pkg/apis/custom_metrics"
	// "sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider/defaults"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider/helpers"
)

var _ provider.CustomMetricsProvider = (*yourProvider)(nil)

type yourProvider struct {
	defaults.DefaultCustomMetricsProvider
	defaults.DefaultExternalMetricsProvider
	client dynamic.Interface
	mapper apimeta.RESTMapper

	// just increment values when they're requested
	values map[provider.CustomMetricInfo]int64
}

func NewProvider(client dynamic.Interface, mapper apimeta.RESTMapper) provider.CustomMetricsProvider {
	return &yourProvider{
		client: client,
		mapper: mapper,
		values: make(map[provider.CustomMetricInfo]int64),
	}
}

func (p *yourProvider) GetMetricByName(ctx context.Context, name types.NamespacedName, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValue, error) {
	log.Println("GetMetricByName:", name, "; metric selector:", metricSelector)
	value, err := p.valueFor(info)
	if err != nil {
		return nil, err
	}
	return p.metricFor(value, name, info)
}

func (p *yourProvider) GetMetricBySelector(ctx context.Context, namespace string, selector labels.Selector, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValueList, error) {
	log.Println("GetMetricBySelector, namespace:", namespace, "; selector:", selector, "; metric selector:", metricSelector)
	totalValue, err := p.valueFor(info)
	if err != nil {
		return nil, err
	}

	names, err := helpers.ListObjectNames(p.mapper, p.client, namespace, selector, info)
	if err != nil {
		return nil, err
	}

	res := make([]custom_metrics.MetricValue, len(names))
	for i, name := range names {
		// in a real adapter, you might want to consider pre-computing the
		// object reference created in metricFor, instead of recomputing it
		// for each object.
		value, err := p.metricFor(100*totalValue/int64(len(res)), types.NamespacedName{Namespace: namespace, Name: name}, info)
		if err != nil {
			return nil, err
		}
		res[i] = *value
	}

	return &custom_metrics.MetricValueList{
		Items: res,
	}, nil
}

// valueFor fetches a value from the fake list and increments it.
func (p *yourProvider) valueFor(info provider.CustomMetricInfo) (int64, error) {
	// normalize the value so that you treat plural resources and singular
	// resources the same (e.g. pods vs pod)
	info, _, err := info.Normalized(p.mapper)
	if err != nil {
		return 0, err
	}

	value := p.values[info]
	value += 1
	p.values[info] = value

	return value, nil
}

// metricFor constructs a result for a single metric value.
func (p *yourProvider) metricFor(value int64, name types.NamespacedName, info provider.CustomMetricInfo) (*custom_metrics.MetricValue, error) {
	// construct a reference referring to the described object
	objRef, err := helpers.ReferenceFor(p.mapper, name, info)
	if err != nil {
		return nil, err
	}

	return &custom_metrics.MetricValue{
		DescribedObject: objRef,
		Metric: custom_metrics.MetricIdentifier{
			Name: info.Metric,
		},
		// you'll want to use the actual timestamp in a real adapter
		Timestamp: metav1.Time{time.Now()},
		Value:     *resource.NewMilliQuantity(value*100, resource.DecimalSI),
	}, nil
}
