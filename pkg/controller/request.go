package controller

type ScaleRequest struct {
	Key         string `json:"key"`
	ScaleType   string `json:"scaleType"`
	InstanceNum int    `json:"instanceNum"`
	TableName   string `json:"tableName"`
}

type NameToClusterRequest struct {
	TableName string `json:"tableName"`
}
