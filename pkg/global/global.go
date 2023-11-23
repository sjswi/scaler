package global

import (
	"conserver/pkg/config"
	"conserver/pkg/util"
	"time"
)

var (
	DbConfig               *config.DatabaseConfig
	CurrentFees            int64 = 0
	RedisConfig            map[string]map[string]string
	TableNameToClusterName *util.ConsistentCache
	ScalerStartTime        time.Time
)
