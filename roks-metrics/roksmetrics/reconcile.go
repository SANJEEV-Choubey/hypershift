package metrics

import (
	"github.com/openshift/hypershift/hypershift-operator/controllers/util"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sutilspointer "k8s.io/utils/pointer"
)

func ReconcileRoksMetricsDeployment(deployment *appsv1.Deployment, sa *corev1.ServiceAccount, roksMetricsImage string) error {
	defaultMode := int32(420)
	roksMetricsLabels := map[string]string{"app": "metrics"}
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
								Optional:    util.True(),
							},
						},
					},
				},
				Containers: []corev1.Container{
					{
						Name:            "metrics",
						Image:           roksMetricsImage,
						ImagePullPolicy: corev1.PullAlways,
						Command:         []string{"/usr/bin/roks-metrics"},
						Args: []string{
							"--alsologtostderr",
							"--v=3",
							"--listen=:8443",
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "serving-cert",
								ReadOnly:  true,
								MountPath: "/var/run/secrets/serving-cert",
							},
						},
					},
				},
			},
		},
	}
	return nil
}

func ReconcileRoksMetricsClusterRole(role *rbacv1.ClusterRole) error {
	role.Rules = []rbacv1.PolicyRule{
		{
			APIGroups: []string{"config.openshift.io"},
			Resources: []string{"infrastructures", "featuregates", "proxies"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"build.openshift.io"},
			Resources: []string{"builds"},
			Verbs:     []string{"get", "list", "watch"},
		},
	}
	return nil
}

func ReconcileRoksMetricsRoleBinding(binding *rbacv1.ClusterRoleBinding, role *rbacv1.ClusterRole, sa *corev1.ServiceAccount) error {
	binding.RoleRef = rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     role.Name,
	}

	binding.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      sa.Name,
			Namespace: sa.Namespace,
		},
	}

	return nil
}

func ReconcileRocksMetricsServiceMonitor(svcMonitor *monitoring.ServiceMonitor) error {
	svcMonitor.Spec.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "metrics",
		},
	}
	relabelConfigs := []*monitoring.RelabelConfig{
		{
			Action:       "drop",
			SourceLabels: []string{"__name__"},
			Regex:        "apiserver_.*",
		},
		{
			Action:       "drop",
			SourceLabels: []string{"__name__"},
			Regex:        "go_.*",
		},
		{
			Action:       "drop",
			SourceLabels: []string{"__name__"},
			Regex:        "promhttp_.*",
		},
	}

	svcMonitor.Spec.Endpoints = []monitoring.Endpoint{{
		Interval:             "30s",
		Port:                 "https",
		Scheme:               "https",
		Path:                 "/metrics",
		MetricRelabelConfigs: relabelConfigs,
		TLSConfig: &monitoring.TLSConfig{
			CAFile: "/etc/prometheus/configmaps/serving-certs-ca-bundle/service-ca.crt",
			SafeTLSConfig: monitoring.SafeTLSConfig{
				ServerName: "roks-metrics.openshift-roks-metrics.svc",
			},
		},
	}}

	return nil
}

func ReconcileRocksMetricsService(svc *corev1.Service) error {
	svc.Spec.Selector = map[string]string{
		"app": "metrics",
	}
	var portSpec corev1.ServicePort
	if len(svc.Spec.Ports) > 0 {
		portSpec = svc.Spec.Ports[0]
	} else {
		svc.Spec.Ports = []corev1.ServicePort{portSpec}
	}
	portSpec.Port = int32(8443)
	portSpec.Name = "https"
	portSpec.Protocol = corev1.ProtocolTCP
	portSpec.TargetPort = intstr.FromInt(8443)
	svc.Spec.Ports[0] = portSpec
	svc.Spec.Type = corev1.ServiceTypeClusterIP
	return nil
}

func ReconcilePrometheusRoleBinding(binding *rbacv1.RoleBinding) error {
	binding.RoleRef = rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     "prometheus-k8s",
	}

	binding.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "prometheus-k8s",
			Namespace: "openshift-monitoring",
		},
	}

	return nil
}
