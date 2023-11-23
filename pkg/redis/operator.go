package redis

import (
	"conserver/pkg/global"
	"conserver/pkg/k8s"
	"conserver/pkg/util"
	"fmt"
	"time"
)

type Operator struct {
}

var operator *Operator

func init() {
	operator = &Operator{}
}

func GetOperator() *Operator {
	return operator
}
func (op *Operator) Scale(key string) string {
	client := k8s.GetRedisClient()
	uid := util.RandomName()
	deployName := fmt.Sprintf("redis-deploy-%s", uid)
	svcName := fmt.Sprintf("redis-svc-%s", uid)

	err := op.createRedisStatefulSet(deployName)
	if err != nil {
		panic(err)
	}
	nodeport, err := op.createService(deployName, svcName)
	if err != nil {
		panic(err)
	}
	addr := fmt.Sprintf("%s:%d", "10.10.150.28", nodeport)
	global.RedisConfig[key] = map[string]string{
		"redisHost":     addr,
		"redisPassword": "",
	}
	op.waitReady(deployName)
	client.SetRedis()
	return addr
}

func (op *Operator) waitReady(name string) {
	client := k8s.GetK8sClient()
	for {
		// 获取最新的 StatefulSet 状态
		ss, err := client.GetStatefulSet(name)
		if err != nil {
			panic(err.Error())
		}
		// 检查 StatefulSet 的状态
		if ss.Status.ReadyReplicas == *ss.Spec.Replicas {
			fmt.Println("StatefulSet is ready")
			break
		}

		time.Sleep(1 * time.Second)
	}
}
