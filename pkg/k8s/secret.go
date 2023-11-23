package k8s

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *K8sClient) CreateSecret(secret *corev1.Secret) error {

	_, err := c.client.CoreV1().Secrets("default").Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *K8sClient) DeleteSecret(name string) error {

	err := c.client.CoreV1().Secrets("default").Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
