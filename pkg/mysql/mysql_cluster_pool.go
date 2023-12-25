package mysql

import (
	"conserver/pkg/config"
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
	//go m.daemon()
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

	return nil
}

// randomDelete 随机删除一个集群
func (m *MySQLClusterPool) Delete(name string) {

	fmt.Println("MySQL resources deleted successfully")
}
