package redis

import (
	"conserver/pkg/global"
	"conserver/pkg/k8s"
	"fmt"
	"log"
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
	pool := GetInstancePool()
	instance := pool.GetInstance()
	global.RedisConfig[key] = map[string]string{
		"redisHost":     instance.Addr,
		"redisPassword": "",
	}
	client.SetRedis()
	log.Default().Printf("增加一个redis实例，key：%s, addr：%s", key, instance.Addr)
	return instance.Addr
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
			fmt.Println("Redis StatefulSet is ready")
			break
		}

		time.Sleep(1 * time.Second)
	}
}
