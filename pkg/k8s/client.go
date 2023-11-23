package k8s

import (
	"conserver/pkg/global"
	"conserver/pkg/util"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

type K8sClient struct {
	client        *kubernetes.Clientset
	dynamicClient *dynamic.DynamicClient
	config        *rest.Config
}

var client *K8sClient

func (c *K8sClient) Listen() {
	componentGVR := schema.GroupVersionResource{
		Group:    "dapr.io",
		Version:  "v1alpha1",
		Resource: "components",
	}

	informer := cache.NewSharedInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return c.dynamicClient.Resource(componentGVR).Namespace("dapr-yxb").List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return c.dynamicClient.Resource(componentGVR).Namespace("dapr-yxb").Watch(context.TODO(), options)
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
									metadataItems[i] = util.ConvertMap(itemMap)
								}
							}
							GetRedisClient().Set(metadataItems)
						}
						fmt.Printf("metadata : %v", spec["metadata"])
					}
				}
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
									metadataItems[i] = util.ConvertMap(itemMap)
								}
							}

							GetRedisClient().Set(metadataItems)
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

func (c *K8sClient) UpdateComponent() {
	marshal, err2 := json.Marshal(global.DbConfig)
	if err2 != nil {
		panic(err2)
	}

	componentGVR := schema.GroupVersionResource{
		Group:    "dapr.io",
		Version:  "v1alpha1",
		Resource: "components",
	}

	namespace := "dapr-yxb"        // 使用适当的命名空间
	componentName := "dapr-tmysql" // Dapr Component 的名称
	// 获取当前的 Component
	component, err := c.dynamicClient.Resource(componentGVR).Namespace(namespace).Get(context.TODO(), componentName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	// 修改 Component 的 metadata 中 name 为 url 的字段
	if metadata, found, _ := unstructured.NestedSlice(component.Object, "spec", "metadata"); found {
		for _, item := range metadata {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if name, ok := itemMap["name"]; ok && name == "url" {
					// 更新 url 的 value
					itemMap["value"] = string(marshal) // 这里替换为新的 value
					break
				}
			}
		}
		unstructured.SetNestedSlice(component.Object, metadata, "spec", "metadata")
	}

	// 提交更新
	_, err = c.dynamicClient.Resource(componentGVR).Namespace(namespace).Update(context.TODO(), component, metav1.UpdateOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Component updated successfully")
}

func newK8sClient() *K8sClient {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	// 使用kubeconfig文件创建一个配置
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	// 创建Kubernetes客户端
	clientSet, err := kubernetes.NewForConfig(config)
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return &K8sClient{client: clientSet, dynamicClient: dynamicClient, config: config}
}

func init() {
	client = newK8sClient()
}

func GetK8sClient() *K8sClient {
	return client
}
