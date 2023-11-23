package controller

type ScaleRequest struct {
	Key         string `json:"key"`
	ScaleType   string `json:"scaleType"`
	InstanceNum int    `json:"instanceNum"`
}

type NameToClusterRequest struct {
	TableName string `json:"tableName"`
}
