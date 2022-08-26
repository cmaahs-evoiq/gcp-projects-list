package main

type GcpProjects struct {
	Projects []GcpProjectsProject `json:"projects"`
}

type GcpProjectsProject struct {
	CreateTime  string                   `json:"createTime"`
	DisplayName string                   `json:"displayName"`
	Etag        string                   `json:"etag"`
	Labels      GcpProjectsProjectLabels `json:"labels"`
	Name        string                   `json:"name"`
	Parent      string                   `json:"parent"`
	ProjectID   string                   `json:"projectId"`
	State       string                   `json:"state"`
	UpdateTime  string                   `json:"updateTime"`
}

type GcpProjectsProjectLabels struct {
	Costalloc  string `json:"costalloc"`
	Department string `json:"department"`
}
