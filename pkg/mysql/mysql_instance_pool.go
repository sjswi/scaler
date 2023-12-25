package mysql

import (
	"conserver/pkg/config"
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
	//go m.daemon()
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

	return nil
}

func (m *MySQLInstancePool) Delete(name string) {

	fmt.Println("MySQL resources deleted successfully")
}
