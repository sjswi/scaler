package redis

import (
	"conserver/pkg/config"
	"sync"
	"time"
)

type InstancePoolConfig struct {
	PoolSize     int                     `json:"poolSize"`
	InitInstance []*config.RedisInstance `json:"initInstance"`
}

// MySQLInstancePool 存储空的MySQL实例
type RedisInstancePool struct {
	pools    []*config.RedisInstance
	poolSize int
	mutex    sync.Mutex
}

var instancePool *RedisInstancePool

func init() {
	instancePool = &RedisInstancePool{
		pools:    nil,
		poolSize: 0,
		mutex:    sync.Mutex{},
	}
}

func (m *RedisInstancePool) GetInstance() *config.RedisInstance {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.pools) > 0 {
		in := m.pools[0]
		m.pools = m.pools[1:]
		return in
	} else {
		return nil
	}
}

func GetInstancePool() *RedisInstancePool {
	return instancePool
}

func (m *RedisInstancePool) Init(conf *InstancePoolConfig) {
	m.poolSize = conf.PoolSize
	m.pools = make([]*config.RedisInstance, len(conf.InitInstance))
	for i := 0; i < len(conf.InitInstance); i++ {
		m.pools[i] = conf.InitInstance[i]
	}
	//go m.daemon()
}

func (m *RedisInstancePool) daemon() {
	ticker := time.NewTicker(2 * time.Second)
	for {
		<-ticker.C

		if len(m.pools) < m.poolSize {
			m.mutex.Lock()
			instance := m.newInstance()
			m.pools = append(m.pools, instance)
			m.mutex.Unlock()

		}
	}
}

func (m *RedisInstancePool) newInstance() *config.RedisInstance {

	return nil
}

func (m *RedisInstancePool) recycle(instance *config.RedisInstance) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.pools = append(m.pools, instance)
}
