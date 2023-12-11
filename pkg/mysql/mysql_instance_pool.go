package mysql

import (
	"conserver/pkg/config"
	"conserver/pkg/k8s"
	"conserver/pkg/util"
	"fmt"
	"sync"
	"time"
)

//	type Pool interface {
//		GetInstance()
//	}
type InstancePoolConfig struct {
	PoolSize     int      `json:"poolSize"`
	InitInstance []string `json:"initInstance"`
}

// MySQLInstancePool 存储空的MySQL实例
type MySQLInstancePool struct {
	pools    []*config.Instance
	poolSize int
	mutex    sync.Mutex
}

var instancePool *MySQLInstancePool

func init() {
	instancePool = &MySQLInstancePool{
		pools:    nil,
		poolSize: 0,
	}
}

func (m *MySQLInstancePool) GetInstance() *config.Instance {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	defer func() {
		m.pools = m.pools[1:]
	}()
	if len(m.pools) > 0 {
		return m.pools[0]
	}

	return m.newInstance()
}

func (m *MySQLInstancePool) Reallocate(instance *config.Instance) {
	m.Delete(instance.Name)
}

func GetInstancePool() *MySQLInstancePool {
	return instancePool
}

func (m *MySQLInstancePool) Init(conf *InstancePoolConfig) {
	m.poolSize = conf.PoolSize
	m.pools = make([]*config.Instance, len(conf.InitInstance))
	for i := 0; i < len(conf.InitInstance); i++ {
		uid := util.RandomName()
		m.pools[i] = &config.Instance{
			Name:          uid,
			CreateTime:    time.Time{},
			CostPerMinute: 0,
			NodePort:      0,
		}
	}
	go m.daemon()
}

func (m *MySQLInstancePool) daemon() {
	ticker := time.NewTicker(2 * time.Second)
	for {
		<-ticker.C

		if len(m.pools) > m.poolSize {
			m.mutex.Lock()

			m.Delete(m.pools[0].Name)
			m.mutex.Unlock()
		} else if len(m.pools) < m.poolSize {
			m.mutex.Lock()
			m.newInstance()
			m.mutex.Unlock()

		}
	}
}

func (m *MySQLInstancePool) newInstance() *config.Instance {
	// 创建一个从数据库，通过mysqldump加载主数据库数据，然后通过exec在pod中运行configmap加载的脚本将实例作为主数据库的slave

	// uuid 生成一个唯一的标识符 uid，确保资源名称的唯一性。
	uid := util.RandomName()

	deployName := fmt.Sprintf("mysql-deploy-%s", uid)

	cmName := fmt.Sprintf("mysql-cm-%s", uid)
	secretName := fmt.Sprintf("mysql-secret-%s", uid[:6])
	svcName := fmt.Sprintf("mysql-svc-%s", uid[:6])
	dbName := "db_test"
	//s := strings.Split(masterEndpoint, ":")
	op := GetOperator()
	err := op.createSecret(secretName, dbName)
	if err != nil {
		panic(err)
	}
	err = op.createDBConfigMap(cmName, util.NewServerID([]int{2, 3}, 2))
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
	dsp := fmt.Sprintf("root:123456@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", "10.10.150.24", nodeport, dbName)
	op.waitReady(deployName)

	//global.DbConfig.ClusterConnConfig[key].Replica = append(global.DbConfig.ClusterConnConfig[key].Replica, dsp)
	//global.DbConfig.ClusterConnConfig[key].ReplicaWeight = append(global.DbConfig.ClusterConnConfig[key].ReplicaWeight, 1)

	newInstance := &config.Instance{
		Name:          uid,
		CreateTime:    time.Now(),
		CostPerMinute: 0,
		NodePort:      int(nodeport),
		DSP:           dsp,
	}

	return newInstance
}

func (m *MySQLInstancePool) Delete(name string) {

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

	fmt.Println("MySQL resources deleted successfully")
}
