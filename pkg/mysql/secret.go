package mysql

import (
	"conserver/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (op *Operator) createSecret(name, dbName string) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"password": []byte("123456"),
			"database": []byte(dbName),
		},
	}
	client := k8s.GetK8sClient()
	err := client.CreateSecret(secret)
	if err != nil {
		return err
	}
	return nil
}
