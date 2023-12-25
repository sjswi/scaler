package redis

import (
	"conserver/pkg/global"
	"conserver/pkg/util"
	"log"
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
	pool := GetInstancePool()
	instance := pool.GetInstance()
	if instance == nil {
		return ""
	}
	global.RedisConfig[key] = map[string]string{
		"redisHost":     instance.Addr,
		"redisPassword": "AxzqDapr2023",
	}
	util.SetRedis()
	log.Default().Printf("增加一个redis实例，key：%s, addr：%s", key, instance.Addr)
	return instance.Addr
}
