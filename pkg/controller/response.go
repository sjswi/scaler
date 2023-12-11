package controller

type ScaleResponse struct {
	Key                string   `json:"key"`
	SourceDSP          string   `json:"sourceDSP"`
	ReplicaDSPs        []string `json:"replicaDSPs"`
	ElasticReplicaDSPs []string `json:"elasticReplicaDSPs"`
	Weights            []int    `json:"weights"`
}
type NameToClusterResponse struct {
	Status      string   `json:"status"`
	ClusterName string   `json:"clusterName"`
	SourceDSP   string   `json:"sourceDSP"`
	ReplicaDSPs []string `json:"replicaDSPs"`
}

type ScaleRedisResponse struct {
	Key  string `json:"key"`
	Addr string `json:"addr"`
}
