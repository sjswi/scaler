package controller

import (
	"conserver/pkg/config"
	"conserver/pkg/global"
	"conserver/pkg/mysql"
	"conserver/pkg/redis"
	"conserver/pkg/util"
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"log"
	"net/http"
)

// scaleHandler 处理scaler请求的函数
func scaleMySQLHandler(writer http.ResponseWriter, request *http.Request) {
	// 处理请求的逻辑

	var scaleReq *ScaleRequest
	all, err2 := io.ReadAll(request.Body)
	if err2 != nil {
		panic(err2)
	}
	err2 = json.Unmarshal(all, &scaleReq)
	if err2 != nil {
		panic(err2)
	}
	resp := new(ScaleResponse)
	resp.Key = scaleReq.Key
	//
	if s := util.GetTableName(scaleReq.TableName); s == "" {
		if scaleReq.Key == "" {
			// 创建主数据库
			key := mysql.GetOperator().NewMaster()
			resp.Key = key
			resp.ReplicaDSPs = global.DbConfig.ClusterConnConfig[key].Replica
			resp.SourceDSP = global.DbConfig.ClusterConnConfig[key].Source

			util.SetTableName(scaleReq.TableName, key)
		} else {
			if _, ok := global.DbConfig.ClusterConnConfig[scaleReq.Key]; ok {
				// 伸缩从数据库

				if scaleReq.ScaleType == "up" {
					dsps := make([]string, 0)
					for i := 0; i < scaleReq.InstanceNum-len(global.DbConfig.ClusterConnConfig[scaleReq.Key].ElasticReplica); i++ {
						fmt.Printf("%s扩容%d个实例", scaleReq.Key, scaleReq.InstanceNum)
						dsp := mysql.GetOperator().ScaleUp(scaleReq.Key)
						dsps = append(dsps, dsp)
					}
					resp.ElasticReplicaDSPs = dsps

				} else if scaleReq.ScaleType == "down" {
					for i := 0; i < len(global.DbConfig.ClusterConnConfig[scaleReq.Key].ElasticReplica)-scaleReq.InstanceNum; i++ {
						fmt.Printf("%s缩容到%d个实例", scaleReq.Key, scaleReq.InstanceNum)
						mysql.GetOperator().ScaleDown(scaleReq.Key)
					}

				}
			} else {
				panic("key not exists")
			}
		}
	} else {
		resp.ReplicaDSPs = global.DbConfig.ClusterConnConfig[s].Replica
		resp.SourceDSP = global.DbConfig.ClusterConnConfig[s].Source
	}

	marshal, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	writer.Write(marshal)
	UpdateComponent()

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
func scaleRedisHandler(writer http.ResponseWriter, request *http.Request) {
	var scaleReq *ScaleRequest
	all, err2 := io.ReadAll(request.Body)
	if err2 != nil {
		panic(err2)
	}
	err2 = json.Unmarshal(all, &scaleReq)
	if err2 != nil {
		panic(err2)
	}
	resp := new(ScaleRedisResponse)
	if len(global.RedisConfig) < scaleReq.InstanceNum {
		for i := 0; i < scaleReq.InstanceNum-len(global.RedisConfig); i++ {
			redis.GetOperator().Scale()
		}
	} else {
		for i := 0; i < len(global.RedisConfig)-scaleReq.InstanceNum; i++ {
			redis.GetOperator().Remove(scaleReq.Key)
		}
	}
	resp.Key = scaleReq.Key
	resp.Addr = "addr"
	marshal, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	writer.Write(marshal)
}
