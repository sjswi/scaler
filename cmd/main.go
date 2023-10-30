package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var rdb *redis.Client

type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Key      string `json:"key"`
}

func readConfigsFromDir(dir string) map[string]*viper.Viper {
	configs := make(map[string]*viper.Viper)

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("Failed to read directory: %s", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
			v := viper.New()
			v.SetConfigFile(dir + "/" + filename)
			if err := v.ReadInConfig(); err != nil {
				log.Printf("Error reading config file %s: %s", filename, err)
				continue
			}
			if v.GetString("spec.type") == "bindings.tmysql" {
				configs[filename] = v
			}

		}
	}

	return configs
}
func initSelfHostMode() {
	yamls := readConfigsFromDir("/Users/yu/.dapr/components")
	for _, v := range yamls {
		var metadataItems []map[string]string
		if err := v.UnmarshalKey("spec.metadata", &metadataItems); err != nil {
			log.Fatalf("Unable to unmarshal spec.metadata, %v", err)
		}
		set(metadataItems)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					v := viper.New()
					v.SetConfigType("yaml")
					v.SetConfigFile(event.Name)
					// 寻找配置文件并读取
					err := v.ReadInConfig()
					if err != nil {
						panic(fmt.Errorf("fatal error config file: %w", err))
					}
					var metadataItems []map[string]string
					if err := v.UnmarshalKey("spec.metadata", &metadataItems); err != nil {
						log.Fatalf("Unable to unmarshal spec.metadata, %v", err)
					}
					set(metadataItems)
					fmt.Printf("%s 更新了", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("/Users/yu/.dapr/components")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", "10.10.150.20", 30003), // Redis 地址
		Password: "123456",                                    // Redis 密码，没有则留空
		DB:       1,                                           // 使用的 Redis 数据库编号
	})
	rdb.FlushDB(context.TODO())

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
func set(metadataItems []map[string]string) {
	//fmt.Println(viper.Get("mysql"))     // map[port:3306 url:127.0.0.1]

	url := ""
	key := ""
	for _, item := range metadataItems {
		if item["name"] == "url" {
			url = item["value"]
		}
		if item["name"] == "redisConfig" {
			var rConfig *RedisConfig
			err := json.Unmarshal([]byte(item["value"]), &rConfig)
			if err != nil {
				break
			}
			key = rConfig.Key
		}
	}
	if url != "" && key != "" {
		err := rdb.Set(context.TODO(), key, url, -1).Err()
		if err != nil {
			log.Fatalf("%s 更新配置失败.\n", viper.GetString("metadata.name"))
		}
	}
}

func main() {

	args := os.Args
	fmt.Println("Program name:", args[0])

	http.HandleFunc("/scale", scaleHandler)
	if len(args) != 2 {
		fmt.Println("参数多余两个")
		return
	}
	if args[1] == "kubernetes" {
		go initKubernetesMode()
	} else if args[1] == "host" {
		go initSelfHostMode()
	} else {
		fmt.Println("只支持kubernetes和host模式")
		return
	}
	http.ListenAndServe(":8080", nil)
}

type ScaleRequest struct {
	MasterHost string `json:"masterHost"`
}

func scaleHandler(writer http.ResponseWriter, request *http.Request) {
	all, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(404)
		writer.Write([]byte("读取body失败"))

	}
	var scaleReq *ScaleRequest
	err = json.Unmarshal(all, &scaleReq)
	if err != nil {
		writer.WriteHeader(404)
		writer.Write([]byte("参数解析失败"))

	}

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// 使用kubeconfig文件创建一个配置
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// 创建Kubernetes客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// 定义Deployment资源
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:1.14.2",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	// 使用clientset创建Deployment
	deploymentsClient := clientset.AppsV1().Deployments(corev1.NamespaceDefault)
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
}
func int32Ptr(i int32) *int32 {
	return &i
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
