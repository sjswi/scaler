package config

import "time"

type ClusterConfig struct {
	Source          string   `json:"source"`
	Replica         []string `json:"replica"`
	ElasticReplica  []string `json:"elasticReplica"`
	ReplicaWeight   []int    `json:"replicaWeight"`
	ElasticInstance map[string]*Instance
	ServerIds       []int
	Name            string
}
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
}

const (
	NodeIP              string = "10.10.150.24"
	PersistentInstance  int64  = 3
	CostPerMinute       int64  = 10
	DatabaseKey         string = "dapr-tmysql-1"
	MySQLReader                = "reader"
	MySQLReaderPassword        = "123456"
	RedisConfigKey             = "statestore-redis"
)

type Instance struct {
	Name          string
	CreateTime    time.Time
	CostPerMinute int64
	NodePort      int
	DSP           string
}

type RedisInstance struct {
	Name          string
	CreateTime    time.Time
	CostPerMinute int64
	NodePort      int
	Addr          string
}

type ScaleRule struct {
	Type string `json:"type"` // 支持执行时间、负载并发数。后续可支持：CPU、Memory等等
	Min  int64  `json:"min"`  // 最小值，最大值
	Max  int64  `json:"max"`
}
type DatabaseConfig struct {
	Strategy          string                    `json:"strategy"`
	ClusterConnConfig map[string]*ClusterConfig `json:"cluster"`
	ScaleFactor       float64                   `json:"scaleFactor"`
	ScaleRuleType     string                    `json:"scaleRuleType"`
	Min               float64                   `json:"min"` // 最小值，当指标小于该值时触发副本实例减小
	Max               float64                   `json:"max"` // 最大值，当指标大于该值时触发副本实例增大
	TimeInterval      int                       `json:"timeInterval"`
}
