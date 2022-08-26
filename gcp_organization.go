package main

type GcpOrganization struct {
	Organizations []GcpOrganizationOrganization `json:"organizations"`
}

type GcpOrganizationOrganization struct {
	CreationTime   string                           `json:"creationTime"`
	DisplayName    string                           `json:"displayName"`
	LifecycleState string                           `json:"lifecycleState"`
	Name           string                           `json:"name"`
	OrganizationID string                           `json:"organizationId"`
	Owner          GcpOrganizationOrganizationOwner `json:"owner"`
}

type GcpOrganizationOrganizationOwner struct {
	DirectoryCustomerID string `json:"directoryCustomerId"`
}
