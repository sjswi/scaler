package controller

import (
	"conserver/pkg/global"
	"encoding/json"
	"fmt"
	"net/http"
)

func RunController() {

	http.HandleFunc("/mysql-scale", scaleMySQLHandler)
	http.HandleFunc("/redis-scale", scaleRedisHandler)
	http.HandleFunc("/expenses", expensesHandler)
	http.HandleFunc("/mysql-count", func(writer http.ResponseWriter, request *http.Request) {
		var count struct {
			Count int `json:"count"`
		}
		count.Count = len(global.DbConfig.ClusterConnConfig)
		marshal, _ := json.Marshal(count)
		writer.Write(marshal)
	})
	http.HandleFunc("/redis-count", func(writer http.ResponseWriter, request *http.Request) {
		var count struct {
			Count int `json:"count"`
		}
		count.Count = len(global.RedisConfig)
		marshal, _ := json.Marshal(count)
		writer.Write(marshal)
	})
	http.HandleFunc("/tableToClusterName", tableHandler)
	// 启动HTTP服务器并指定监听地址和端口
	err1 := http.ListenAndServe(":18080", nil)
	if err1 != nil {
		fmt.Println("Error starting the server:", err1)
	} else {
		fmt.Println("Server is running on :8080")
	}
}
