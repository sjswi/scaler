package k8s

import (
	"conserver/pkg/util"
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *K8sClient) CreatePVC(name string) error {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
			// 如果你有特定的StorageClass，可以指定它
			StorageClassName: util.StringPtr("rook-cephfs"),
		},
	}

	_, err := c.client.CoreV1().PersistentVolumeClaims("default").Create(context.TODO(), pvc, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *K8sClient) DeletePVC(name string) error {

	err := c.client.CoreV1().PersistentVolumeClaims("default").Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
