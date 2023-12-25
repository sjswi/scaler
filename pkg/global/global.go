package global

import (
	"conserver/pkg/config"
	"conserver/pkg/util"
	"github.com/go-redis/redis/v8"
	"time"
)

var (
	DbConfig               *config.DatabaseConfig
	CurrentFees            int64 = 0
	RedisConfig            map[string]map[string]string
	TableNameToClusterName *util.ConsistentCache
	ScalerStartTime        time.Time
	ConfigHost             string
	ConfigClient           *redis.Client
)
