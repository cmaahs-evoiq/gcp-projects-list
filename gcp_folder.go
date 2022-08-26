package main

type GcpFolders struct {
	Folders []GcpFolder `json:"folders"`
}

type GcpFolder struct {
	CreateTime     string `json:"createTime"`
	DisplayName    string `json:"displayName"`
	LifecycleState string `json:"lifecycleState"`
	Name           string `json:"name"`
	Parent         string `json:"parent"`
}
