package redis

import (
	"conserver/pkg/config"
	"conserver/pkg/global"
	"conserver/pkg/util"
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
func (op *Operator) Scale() string {
	pool := GetInstancePool()
	instance := pool.GetInstance()
	if instance == nil {
		return ""
	}
	global.RedisConfig[instance.Name] = map[string]string{
		"redisHost":     instance.Addr,
		"redisPassword": "AxzqDapr2023",
	}
	util.SetRedis()
	log.Default().Printf("增加一个redis实例，key：%s, addr：%s", instance.Name, instance.Addr)
	return instance.Addr
}

func (op *Operator) Remove(name string) {
	if _, ok := global.RedisConfig[name]; !ok {
		return
	}
	rConfig := global.RedisConfig[name]
	delete(global.RedisConfig, name)
	util.SetRedis()
	pool := GetInstancePool()
	instance := &config.RedisInstance{
		Name:          name,
		CreateTime:    time.Time{},
		CostPerMinute: 0,
		Password:      rConfig["redisPassword"],
		NodePort:      0,
		Addr:          rConfig["redisHost"],
	}
	pool.recycle(instance)
	log.Default().Printf("删除一个redis实例，key：%s, addr：%s", instance.Name, instance.Addr)
}
