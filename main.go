package main

import (
	"context"
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"time"
)

type DBConfig struct {
	Strategy        string   `json:"strategy"`
	Source          []string `json:"source"`
	Replica         []string `json:"replica"`
	SourceWeight    []int    `json:"sourceWeight"`
	ReplicaWeight   []int    `json:"replicaWeight"`
	elasticInstance map[string]*Instance
}
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Key      string `json:"key"`
}

const (
	NodeIP             string = "10.10.150.28"
	PersistentInstance int64  = 3
	CostPerMinute      int64  = 10
)

type Instance struct {
	Name          string
	CreateTime    time.Time
	CostPerMinute int64
	NodePort      int
}

var (
	dbConfig        map[string]*DBConfig
	CurrentFees     int64 = 0
	Clientset       *kubernetes.Clientset
	scalerStartTime time.Time = time.Now()
)

func main() {
	dbConfig = make(map[string]*DBConfig)
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	// 解析命令行参数
	flag.Parse()
	// 使用kubeconfig文件创建一个配置
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	// 创建Kubernetes客户端
	Clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	go initKubernetesMode()
	start()

}

func convertMap(input map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range input {
		if strValue, ok := v.(string); ok {
			result[k] = strValue
		}
	}
	return result
}

func initKubernetesMode() {
	kubeconfig := "/Users/yu/.kube/config" // 例如: "~/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	componentGVR := schema.GroupVersionResource{
		Group:    "dapr.io",
		Version:  "v1alpha1",
		Resource: "components",
	}

	informer := cache.NewSharedInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return dynamicClient.Resource(componentGVR).Namespace("dapr-yxb").List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return dynamicClient.Resource(componentGVR).Namespace("dapr-yxb").Watch(context.TODO(), options)
			},
		},
		&unstructured.Unstructured{},
		0, //Skip resync
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if u, ok := obj.(*unstructured.Unstructured); ok {
				fmt.Printf("Component Updated: %s\n", u.GetName())
				if spec, ok := u.Object["spec"].(map[string]interface{}); ok {
					if spec["type"] == "bindings.tmysql" {
						if metadataItemsInterface, ok := spec["metadata"].([]interface{}); ok {
							metadataItems := make([]map[string]string, len(metadataItemsInterface))
							for i, item := range metadataItemsInterface {
								if itemMap, ok := item.(map[string]interface{}); ok {
									metadataItems[i] = convertMap(itemMap)
								}
							}
							set(metadataItems)
						}
						fmt.Printf("metadata : %v", spec["metadata"])
					}
				}

				// 如果你需要比较新旧对象，你可以同时转换oldObj
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if u, ok := newObj.(*unstructured.Unstructured); ok {
				fmt.Printf("Component Updated: %s\n", u.GetName())
				if spec, ok := u.Object["spec"].(map[string]interface{}); ok {
					if spec["type"] == "bindings.tmysql" {
						if metadataItemsInterface, ok := spec["metadata"].([]interface{}); ok {
							metadataItems := make([]map[string]string, len(metadataItemsInterface))
							for i, item := range metadataItemsInterface {
								if itemMap, ok := item.(map[string]interface{}); ok {
									metadataItems[i] = convertMap(itemMap)
								}
							}

							set(metadataItems)
						}
						fmt.Printf("metadata : %v", spec["metadata"])
					}
				}

				// 如果你需要比较新旧对象，你可以同时转换oldObj
			}
		},
		DeleteFunc: func(obj interface{}) {
			if u, ok := obj.(*unstructured.Unstructured); ok {
				fmt.Printf("Component Deleted: %s\n", u.GetName())
			}
		},
	})

	stopper := make(chan struct{})
	defer close(stopper)
	informer.Run(stopper)
}
