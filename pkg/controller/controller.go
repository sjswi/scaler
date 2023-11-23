package controller

import (
	"fmt"
	"net/http"
)

func RunController() {

	http.HandleFunc("/mysql-scale", scaleMySQLHandler)
	http.HandleFunc("/redis-scale", scaleRedisHandler)
	http.HandleFunc("/expenses", expensesHandler)

	http.HandleFunc("/tableToClusterName", tableHandler)
	// 启动HTTP服务器并指定监听地址和端口
	err1 := http.ListenAndServe(":18080", nil)
	if err1 != nil {
		fmt.Println("Error starting the server:", err1)
	} else {
		fmt.Println("Server is running on :8080")
	}
}
