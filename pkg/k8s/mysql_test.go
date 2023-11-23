package k8s

import (
	"testing"
)

func TestMySQL(t *testing.T) {

	client := GetK8sClient()

	client.CreateReader("mysql-deploy-1a2f72-0")

}
