package controller

import (
	"conserver/pkg/global"
	"conserver/pkg/k8s"
	"conserver/pkg/mysql"
	"conserver/pkg/redis"
	"encoding/json"
	"fmt"
	"io"
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
	//

	resp.Key = scaleReq.Key
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
		// 创建主数据库
		sdsp, rdsps := mysql.GetOperator().NewMaster(scaleReq.Key)
		resp.ReplicaDSPs = rdsps
		resp.SourceDSP = sdsp
	}
	marshal, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	writer.Write(marshal)
	k8s.GetK8sClient().UpdateComponent()

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
	//
	addr := redis.GetOperator().Scale(scaleReq.Key)

	resp.Key = scaleReq.Key
	resp.Addr = addr
	marshal, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	writer.Write(marshal)

}
