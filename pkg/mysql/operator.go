package mysql

import (
	"conserver/pkg/config"
	"conserver/pkg/global"
	"conserver/pkg/k8s"
	"conserver/pkg/util"
	"fmt"
	"math"
	"time"
)

type Operator struct {
}

var operator *Operator

func (op *Operator) ScaleUp(key string) string {
	// 创建一个从数据库，通过mysqldump加载主数据库数据，然后通过exec在pod中运行configmap加载的脚本将实例作为主数据库的slave

	// uuid 生成一个唯一的标识符 uid，确保资源名称的唯一性。
	uid := util.RandomName()

	deployName := fmt.Sprintf("mysql-deploy-%s", uid)

	cmName := fmt.Sprintf("mysql-cm-%s", uid)
	secretName := fmt.Sprintf("mysql-secret-%s", uid[:6])
	svcName := fmt.Sprintf("mysql-svc-%s", uid[:6])
	dbName := "db_test"
	//s := strings.Split(masterEndpoint, ":")

	err := op.createSecret(secretName, dbName)
	if err != nil {
		panic(err)
	}
	err = op.createDBConfigMap(cmName, util.NewServerID(global.DbConfig.ClusterConnConfig[key].ServerIds, len(global.DbConfig.ClusterConnConfig[key].ServerIds)))
	if err != nil {
		panic(err)
	}
	err = op.createStatefulSet(deployName, secretName, cmName)
	if err != nil {
		panic(err)
	}
	fmt.Println(uid[:6])
	nodeport, err := op.createService(deployName, svcName)
	if err != nil {
		panic(err)
	}
	dsp := fmt.Sprintf("root:123456@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", "10.10.150.28", nodeport, dbName)
	op.waitReady(deployName)
	op.setup(deployName, key, "slave")
	//global.DbConfig.ClusterConnConfig[key].Replica = append(global.DbConfig.ClusterConnConfig[key].Replica, dsp)
	//global.DbConfig.ClusterConnConfig[key].ReplicaWeight = append(global.DbConfig.ClusterConnConfig[key].ReplicaWeight, 1)
	global.DbConfig.ClusterConnConfig[key].ElasticInstance[deployName] = &config.Instance{
		Name:          deployName,
		CreateTime:    time.Now(),
		CostPerMinute: 3,
		NodePort:      int(nodeport),
	}
	global.DbConfig.ClusterConnConfig[key].ElasticReplica = append(global.DbConfig.ClusterConnConfig[key].ElasticReplica, dsp)
	return dsp
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
	pvcBackupName := fmt.Sprintf("mysql-pvc-backup-%s", name)
	cmName := fmt.Sprintf("mysql-cm-%s", name)
	initCmName := fmt.Sprintf("mysql-cm-init-%s", name)
	secretName := fmt.Sprintf("mysql-secret-%s", name)
	svcName := fmt.Sprintf("mysql-svc-%s", name)
	client := k8s.GetK8sClient()
	// 删除 StatefulSet
	if err := client.DeleteStatefulSet(deployName); err != nil {
		fmt.Printf("Error deleting StatefulSet: %v\n", err)
		// Optionally handle the error, e.g., log it, return it, etc.
	}

	// 删除 Service
	if err := client.DeleteService(svcName); err != nil {
		fmt.Printf("Error deleting Service: %v\n", err)
		// Optionally handle the error
	}

	// 删除 PVCs
	if err := client.DeletePVC(pvcBackupName); err != nil {
		fmt.Printf("Error deleting PVC (backup): %v\n", err)
		// Optionally handle the error
	}

	// 删除 ConfigMaps
	if err := client.DeleteConfigMap(cmName); err != nil {
		fmt.Printf("Error deleting ConfigMap: %v\n", err)
		// Optionally handle the error
	}
	if err := client.DeleteConfigMap(initCmName); err != nil {
		fmt.Printf("Error deleting DB ConfigMap: %v\n", err)
		// Optionally handle the error
	}

	// 删除 Secret
	if err := client.DeleteSecret(secretName); err != nil {
		fmt.Printf("Error deleting Secret: %v\n", err)
		// Optionally handle the error
	}

	dsp := fmt.Sprintf("root:123456@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", "10.10.150.28", global.DbConfig.ClusterConnConfig[key].ElasticInstance[deployName].NodePort, "db_test")
	for i := 0; i < len(global.DbConfig.ClusterConnConfig[key].ElasticReplica); i++ {
		if global.DbConfig.ClusterConnConfig[key].ElasticReplica[i] == dsp {
			global.DbConfig.ClusterConnConfig[key].ElasticReplica = append(global.DbConfig.ClusterConnConfig[key].ElasticReplica[:i], global.DbConfig.ClusterConnConfig[key].ElasticReplica[i+1:]...)
			break
		}
	}
	delete(global.DbConfig.ClusterConnConfig[key].ElasticInstance, deployName)
	fmt.Println("MySQL resources deleted successfully")

}

func (op *Operator) NewMaster(key string) (string, []string) {
	// uuid 生成一个唯一的标识符 uid，确保资源名称的唯一性。
	uid := util.RandomName()
	deployName := fmt.Sprintf("mysql-deploy-%s", uid)
	cmName := fmt.Sprintf("mysql-cm-%s", uid)
	secretName := fmt.Sprintf("mysql-secret-%s", uid)
	svcName := fmt.Sprintf("mysql-svc-%s", uid)
	dbName := "db_test"

	err := op.createSecret(secretName, dbName)
	if err != nil {
		panic(err)
	}
	err = op.createDBConfigMap(cmName, 1)
	if err != nil {
		panic(err)
	}
	err = op.createStatefulSet(deployName, secretName, cmName)
	if err != nil {
		panic(err)
	}
	fmt.Println(uid[:6])
	nodeport, err := op.createService(deployName, svcName)
	if err != nil {
		panic(err)
	}
	dsp := fmt.Sprintf("root:123456@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", "10.10.150.28", nodeport, dbName)
	op.waitReady(deployName)
	global.DbConfig.ClusterConnConfig[key] = &config.ClusterConfig{
		Source:          dsp,
		Replica:         []string{dsp},
		ElasticInstance: map[string]*config.Instance{},
		ElasticReplica:  make([]string, 0),
		ServerIds:       []int{1},
		ReplicaWeight:   []int{1},
	}
	//dsp1 := scaleUp(Clientset, key, fmt.Sprintf("%s:%d", "10.10.150.28", nodeport))
	op.setup(deployName, key, "master")

	return dsp, []string{dsp}
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

func (op *Operator) setup(name, key, s string) {
	host, port := util.ParseDSP(global.DbConfig.ClusterConnConfig[key].Source)
	podName := util.GetPodName(name)
	client := k8s.GetK8sClient()

	if s == "slave" {
		// 设置从数据库
		file, pos := client.DumpData(podName, port, host)
		client.LoadData(podName)
		client.StartSlave(podName, port, host, file, pos)
	} else if s == "master" {
		//设置主数据库
		client.CreateReader(podName)
	}
}
