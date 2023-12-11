package mysql

import (
	"conserver/pkg/config"
	"conserver/pkg/k8s"
	"conserver/pkg/util"
	"fmt"
	"k8s.io/apimachinery/pkg/util/rand"
	"sync"
	"time"
)

type ClusterPoolConfig struct {
	PoolSize    int                    `json:"poolSize"`
	InitCluster []config.ClusterConfig `json:"initCluster"`
}

// MySQLClusterPool 存储MySQL集群的配置，
type MySQLClusterPool struct {
	pools    []*config.ClusterConfig
	poolSize int
	mutex    sync.Mutex
}

var clusterPool *MySQLClusterPool

func (m *MySQLClusterPool) GetCluster() *config.ClusterConfig {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if len(m.pools) > 0 {
		intn := rand.Intn(m.poolSize)
		fmt.Printf("从集群池中取出集群：%s\n", m.pools[intn].Name)
		return m.pools[intn]
	}

	return m.newCluster()
}

func (m *MySQLClusterPool) Reallocate(instance *config.Instance) {

}
func init() {
	clusterPool = &MySQLClusterPool{
		pools:    nil,
		poolSize: 0,
	}
}
func GetClusterPool() *MySQLClusterPool {
	return clusterPool
}

func (m *MySQLClusterPool) Init(conf *ClusterPoolConfig) {
	m.poolSize = conf.PoolSize
	m.pools = make([]*config.ClusterConfig, len(conf.InitCluster))
	for i := 0; i < len(conf.InitCluster); i++ {

		m.pools[i] = &conf.InitCluster[i]
	}
	go m.daemon()
}

func (m *MySQLClusterPool) daemon() {
	ticker := time.NewTicker(2 * time.Second)
	for {
		<-ticker.C

		if len(m.pools) < m.poolSize {
			m.mutex.Lock()
			cluster := m.newCluster()
			m.pools = append(m.pools, cluster)
			m.mutex.Unlock()
		}
	}
}

// newCluster 创建一个一主一从的数据库集群
func (m *MySQLClusterPool) newCluster() *config.ClusterConfig {
	// uuid 生成一个唯一的标识符 uid，确保资源名称的唯一性。
	uid := util.RandomName()
	deployName := fmt.Sprintf("mysql-deploy-%s", uid)
	cmName := fmt.Sprintf("mysql-cm-%s", uid)
	secretName := fmt.Sprintf("mysql-secret-%s", uid)
	svcName := fmt.Sprintf("mysql-svc-%s", uid)
	dbName := "db_test"
	op := GetOperator()
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
	dsp := fmt.Sprintf("root:123456@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", "10.10.150.24", nodeport, dbName)
	op.waitReady(deployName)

	op.setup(deployName, "", "master")
	instance := GetInstancePool().GetInstance()
	op.setup(instance.Name, dsp, "slave")

	return &config.ClusterConfig{
		Source:          dsp,
		Replica:         []string{dsp, instance.DSP},
		ElasticReplica:  nil,
		ReplicaWeight:   []int{1, 1},
		ElasticInstance: nil,
		ServerIds:       []int{1},
		Name:            uid,
	}
}

// randomDelete 随机删除一个集群
func (m *MySQLClusterPool) Delete(name string) {

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
