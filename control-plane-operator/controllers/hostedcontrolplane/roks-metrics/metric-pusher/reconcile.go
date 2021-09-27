package metrics_pusher

import (
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func reconcileRocksMetricsPusherServiceMonitor(svcMonitor *monitoring.ServiceMonitor) error {
	svcMonitor.Spec.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "push-gateway",
		},
	}

	svcMonitor.Spec.Endpoints = []monitoring.Endpoint{{
		Interval:    "30s",
		Port:        "http",
		Path:        "/metrics",
		HonorLabels: true,
	}}

	return nil
}

func reconcileRocksMetricsPusherService(svc *corev1.Service) error {
	svc.Spec.Selector = map[string]string{
		"app": "push-gateway",
	}
	var portSpec corev1.ServicePort
	if len(svc.Spec.Ports) > 0 {
		portSpec = svc.Spec.Ports[0]
	} else {
		svc.Spec.Ports = []corev1.ServicePort{portSpec}
	}
	portSpec.Port = int32(9091)
	portSpec.Name = "https"
	portSpec.Protocol = corev1.ProtocolTCP
	portSpec.TargetPort = intstr.FromInt(9091)
	svc.Spec.Ports[0] = portSpec
	svc.Spec.Type = corev1.ServiceTypeClusterIP
	return nil
}
