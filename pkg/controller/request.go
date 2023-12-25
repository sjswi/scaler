package controller

type ScaleRequest struct {
	Key         string `json:"key"`
	ScaleType   string `json:"scaleType"` //redis伸缩忽略
	InstanceNum int    `json:"instanceNum"`
	TableName   string `json:"tableName"` // redis伸缩忽略
}

type NameToClusterRequest struct {
	TableName string `json:"tableName"`
}
