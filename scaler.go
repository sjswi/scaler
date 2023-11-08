package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"math"
	"net/http"
	"time"
)

func createPVC(clientset *kubernetes.Clientset, name string) error {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
			// 如果你有特定的StorageClass，可以指定它
			StorageClassName: stringPtr("rook-cephfs"),
		},
	}

	_, err := clientset.CoreV1().PersistentVolumeClaims("default").Create(context.TODO(), pvc, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func createConfigMap(clientset *kubernetes.Clientset, name string) error {
	configData := map[string]string{
		"my.cnf": `
[mysqld]
# Your MySQL configuration here
default_authentication_plugin=mysql_native_password
server-id=1
## 开启binlog
log-bin=mysql-bin
## binlog缓存
binlog_cache_size=1M
## binlog格式(mixed、statement、row,默认格式是statement)
binlog_format=mixed
##设置字符编码为utf8mb4
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci
init_connect='SET NAMES utf8mb4'
[client]
default-character-set = utf8mb4
[mysql]
default-character-set = utf8mb4

`,
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: configData,
	}

	_, err := clientset.CoreV1().ConfigMaps("default").Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func createDBConfigMap(clientset *kubernetes.Clientset, name string, dbName string) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string]string{
			"01-create-database.sql": fmt.Sprintf("CREATE DATABASE test1;"), // 创建数据库的SQL内容
		},
	}

	_, err := clientset.CoreV1().ConfigMaps("default").Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func createSecret(clientset *kubernetes.Clientset, name, dbName string) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"password": []byte("123456"),
			"database": []byte(dbName),
		},
	}

	_, err := clientset.CoreV1().Secrets("default").Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func createStatefulSet(clientset *kubernetes.Clientset, name, secretName, cmName, pvcBackupName, dbName, initCmName string) error {
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:  "mysqldump",
							Image: "mysql:8.1",
							Command: []string{
								"sh",
								"-c",
								fmt.Sprintf("mysqldump -h 10.10.150.28 -u root -P 30340 -p123456 %s > /backup/02-load-data.sql", dbName),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "backup",
									MountPath: "/backup",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "mysql",
							Image: "mysql:8.1",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "mysql-data",
									MountPath: "/var/lib/mysql",
								},
								{
									Name:      "config",
									MountPath: "/etc/mysql/conf.d", // 或你希望的其他路径
								},
								{
									Name:      "init-scripts-volume",
									MountPath: "/docker-entrypoint-initdb.d/01-create-database.sql",
									SubPath:   "01-create-database.sql",
								},
								{
									Name:      "backup",
									MountPath: "/docker-entrypoint-initdb.d/02-load-data.sql",
									SubPath:   "02-load-data.sql",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "MYSQL_ROOT_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: secretName,
											},
											Key: "password",
										},
									},
								},
								{
									Name: "MYSQL_DATABASE",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: secretName,
											},
											Key: "database",
										},
									},
								},
								// ... 其他的环境变量 ...
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "backup",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcBackupName,
								},
							},
						},
						{
							Name: "mysql-data",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cmName,
									},
								},
							},
						},
						{
							Name: "init-scripts-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: initCmName,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := clientset.AppsV1().StatefulSets("default").Create(context.TODO(), statefulSet, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
func createMySQLService(clientset *kubernetes.Clientset, statefulSetName, serviceName string) (int32, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: map[string]string{"app": statefulSetName},
			Ports: []corev1.ServicePort{
				{
					Port: 3306,
					Name: "mysql",
				},
			},
		},
	}

	svc, err := clientset.CoreV1().Services("default").Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return 0, err
	}

	// 获取分配的NodePort
	nodePort := svc.Spec.Ports[0].NodePort

	return nodePort, nil
}

func start() {

	http.HandleFunc("/scale", scaleHandler)
	http.HandleFunc("/expenses", expensesHandler)
	// 启动HTTP服务器并指定监听地址和端口
	err1 := http.ListenAndServe(":18080", nil)
	if err1 != nil {
		fmt.Println("Error starting the server:", err1)
	} else {
		fmt.Println("Server is running on :8080")
	}
}

