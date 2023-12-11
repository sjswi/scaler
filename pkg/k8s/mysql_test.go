package k8s

import (
	"fmt"
	"testing"
)

func TestMySQL(t *testing.T) {

	client := GetK8sClient()

	client.CreateReader("mysql-deploy-1a2f72-0")

}

func TestGetInstance(t *testing.T) {
	endpoint := "sdasds"
	if true {
		endpoint = "sdadasda"
	}
	fmt.Println(endpoint)
}
