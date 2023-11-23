package k8s

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *K8sClient) CreateStatefulSet(statefulSet *appsv1.StatefulSet) error {
	_, err := c.client.AppsV1().StatefulSets("default").Create(context.TODO(), statefulSet, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *K8sClient) DeleteStatefulSet(name string) error {
	err := c.client.AppsV1().StatefulSets("default").Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *K8sClient) GetStatefulSet(name string) (*appsv1.StatefulSet, error) {
	statefulSet, err := c.client.AppsV1().StatefulSets("default").Get(context.TODO(), name, metav1.GetOptions{})
	return statefulSet, err
}
