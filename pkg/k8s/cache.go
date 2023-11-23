package k8s

import (
	"bytes"
	"conserver/pkg/util"
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

func (c *K8sClient) LoadData(podName string) {

	cmd := []string{
		"bash",
		"/etc/mysql/conf.d/load.sh",
	}
	c.execCommand(cmd, podName)

}

func (c *K8sClient) DumpData(podName, masterPort, masterHost string) (string, string) {

	cmd := []string{
		"bash",
		"/etc/mysql/conf.d/dump.sh",
		masterHost,
		masterPort,
	}
	output := c.execCommand(cmd, podName)
	fmt.Println(output)
	return util.ParseFileAndPos(output)
}

func (c *K8sClient) StartSlave(podName, masterPort, masterHost, file, pos string) {
	cmd := []string{
		"bash",
		"/etc/mysql/conf.d/start_slave.sh",
		masterHost,
		masterPort,
		file,
		pos,
	}
	c.execCommand(cmd, podName)

}

func (c *K8sClient) CreateReader(podName string) {
	cmd := []string{
		"bash",
		"/etc/mysql/conf.d/create_reader.sh",
	}
	c.execCommand(cmd, podName)

}

func (c *K8sClient) execCommand(cmd []string, podName string) string {
	namespace := "default"

	req := c.client.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Command: cmd,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		panic(err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Stdin:  nil,
	})
	if err != nil {
		fmt.Println("STDERR:", stderr.String())
		panic(err)
	}

	fmt.Println("STDOUT:", stdout.String())
	return stdout.String()
}
