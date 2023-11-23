package k8s

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *K8sClient) CreateService(svc *corev1.Service) (int32, error) {

	svc, err := c.client.CoreV1().Services("default").Create(context.TODO(), svc, metav1.CreateOptions{})
	if err != nil {
		return 0, err
	}

	// 获取分配的NodePort
	nodePort := svc.Spec.Ports[0].NodePort

	return nodePort, nil
}

func (c *K8sClient) DeleteService(name string) error {

	err := c.client.CoreV1().Services("default").Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}
