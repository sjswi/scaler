package k8s

import (
	"conserver/pkg/config"
	"conserver/pkg/global"
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type RedisClient struct {
	client *redis.Client
}

var redisClient *RedisClient

func init() {
	redisClient = newRedisClient()
}
func GetRedisClient() *RedisClient {
	return redisClient
}

func newRedisClient() *RedisClient {
	c := redis.NewClient(&redis.Options{
		Addr:     "10.10.150.20:30003",
		Password: "123456",
		DB:       2,
	})
	return &RedisClient{client: c}
}

func (r *RedisClient) Set(metadata []map[string]string) {
	//fmt.Println(viper.Get("mysql"))     // map[port:3306 url:127.0.0.1]
	var db *config.DatabaseConfig
	url := ""

	for _, item := range metadata {
		if item["name"] == "url" {
			err := json.Unmarshal([]byte(item["value"]), &db)
			if err != nil {
				panic(err)
			}
			url = item["value"]
		}

	}
	global.DbConfig = db
	if url != "" {
		err := r.client.Set(context.TODO(), config.DatabaseKey, url, -1).Err()
		if err != nil {
			log.Fatalf("%s 更新配置失败.\n", viper.GetString("metadata.name"))
		}
	}

}
func (r *RedisClient) SetRedis() {
	var r2 struct {
		Config map[string]map[string]string `json:"config"`
	}
	r2.Config = global.RedisConfig
	bytes, err := json.Marshal(r2)
	if err != nil {
		panic(err)
	}

	err = r.client.Set(context.TODO(), config.RedisConfigKey, string(bytes), -1).Err()
	if err != nil {
		panic(err)
	}

}
