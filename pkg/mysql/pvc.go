package mysql

import (
	"conserver/pkg/k8s"
)

func (op *Operator) createPVC(name string) error {
	client := k8s.GetK8sClient()
	err := client.CreatePVC(name)
	if err != nil {
		return err
	}
	return nil
}
