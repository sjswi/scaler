package mysql

import (
	"conserver/pkg/k8s"
	"fmt"
)

func (op *Operator) createDBConfigMap(name string, serverID int) error {
	configData := map[string]string{
		"my.cnf": fmt.Sprintf(
			`
[mysqld]
# Your MySQL configuration here
default_authentication_plugin=mysql_native_password
server-id=%d
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

`, serverID),
		"dump.sh": `
   #!/bin/bash -e
    # MySQL 主数据库的相关信息
    MASTER_HOST=$1
    MASTER_USER="root"
    MASTER_PASSWORD="123456"
    MASTER_PORT=$2
    # 设置主数据库为只读
    mysql -h "$MASTER_HOST" -P "$MASTER_PORT" -u "$MASTER_USER" -p"$MASTER_PASSWORD" -e "SET GLOBAL read_only = ON;"

    # 导出数据库
    mysqldump -h "$MASTER_HOST" -P "$MASTER_PORT" -u "$MASTER_USER" -p"$MASTER_PASSWORD" --all-databases --master-data=2 > /etc/all_databases.sql

    # 获取MASTER_LOG_FILE和Log Position
    cat /etc/all_databases.sql | grep "CHANGE MASTER TO MASTER_LOG_FILE" | awk -F"[\047= ]+" '{print "Log File: "$6", Log Position: "$9}'

    # 设置主数据库为可写
    mysql -h "$MASTER_HOST" -P "$MASTER_PORT" -u "$MASTER_USER" -p"$MASTER_PASSWORD" -e "SET GLOBAL read_only = OFF;"
`,
		"create_reader.sh": `
#!/bin/bash -e

# 替换为您的实际用户名、密码和主机
REPLICATION_USER="reader"
REPLICATION_PASSWORD="123456"
MYSQL_ROOT_PASSWORD="123456"
MYSQL_HOST="localhost"
MYSQL_PORT="3306"

# 创建复制用户并授予权限
mysql -u root -p"$MYSQL_ROOT_PASSWORD" -h "$MYSQL_HOST" -P "$MYSQL_PORT" -e "CREATE USER '$REPLICATION_USER'@'%' IDENTIFIED BY '$REPLICATION_PASSWORD';"
mysql -u root -p"$MYSQL_ROOT_PASSWORD" -h "$MYSQL_HOST" -P "$MYSQL_PORT" -e "GRANT REPLICATION SLAVE ON *.* TO '$REPLICATION_USER'@'%';"
mysql -u root -p"$MYSQL_ROOT_PASSWORD" -h "$MYSQL_HOST" -P "$MYSQL_PORT" -e "FLUSH PRIVILEGES;"
`,
		"start_slave.sh": `
#!/bin/bash -e

# 主数据库信息 - 需要替换为实际值
MASTER_HOST=$1
MASTER_PORT=$2
MASTER_USER="reader"
MASTER_PASSWORD="123456"
MASTER_LOG_FILE=$3 # 这些值应从主数据库获得
MASTER_LOG_POS=$4

# 从数据库登录信息 - 需要替换为实际值
SLAVE_USER="root"
SLAVE_PASSWORD="123456"
SLAVE_HOST="localhost"

# 在从数据库上配置复制
mysql -u "$SLAVE_USER" -p"$SLAVE_PASSWORD" -h "$SLAVE_HOST" -e "CHANGE MASTER TO MASTER_HOST='$MASTER_HOST', MASTER_PORT=$MASTER_PORT, MASTER_USER='$MASTER_USER', MASTER_PASSWORD='$MASTER_PASSWORD', MASTER_LOG_FILE='$MASTER_LOG_FILE', MASTER_LOG_POS=$MASTER_LOG_POS;"
mysql -u "$SLAVE_USER" -p"$SLAVE_PASSWORD" -h "$SLAVE_HOST" -e "START SLAVE;"

`,
		"load.sh": `
#!/bin/bash -e

# MySQL数据库的相关信息
MYSQL_USER="root" # 替换为您的MySQL用户名
MYSQL_PASSWORD="123456" # 替换为您的MySQL密码
MYSQL_HOST="localhost" # 或者替换为您的MySQL服务器地址
MYSQL_PORT=3306 # 根据实际情况进行替换

# SQL文件路径
SQL_FILE="/etc/all_databases.sql" # 确保这是正确的路径到您的SQL文件

# 检查SQL文件是否存在
if [ ! -f "$SQL_FILE" ]; then
    echo "SQL文件未找到: $SQL_FILE"
    exit 1
fi

# 导入数据库
echo "开始导入数据库..."
mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" < "$SQL_FILE"
echo "数据库导入完成。"



`,
	}

	client := k8s.GetK8sClient()
	err := client.CreateConfigMap(name, configData)
	if err != nil {
		return err
	}
	return nil
}
