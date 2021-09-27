package manifests

import (
	routev1 "github.com/openshift/api/route/v1"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MetricPusherRoute(namespace string) *routev1.Route {
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "push-gateway",
		},
	}
}

func MetricPusherService(namespace string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "push-gateway",
		},
	}
}

func MetricPusherServiceMonitor(namespace string) *monitoring.ServiceMonitor {
	return &monitoring.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "push-gateway",
		},
	}
}

func MetricPusherDeployment(namespace string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "push-gateway",
		},
	}
}
