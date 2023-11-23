package controller

import (
	"conserver/pkg/config"
	"conserver/pkg/global"
	"fmt"
	"math"
	"net/http"
	"time"
)

func expensesHandler(writer http.ResponseWriter, request *http.Request) {
	fee := global.CurrentFees + 3*config.CostPerMinute*int64(math.Ceil(time.Since(global.ScalerStartTime).Minutes()))

	for _, v := range global.DbConfig.ClusterConnConfig {
		for _, v2 := range v.ElasticInstance {
			fee += config.CostPerMinute * int64(math.Ceil(time.Since(v2.CreateTime).Minutes()))
		}
	}
	fmt.Fprintf(writer, "当前总消费：%d元\n", fee)
}
