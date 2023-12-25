package util

import (
	"conserver/pkg/config"
	"conserver/pkg/global"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/uuid"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"
)

func Int32Ptr(i int32) *int32 { return &i }

func StringPtr(s string) *string { return &s }

func ConvertMap(input map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range input {
		if strValue, ok := v.(string); ok {
			result[k] = strValue
		}
	}
	return result
}

func RandomName() string {
	newUUID := uuid.NewUUID()

	return string(newUUID[:6])
}

func ParseDSP(dsp string) (string, string) {
	re := regexp.MustCompile(`@tcp\(([^)]+)\)`)
	matches := re.FindStringSubmatch(dsp)

	if len(matches) > 1 {
		hostPort := matches[1]
		split := strings.Split(hostPort, ":")
		return split[0], split[1]
	}
	panic(errors.New("dsp parse error"))
}

func GetPodName(statefulsetName string) string {
	return fmt.Sprintf("%s-0", statefulsetName)
}

func NewServerID(ids []int, n int) int {
	sort.Ints(ids)
	if len(ids) == 0 {
		return 100
	}
	return ids[len(ids)-1] + 1
}

func ParseFileAndPos(str string) (string, string) {

	re := regexp.MustCompile(`Log File: (.*), Log Position: (\d+);`)
	matches := re.FindStringSubmatch(str)

	if len(matches) >= 3 {

		return matches[1], matches[2]
	} else {
		panic(errors.New("error"))
	}
	return "", ""
}

func UpdateComponent() {
	marshal, err2 := json.Marshal(global.DbConfig)
	if err2 != nil {
		panic(err2)
	}

	err := global.ConfigClient.Set(context.TODO(), config.DatabaseKey, string(marshal), -1).Err()
	if err != nil {
		log.Fatalf("%s 更新配置失败.\n", viper.GetString("metadata.name"))
	}

}

func SetRedis() {
	var r2 struct {
		Config map[string]map[string]string `json:"config"`
	}
	r2.Config = global.RedisConfig
	bytes, err := json.Marshal(r2)
	if err != nil {
		panic(err)
	}
	fmt.Printf("更新redis配置信息，新配置：%v, 当前时间：%v\n", r2.Config, time.Now())
	err = global.ConfigClient.Set(context.TODO(), config.RedisConfigKey, string(bytes), -1).Err()
	if err != nil {
		panic(err)
	}

}
