package controller

import (
	"conserver/pkg/global"
	"conserver/pkg/mysql"
	"conserver/pkg/util"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

var mutex sync.Mutex

func init() {
	mutex = sync.Mutex{}
}
func tableHandler(writer http.ResponseWriter, request *http.Request) {
	resp := new(NameToClusterResponse)
	req := new(NameToClusterRequest)
	bytes, err := io.ReadAll(request.Body)
	if err != nil {
		resp.Status = "error"

	}
	mutex.Lock()
	defer mutex.Unlock()
	json.Unmarshal(bytes, &req)
	if err != nil {
		panic(err)
	}
	name := util.GetTableName(req.TableName)
	if name != "" {
		resp.Status = "success"
		resp.ClusterName = name
		resp.ReplicaDSPs = global.DbConfig.ClusterConnConfig[name].Replica
		resp.SourceDSP = global.DbConfig.ClusterConnConfig[name].Source
	} else {
		key := mysql.GetOperator().NewMaster()
		resp.ClusterName = key
		resp.ReplicaDSPs = global.DbConfig.ClusterConnConfig[key].Replica
		resp.SourceDSP = global.DbConfig.ClusterConnConfig[key].Source
		resp.Status = "success"
		util.SetTableName(req.TableName, key)
	}
	fmt.Printf("表名：%s，集群名：%s，映射：%s\n", req.TableName, resp.ClusterName, util.GetTableName(req.TableName))
	marshal, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	writer.Write(marshal)

}
