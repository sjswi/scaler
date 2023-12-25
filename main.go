package main

import (
	"conserver/pkg/config"
	"conserver/pkg/controller"
	"conserver/pkg/global"
	"conserver/pkg/mysql"
	"conserver/pkg/redis"
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"time"
)

func main() {
	global.DbConfig = &config.DatabaseConfig{
		Strategy:          "table",
		ClusterConnConfig: make(map[string]*config.ClusterConfig),
		ScaleFactor:       0.1,
		ScaleRuleType:     "",
		Min:               0,
		Max:               0,
		TimeInterval:      0,
	}
	viper.SetConfigFile("./pool.yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	nodeports := viper.GetStringSlice("cluster.ports")
	user := viper.GetString("cluster.user")
	password := viper.GetString("cluster.password")
	global.ConfigHost = viper.GetString("configHost")
	confs := make([]config.ClusterConfig, len(nodeports))
	for i, v := range nodeports {
		dsp := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", user, password, v, "db_test")
		confs[i] = config.ClusterConfig{
			Source:          dsp,
			Replica:         []string{dsp},
			ElasticReplica:  nil,
			ReplicaWeight:   nil,
			ElasticInstance: nil,
			ServerIds:       nil,
			Name:            fmt.Sprintf("cluster-%s", v),
		}

	}

	global.RedisConfig = make(map[string]map[string]string)

	//mysql.GetInstancePool().Init(&mysql.InstancePoolConfig{
	//	PoolSize:     0,
	//	InitInstance: []string{},
	//})
	mysql.GetClusterPool().Init(&mysql.ClusterPoolConfig{
		PoolSize:    len(nodeports),
		InitCluster: confs,
	})

	rNodeports := viper.GetStringSlice("redis.ports")
	rConfs := make([]*config.RedisInstance, len(nodeports))
	for i, v := range rNodeports {
		addr := fmt.Sprintf("%s", v)
		rConfs[i] = &config.RedisInstance{
			Name:          fmt.Sprintf("redis-%d", i),
			CreateTime:    time.Now(),
			CostPerMinute: 0,
			NodePort:      0,
			Password:      "AxzqDapr2023",
			Addr:          addr,
		}

	}

	redis.GetInstancePool().Init(&redis.InstancePoolConfig{
		PoolSize:     len(rNodeports),
		InitInstance: rConfs,
	})
	global.ScalerStartTime = time.Now()
	// 解析命令行参数
	flag.Parse()
	controller.RunController()
	fmt.Println("Asadsass")
}
