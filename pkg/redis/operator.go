package redis

import (
	"conserver/pkg/config"
	"conserver/pkg/global"
	"context"
	"encoding/json"
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
	updateConfig()
	log.Default().Printf("增加一个redis实例，key：%s, addr：%s", instance.Name, instance.Addr)
	return instance.Addr
}

func (op *Operator) Remove(name string) {
	if _, ok := global.RedisConfig[name]; !ok {
		return
	}
	rConfig := global.RedisConfig[name]
	delete(global.RedisConfig, name)
	updateConfig()
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

func updateConfig() {
	var r2 struct {
		Config map[string]map[string]string `json:"config"`
	}
	r2.Config = global.RedisConfig
	bytes, err := json.Marshal(r2)
	if err != nil {
		panic(err)
	}
	fmt.Printf("更新redis配置信息，新配置：%v, 当前时间：%v\n", r2.Config, time.Now())
	err = global.ConfigClient.Set(context.TODO(), config.RedisConfigKey, string(bytes), -1).Err()
	if err != nil {
		panic(err)
	}

}
