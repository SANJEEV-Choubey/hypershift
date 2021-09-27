package metrics_pusher

import (
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sutilspointer "k8s.io/utils/pointer"
)

func reconcileRoksMetricsDeployment(deployment *appsv1.Deployment, sa *corev1.ServiceAccount, roksMetricsImage string) error {
	defaultMode := int32(420)
	roksMetricsLabels := map[string]string{"app": "push-gateway"}
	deployment.Spec = appsv1.DeploymentSpec{
		Replicas: k8sutilspointer.Int32Ptr(1),
		Selector: &metav1.LabelSelector{
			MatchLabels: roksMetricsLabels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: roksMetricsLabels,
			},
			Spec: corev1.PodSpec{
				ServiceAccountName: sa.Name,
				Volumes: []corev1.Volume{
					{
						Name: "serving-cert",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								DefaultMode: &defaultMode,
								SecretName:  "serving-cert",
								//Optional:    &true,
							},
						},
					},
				},
				Tolerations: []corev1.Toleration{
					{
						Key:      "multi-az-worker",
						Operator: "Equal",
						Value:    "true",
						Effect:   corev1.TaintEffectNoSchedule,
					},
				},
				Containers: []corev1.Container{
					{
						Name:            "push-gateway",
						Image:           roksMetricsImage,
						ImagePullPolicy: corev1.PullAlways,
						Command:         []string{"pushgateway"},
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: 9091,
							},
						},
					},
				},
			},
		},
	}
	return nil
}

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
