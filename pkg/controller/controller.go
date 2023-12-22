package controller

import (
	"conserver/pkg/global"
	"encoding/json"
	"fmt"
	"net/http"
)

// 设置CORS的函数
func setCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")                                                                                   // 这里设置为*表示允许任何域的请求，出于安全考虑，应该设置为具体的域名
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")                                                    // 允许的HTTP方法
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization") // 你希望支持的头
}

func RunController() {

	http.HandleFunc("/mysql-scale", scaleMySQLHandler)
	http.HandleFunc("/redis-scale", scaleRedisHandler)
	http.HandleFunc("/expenses", expensesHandler)
	http.HandleFunc("/mysql-count", func(writer http.ResponseWriter, request *http.Request) {
		setCors(&writer)
		var count struct {
			Count int `json:"count"`
		}
		count.Count = len(global.DbConfig.ClusterConnConfig)
		marshal, _ := json.Marshal(count)
		writer.Write(marshal)
	})
	http.HandleFunc("/redis-count", func(writer http.ResponseWriter, request *http.Request) {
		setCors(&writer)
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
