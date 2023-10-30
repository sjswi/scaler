package main

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"testing"
)

func TestViper(t *testing.T) {
	viper.SetConfigType("yaml")
	viper.SetConfigFile("/Users/yu/.dapr/components/tmysql.yaml")
	// 寻找配置文件并读取
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	//fmt.Println(viper.Get("mysql"))     // map[port:3306 url:127.0.0.1]
	fmt.Println(viper.Get("spec.metadata")) // 127.0.0.1
	var metadataItems []map[string]string
	if err := viper.UnmarshalKey("spec.metadata", &metadataItems); err != nil {
		log.Fatalf("Unable to unmarshal spec.metadata, %v", err)
	}
	for _, item := range metadataItems {
		itemName := item["name"]
		itemValue := item["value"]
		fmt.Printf("Item Name: %s, Item Value: %v\n", itemName, itemValue)
	}
}
