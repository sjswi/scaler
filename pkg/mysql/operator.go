package mysql

import (
	"conserver/pkg/config"
	"conserver/pkg/global"
	"fmt"
	"math"
	"time"
)

type Operator struct {
}

var operator *Operator

func (op *Operator) ScaleUp(key string) string {
	// 创建一个从数据库，通过mysqldump加载主数据库数据，然后通过exec在pod中运行configmap加载的脚本将实例作为主数据库的slave
	pools := GetInstancePool()
	instance := pools.GetInstance()

	global.DbConfig.ClusterConnConfig[key].ElasticReplica = append(global.DbConfig.ClusterConnConfig[key].ElasticReplica, instance.DSP)
	return instance.DSP
}
func init() {
	operator = &Operator{}
}

func GetOperator() *Operator {
	return operator
}

func (op *Operator) ScaleDown(key string) {
	if len(global.DbConfig.ClusterConnConfig[key].ElasticInstance) == 0 {
		return
	}

	name := ""

	for k, v := range global.DbConfig.ClusterConnConfig[key].ElasticInstance {
		name = k
		global.CurrentFees += config.CostPerMinute * int64(math.Ceil(time.Since(v.CreateTime).Minutes()))
		delete(global.DbConfig.ClusterConnConfig[key].ElasticInstance, name)

		break
	}
	deployName := fmt.Sprintf("mysql-deploy-%s", name)

	dsp := global.DbConfig.ClusterConnConfig[key].ElasticInstance[deployName].DSP
	for i := 0; i < len(global.DbConfig.ClusterConnConfig[key].ElasticReplica); i++ {
		if global.DbConfig.ClusterConnConfig[key].ElasticReplica[i] == dsp {
			global.DbConfig.ClusterConnConfig[key].ElasticReplica = append(global.DbConfig.ClusterConnConfig[key].ElasticReplica[:i], global.DbConfig.ClusterConnConfig[key].ElasticReplica[i+1:]...)
			break
		}
	}
	delete(global.DbConfig.ClusterConnConfig[key].ElasticInstance, deployName)
	fmt.Println("MySQL resources deleted successfully")

}

func (op *Operator) NewMaster() string {
	pool := GetClusterPool()
	cluster := pool.GetCluster()
	if _, ok := global.DbConfig.ClusterConnConfig[cluster.Name]; !ok {
		global.DbConfig.ClusterConnConfig[cluster.Name] = cluster
	}

	return cluster.Name
}