var rdb *redis.Client

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", "10.10.150.20", 30003), // Redis 地址
		Password: "123456",                                    // Redis 密码，没有则留空
		DB:       1,                                           // 使用的 Redis 数据库编号
	})
	rdb.FlushDB(context.TODO())
}

type ScaleRequest struct {
	Key         string `json:"key"`
	ScaleUp     bool   `json:"scaleUp"`
	InstanceNum int    `json:"instanceNum"`
}

func set(metadataItems []map[string]string) {
	//fmt.Println(viper.Get("mysql"))     // map[port:3306 url:127.0.0.1]
	var db *DBConfig
	url := ""
	key := ""
	for _, item := range metadataItems {
		if item["name"] == "url" {
			err := json.Unmarshal([]byte(item["value"]), &db)
			if err != nil {
				panic(err)
			}
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
	db.elasticInstance = make(map[string]*Instance)
	if url != "" && key != "" {
		dbConfig[key] = db
		err := rdb.Set(context.TODO(), key, url, -1).Err()
		if err != nil {
			log.Fatalf("%s 更新配置失败.\n", viper.GetString("metadata.name"))
		}
	}
}

func scaleUp(clientset *kubernetes.Clientset, key string) {
	// uuid 生成一个唯一的标识符 uid，确保资源名称的唯一性。
	uid := uuid.NewUUID()

	deployName := fmt.Sprintf("mysql-deploy-%s", uid[:6])
	pvcBackupName := fmt.Sprintf("mysql-pvc-backup-%s", uid[:6])
	cmName := fmt.Sprintf("mysql-cm-%s", uid[:6])
	initCmName := fmt.Sprintf("mysql-cm-init-%s", uid[:6])
	secretName := fmt.Sprintf("mysql-secret-%s", uid[:6])
	svcName := fmt.Sprintf("mysql-svc-%s", uid[:6])
	dbName := "db_test"

	err := createSecret(clientset, secretName, dbName)
	if err != nil {
		panic(err)
	}
	err = createPVC(clientset, pvcBackupName)
	if err != nil {
		panic(err)
	}
	err = createConfigMap(clientset, cmName)
	if err != nil {
		panic(err)
	}
	err = createDBConfigMap(clientset, initCmName, dbName)
	if err != nil {
		panic(err)
	}
	err = createStatefulSet(clientset, deployName, secretName, cmName, pvcBackupName, dbName, initCmName)
	if err != nil {
		panic(err)
	}
	fmt.Println(uid[:6])
	nodeport, err := createMySQLService(clientset, deployName, svcName)
	if err != nil {
		panic(err)
	}
	dbConfig[key].Replica = append(dbConfig[key].Replica, fmt.Sprintf("root:123456@(%s:%d)/db_test", NodeIP, nodeport))
	dbConfig[key].ReplicaWeight = append(dbConfig[key].ReplicaWeight, 1)
	newInstance := &Instance{
		Name:          string(uid[:6]),
		CreateTime:    time.Now(),
		CostPerMinute: CostPerMinute,
		NodePort:      int(nodeport),
	}
	dbConfig[key].elasticInstance[string(uid[:6])] = newInstance
	updateConfiguration(key)
}
func scaleDown(clientset *kubernetes.Clientset, key string) {
	if len(dbConfig[key].elasticInstance) == 0 {
		return
	}

	name := ""
	port := 1
	for k, v := range dbConfig[key].elasticInstance {
		name = k
		port = v.NodePort
		CurrentFees += CostPerMinute * int64(math.Ceil(time.Since(v.CreateTime).Minutes()))
		delete(dbConfig[key].elasticInstance, name)

		break
	}
	deployName := fmt.Sprintf("mysql-deploy-%s", name)
	pvcBackupName := fmt.Sprintf("mysql-pvc-backup-%s", name)
	cmName := fmt.Sprintf("mysql-cm-%s", name)
	initCmName := fmt.Sprintf("mysql-cm-init-%s", name)
	secretName := fmt.Sprintf("mysql-secret-%s", name)
	svcName := fmt.Sprintf("mysql-svc-%s", name)
	// 删除 StatefulSet
	if err := clientset.AppsV1().StatefulSets("default").Delete(context.Background(), deployName, metav1.DeleteOptions{}); err != nil {
		fmt.Printf("Error deleting StatefulSet: %v\n", err)
		// Optionally handle the error, e.g., log it, return it, etc.
	}

	// 删除 Service
	if err := clientset.CoreV1().Services("default").Delete(context.Background(), svcName, metav1.DeleteOptions{}); err != nil {
		fmt.Printf("Error deleting Service: %v\n", err)
		// Optionally handle the error
	}

	// 删除 PVCs
	if err := clientset.CoreV1().PersistentVolumeClaims("default").Delete(context.Background(), pvcBackupName, metav1.DeleteOptions{}); err != nil {
		fmt.Printf("Error deleting PVC (backup): %v\n", err)
		// Optionally handle the error
	}

	// 删除 ConfigMaps
	if err := clientset.CoreV1().ConfigMaps("default").Delete(context.Background(), cmName, metav1.DeleteOptions{}); err != nil {
		fmt.Printf("Error deleting ConfigMap: %v\n", err)
		// Optionally handle the error
	}
	if err := clientset.CoreV1().ConfigMaps("default").Delete(context.Background(), initCmName, metav1.DeleteOptions{}); err != nil {
		fmt.Printf("Error deleting DB ConfigMap: %v\n", err)
		// Optionally handle the error
	}

	// 删除 Secret
	if err := clientset.CoreV1().Secrets("default").Delete(context.Background(), secretName, metav1.DeleteOptions{}); err != nil {
		fmt.Printf("Error deleting Secret: %v\n", err)
		// Optionally handle the error
	}

	fmt.Println("MySQL resources deleted successfully")
	dsp := fmt.Sprintf("root:123456@(%s:%d)/db_test", NodeIP, port)
	for i, v := range dbConfig[key].Replica {
		if dsp == v {
			dbConfig[key].Replica = append(dbConfig[key].Replica[:i], dbConfig[key].Replica[i+1:]...)
			dbConfig[key].ReplicaWeight = append(dbConfig[key].ReplicaWeight[:i], dbConfig[key].ReplicaWeight[i+1:]...)
			break
		}
	}
	updateConfiguration(key)
}

// 处理scaler请求的函数
func scaleHandler(writer http.ResponseWriter, request *http.Request) {
	// 处理请求的逻辑
	fmt.Fprintln(writer, "Hello, World!!!!!!!!!!") // 向客户端发送响应
	var scaleReq *ScaleRequest
	all, err2 := io.ReadAll(request.Body)
	if err2 != nil {
		panic(err2)
	}
	err2 = json.Unmarshal(all, &scaleReq)
	if err2 != nil {
		panic(err2)
	}

	if _, ok := dbConfig[scaleReq.Key]; !ok {
		fmt.Printf("扩缩容失败，不存在key为%s的数据库", scaleReq.Key)
		return
	}
	if scaleReq.ScaleUp {
		for i := 0; i < scaleReq.InstanceNum; i++ {
			fmt.Printf("%s扩容%d个实例", scaleReq.Key, scaleReq.InstanceNum)
			scaleUp(Clientset, scaleReq.Key)
		}

	} else {
		for i := 0; i < scaleReq.InstanceNum; i++ {
			fmt.Printf("%s缩容%d个实例", scaleReq.Key, scaleReq.InstanceNum)
			scaleDown(Clientset, scaleReq.Key)
		}

	}

}

func expensesHandler(writer http.ResponseWriter, request *http.Request) {
	fee := CurrentFees + 3*CostPerMinute*int64(math.Ceil(time.Since(scalerStartTime).Minutes()))

	for _, v := range dbConfig {
		for _, v2 := range v.elasticInstance {
			fee += CostPerMinute * int64(math.Ceil(time.Since(v2.CreateTime).Minutes()))
		}
	}
	fmt.Fprintf(writer, "当前总消费：%d元\n", fee)
}

func updateConfiguration(key string) {
	marshal, err := json.Marshal(dbConfig[key])
	if err != nil {
		panic(err)
	}
	err = rdb.Set(context.TODO(), key, string(marshal), -1).Err()
	if err != nil {
		log.Fatalf("%s 更新配置失败.\n", viper.GetString("metadata.name"))
	}

}

func int32Ptr(i int32) *int32 { return &i }

func stringPtr(s string) *string { return &s }
