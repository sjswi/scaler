package redis

import (
	"conserver/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (op *Operator) createService(appName, serviceName string) (int32, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: map[string]string{"app": appName},
			Ports: []corev1.ServicePort{
				{
					Port: 6379,
					Name: "redis",
				},
			},
		},
	}
	client := k8s.GetK8sClient()

	port, err := client.CreateService(service)
	if err != nil {
		return 0, err
	}

	return port, nil
}

