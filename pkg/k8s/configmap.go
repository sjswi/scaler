package k8s

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *K8sClient) CreateConfigMap(name string, data map[string]string) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: data,
	}

	_, err := c.client.CoreV1().ConfigMaps("default").Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *K8sClient) DeleteConfigMap(name string) error {

	err := c.client.CoreV1().ConfigMaps("default").Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
