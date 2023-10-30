package main

import (
	"context"
	"flag"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
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

func createStatefulSet(clientset *kubernetes.Clientset, name, secretName, cmName, pvcBackupName, pvcDBName, dbName, initCmName string) error {
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
								fmt.Sprintf("mysqldump -h 10.10.150.28 -u root -P 30487 -p123456 %s > /backup/02-load-data.sql", dbName),
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
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcDBName,
								},
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
func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	uid := uuid.NewUUID()
	deployName := fmt.Sprintf("mysql-deploy-%s", uid[:6])
	pvcBackupName := fmt.Sprintf("mysql-pvc-backup-%s", uid[:6])
	pvcDBName := fmt.Sprintf("mysql-pvc-db-%s", uid[:6])
	cmName := fmt.Sprintf("mysql-cm-%s", uid[:6])
	initCmName := fmt.Sprintf("mysql-cm-init-%s", uid[:6])
	secretName := fmt.Sprintf("mysql-secret-%s", uid[:6])
	dbName := "db_test"
	err = createSecret(clientset, secretName, dbName)
	if err != nil {
		panic(err)
	}
	err = createPVC(clientset, pvcBackupName)
	if err != nil {
		panic(err)
	}
	err = createPVC(clientset, pvcDBName)
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
	err = createStatefulSet(clientset, deployName, secretName, cmName, pvcBackupName, pvcDBName, dbName, initCmName)
	if err != nil {
		panic(err)
	}
	fmt.Println(uid[:6])
}

func int32Ptr(i int32) *int32 { return &i }

func stringPtr(s string) *string { return &s }
