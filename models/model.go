package models

import (
	"bytes"
	"golang.org/x/crypto/ssh"
	"net"
)

type ResouceType string

const (
	Cluster ResouceType = "cluster"
)

type Type string

const (
	New                      Type = "New"
	Created                  Type = "created"
	ClusterCreated           Type = "Cluster Created"
	ClusterCreationFailed    Type = "Cluster Creation Failed"
	ClusterTerminationFailed Type = "Cluster Termination Failed"
	ClusterTerminated        Type = "Cluster Terminated"
	Deploying                Type = "Creating"
	Terminating              Type = "Terminating"
	ClusterUpdateFailed      Type = "Cluster Update Failed"
)

type StatusCode int

const (
	CloudStatusCode     StatusCode = 512
	ParamMissing        StatusCode = 404
	StateConflict       StatusCode = 409
	Unauthorized        StatusCode = 401
	InternalServerError StatusCode = 500
	BadRequest          StatusCode = 400
	Conflict            StatusCode = 409
	NotFound            StatusCode = 404
)

type ErrorMessage string

const (
	IsEmpty                ErrorMessage = "is empty"
	Notauthorized          ErrorMessage = "User is unauthorized to perform this action"
	AlreadyExist           ErrorMessage = "Cluster against same project id already exists"
	SuccessfullyAdded      ErrorMessage = "Cluster added successfully"
	SuccessfullyUpdated    ErrorMessage = "Cluster updated successfully"
	SuccessfullyDeleted    ErrorMessage = "Cluster deleted successfully"
	CreationInitialised    ErrorMessage = "Cluster creation initiated"
	TerminationInitialised ErrorMessage = "Cluster termination initialized"
	KeySuccessfullyDeleted ErrorMessage = "Key deleted successfully"
	KeySuccessfullyAdded   ErrorMessage = "Key added successfully"
	ValidProfile           ErrorMessage = "Profile is valid"
)

type HeaderVariable string

const (
	Token     HeaderVariable = "X-Auth-Token"
	ProfileId HeaderVariable = "X-Profile-Id"
	ProjectId HeaderVariable = "Project Id"
)

type PathVariable string

const (
	KeyName       PathVariable = "keyname"
	RegionV       PathVariable = "region"
	ClusterName   PathVariable = "clusterName"
	ResourceGroup PathVariable = "resourceGroup"
)

type RequestType string

const (
	POST RequestType = "post"
	PUT  RequestType = "put"
)

type PoolRole string

const (
	Master PoolRole = "master"
	Slave  PoolRole = "slave"
)

type Role string

const (
	SuperUser Role = "Super-User"
	Admin     Role = "Admin"
)

// cloud type of the cluster
// swagger:enum Cloud
type Cloud string

const (
	AWS   Cloud = "aws"
	Azure Cloud = "azure"
	GCP   Cloud = "gcp"
	GKE   Cloud = "gke"
	EKS   Cloud = "eks"
	DO    Cloud = "do"
	DOKS  Cloud = "doks"
	IKS   Cloud = "iks"
	OP    Cloud = "op"
	AKS   Cloud = "aks"
)

type KeyType string

const (
	NEWKey   KeyType = "new"
	CPKey    KeyType = "cp"
	AWSKey   KeyType = "aws"
	AZUREKey KeyType = "azure"
	USERKey  KeyType = "user"
)

type OsDiskType string

const (
	StandardHDD OsDiskType = "standard hdd"
	StandardSSD OsDiskType = "standard ssd"
	PremiumSSD  OsDiskType = "premium ssd"
)

type GCPDiskType string

const (
	PdStandard GCPDiskType = "pd-standard"
	PdSSD      GCPDiskType = "pd-ssd"
)

//type AKSVMType string
//
//const (
//	// AKSVMTypeStandardA1 ...
//	AKSVMTypeStandardA1 AKSVMType = "Standard_A1"
//	// AKSVMTypeStandardA10 ...
//	AKSVMTypeStandardA10 AKSVMType = "Standard_A10"
//	// AKSVMTypeStandardA11 ...
//	AKSVMTypeStandardA11 AKSVMType = "Standard_A11"
//	// AKSVMTypeStandardA1V2 ...
//	AKSVMTypeStandardA1V2 AKSVMType = "Standard_A1_v2"
//	// AKSVMTypeStandardA2 ...
//	AKSVMTypeStandardA2 AKSVMType = "Standard_A2"
//	// AKSVMTypeStandardA2mV2 ...
//	AKSVMTypeStandardA2mV2 AKSVMType = "Standard_A2m_v2"
//	// AKSVMTypeStandardA2V2 ...
//	AKSVMTypeStandardA2V2 AKSVMType = "Standard_A2_v2"
//	// AKSVMTypeStandardA3 ...
//	AKSVMTypeStandardA3 AKSVMType = "Standard_A3"
//	// AKSVMTypeStandardA4 ...
//	AKSVMTypeStandardA4 AKSVMType = "Standard_A4"
//	// AKSVMTypeStandardA4mV2 ...
//	AKSVMTypeStandardA4mV2 AKSVMType = "Standard_A4m_v2"
//	// AKSVMTypeStandardA4V2 ...
//	AKSVMTypeStandardA4V2 AKSVMType = "Standard_A4_v2"
//	// AKSVMTypeStandardA5 ...
//	AKSVMTypeStandardA5 AKSVMType = "Standard_A5"
//	// AKSVMTypeStandardA6 ...
//	AKSVMTypeStandardA6 AKSVMType = "Standard_A6"
//	// AKSVMTypeStandardA7 ...
//	AKSVMTypeStandardA7 AKSVMType = "Standard_A7"
//	// AKSVMTypeStandardA8 ...
//	AKSVMTypeStandardA8 AKSVMType = "Standard_A8"
//	// AKSVMTypeStandardA8mV2 ...
//	AKSVMTypeStandardA8mV2 AKSVMType = "Standard_A8m_v2"
//	// AKSVMTypeStandardA8V2 ...
//	AKSVMTypeStandardA8V2 AKSVMType = "Standard_A8_v2"
//	// AKSVMTypeStandardA9 ...
//	AKSVMTypeStandardA9 AKSVMType = "Standard_A9"
//	// AKSVMTypeStandardB2ms ...
//	AKSVMTypeStandardB2ms AKSVMType = "Standard_B2ms"
//	// AKSVMTypeStandardB2s ...
//	AKSVMTypeStandardB2s AKSVMType = "Standard_B2s"
//	// AKSVMTypeStandardB4ms ...
//	AKSVMTypeStandardB4ms AKSVMType = "Standard_B4ms"
//	// AKSVMTypeStandardB8ms ...
//	AKSVMTypeStandardB8ms AKSVMType = "Standard_B8ms"
//	// AKSVMTypeStandardD1 ...
//	AKSVMTypeStandardD1 AKSVMType = "Standard_D1"
//	// AKSVMTypeStandardD11 ...
//	AKSVMTypeStandardD11 AKSVMType = "Standard_D11"
//	// AKSVMTypeStandardD11V2 ...
//	AKSVMTypeStandardD11V2 AKSVMType = "Standard_D11_v2"
//	// AKSVMTypeStandardD11V2Promo ...
//	AKSVMTypeStandardD11V2Promo AKSVMType = "Standard_D11_v2_Promo"
//	// AKSVMTypeStandardD12 ...
//	AKSVMTypeStandardD12 AKSVMType = "Standard_D12"
//	// AKSVMTypeStandardD12V2 ...
//	AKSVMTypeStandardD12V2 AKSVMType = "Standard_D12_v2"
//	// AKSVMTypeStandardD12V2Promo ...
//	AKSVMTypeStandardD12V2Promo AKSVMType = "Standard_D12_v2_Promo"
//	// AKSVMTypeStandardD13 ...
//	AKSVMTypeStandardD13 AKSVMType = "Standard_D13"
//	// AKSVMTypeStandardD13V2 ...
//	AKSVMTypeStandardD13V2 AKSVMType = "Standard_D13_v2"
//	// AKSVMTypeStandardD13V2Promo ...
//	AKSVMTypeStandardD13V2Promo AKSVMType = "Standard_D13_v2_Promo"
//	// AKSVMTypeStandardD14 ...
//	AKSVMTypeStandardD14 AKSVMType = "Standard_D14"
//	// AKSVMTypeStandardD14V2 ...
//	AKSVMTypeStandardD14V2 AKSVMType = "Standard_D14_v2"
//	// AKSVMTypeStandardD14V2Promo ...
//	AKSVMTypeStandardD14V2Promo AKSVMType = "Standard_D14_v2_Promo"
//	// AKSVMTypeStandardD15V2 ...
//	AKSVMTypeStandardD15V2 AKSVMType = "Standard_D15_v2"
//	// AKSVMTypeStandardD16sV3 ...
//	AKSVMTypeStandardD16sV3 AKSVMType = "Standard_D16s_v3"
//	// AKSVMTypeStandardD16V3 ...
//	AKSVMTypeStandardD16V3 AKSVMType = "Standard_D16_v3"
//	// AKSVMTypeStandardD1V2 ...
//	AKSVMTypeStandardD1V2 AKSVMType = "Standard_D1_v2"
//	// AKSVMTypeStandardD2 ...
//	AKSVMTypeStandardD2 AKSVMType = "Standard_D2"
//	// AKSVMTypeStandardD2sV3 ...
//	AKSVMTypeStandardD2sV3 AKSVMType = "Standard_D2s_v3"
//	// AKSVMTypeStandardD2V2 ...
//	AKSVMTypeStandardD2V2 AKSVMType = "Standard_D2_v2"
//	// AKSVMTypeStandardD2V2Promo ...
//	AKSVMTypeStandardD2V2Promo AKSVMType = "Standard_D2_v2_Promo"
//	// AKSVMTypeStandardD2V3 ...
//	AKSVMTypeStandardD2V3 AKSVMType = "Standard_D2_v3"
//	// AKSVMTypeStandardD3 ...
//	AKSVMTypeStandardD3 AKSVMType = "Standard_D3"
//	// AKSVMTypeStandardD32sV3 ...
//	AKSVMTypeStandardD32sV3 AKSVMType = "Standard_D32s_v3"
//	// AKSVMTypeStandardD32V3 ...
//	AKSVMTypeStandardD32V3 AKSVMType = "Standard_D32_v3"
//	// AKSVMTypeStandardD3V2 ...
//	AKSVMTypeStandardD3V2 AKSVMType = "Standard_D3_v2"
//	// AKSVMTypeStandardD3V2Promo ...
//	AKSVMTypeStandardD3V2Promo AKSVMType = "Standard_D3_v2_Promo"
//	// AKSVMTypeStandardD4 ...
//	AKSVMTypeStandardD4 AKSVMType = "Standard_D4"
//	// AKSVMTypeStandardD4sV3 ...
//	AKSVMTypeStandardD4sV3 AKSVMType = "Standard_D4s_v3"
//	// AKSVMTypeStandardD4V2 ...
//	AKSVMTypeStandardD4V2 AKSVMType = "Standard_D4_v2"
//	// AKSVMTypeStandardD4V2Promo ...
//	AKSVMTypeStandardD4V2Promo AKSVMType = "Standard_D4_v2_Promo"
//	// AKSVMTypeStandardD4V3 ...
//	AKSVMTypeStandardD4V3 AKSVMType = "Standard_D4_v3"
//	// AKSVMTypeStandardD5V2 ...
//	AKSVMTypeStandardD5V2 AKSVMType = "Standard_D5_v2"
//	// AKSVMTypeStandardD5V2Promo ...
//	AKSVMTypeStandardD5V2Promo AKSVMType = "Standard_D5_v2_Promo"
//	// AKSVMTypeStandardD64sV3 ...
//	AKSVMTypeStandardD64sV3 AKSVMType = "Standard_D64s_v3"
//	// AKSVMTypeStandardD64V3 ...
//	AKSVMTypeStandardD64V3 AKSVMType = "Standard_D64_v3"
//	// AKSVMTypeStandardD8sV3 ...
//	AKSVMTypeStandardD8sV3 AKSVMType = "Standard_D8s_v3"
//	// AKSVMTypeStandardD8V3 ...
//	AKSVMTypeStandardD8V3 AKSVMType = "Standard_D8_v3"
//	// AKSVMTypeStandardDS1 ...
//	AKSVMTypeStandardDS1 AKSVMType = "Standard_DS1"
//	// AKSVMTypeStandardDS11 ...
//	AKSVMTypeStandardDS11 AKSVMType = "Standard_DS11"
//	// AKSVMTypeStandardDS11V2 ...
//	AKSVMTypeStandardDS11V2 AKSVMType = "Standard_DS11_v2"
//	// AKSVMTypeStandardDS11V2Promo ...
//	AKSVMTypeStandardDS11V2Promo AKSVMType = "Standard_DS11_v2_Promo"
//	// AKSVMTypeStandardDS12 ...
//	AKSVMTypeStandardDS12 AKSVMType = "Standard_DS12"
//	// AKSVMTypeStandardDS12V2 ...
//	AKSVMTypeStandardDS12V2 AKSVMType = "Standard_DS12_v2"
//	// AKSVMTypeStandardDS12V2Promo ...
//	AKSVMTypeStandardDS12V2Promo AKSVMType = "Standard_DS12_v2_Promo"
//	// AKSVMTypeStandardDS13 ...
//	AKSVMTypeStandardDS13 AKSVMType = "Standard_DS13"
//	// AKSVMTypeStandardDS132V2 ...
//	AKSVMTypeStandardDS132V2 AKSVMType = "Standard_DS13-2_v2"
//	// AKSVMTypeStandardDS134V2 ...
//	AKSVMTypeStandardDS134V2 AKSVMType = "Standard_DS13-4_v2"
//	// AKSVMTypeStandardDS13V2 ...
//	AKSVMTypeStandardDS13V2 AKSVMType = "Standard_DS13_v2"
//	// AKSVMTypeStandardDS13V2Promo ...
//	AKSVMTypeStandardDS13V2Promo AKSVMType = "Standard_DS13_v2_Promo"
//	// AKSVMTypeStandardDS14 ...
//	AKSVMTypeStandardDS14 AKSVMType = "Standard_DS14"
//	// AKSVMTypeStandardDS144V2 ...
//	AKSVMTypeStandardDS144V2 AKSVMType = "Standard_DS14-4_v2"
//	// AKSVMTypeStandardDS148V2 ...
//	AKSVMTypeStandardDS148V2 AKSVMType = "Standard_DS14-8_v2"
//	// AKSVMTypeStandardDS14V2 ...
//	AKSVMTypeStandardDS14V2 AKSVMType = "Standard_DS14_v2"
//	// AKSVMTypeStandardDS14V2Promo ...
//	AKSVMTypeStandardDS14V2Promo AKSVMType = "Standard_DS14_v2_Promo"
//	// AKSVMTypeStandardDS15V2 ...
//	AKSVMTypeStandardDS15V2 AKSVMType = "Standard_DS15_v2"
//	// AKSVMTypeStandardDS1V2 ...
//	AKSVMTypeStandardDS1V2 AKSVMType = "Standard_DS1_v2"
//	// AKSVMTypeStandardDS2 ...
//	AKSVMTypeStandardDS2 AKSVMType = "Standard_DS2"
//	// AKSVMTypeStandardDS2V2 ...
//	AKSVMTypeStandardDS2V2 AKSVMType = "Standard_DS2_v2"
//	// AKSVMTypeStandardDS2V2Promo ...
//	AKSVMTypeStandardDS2V2Promo AKSVMType = "Standard_DS2_v2_Promo"
//	// AKSVMTypeStandardDS3 ...
//	AKSVMTypeStandardDS3 AKSVMType = "Standard_DS3"
//	// AKSVMTypeStandardDS3V2 ...
//	AKSVMTypeStandardDS3V2 AKSVMType = "Standard_DS3_v2"
//	// AKSVMTypeStandardDS3V2Promo ...
//	AKSVMTypeStandardDS3V2Promo AKSVMType = "Standard_DS3_v2_Promo"
//	// AKSVMTypeStandardDS4 ...
//	AKSVMTypeStandardDS4 AKSVMType = "Standard_DS4"
//	// AKSVMTypeStandardDS4V2 ...
//	AKSVMTypeStandardDS4V2 AKSVMType = "Standard_DS4_v2"
//	// AKSVMTypeStandardDS4V2Promo ...
//	AKSVMTypeStandardDS4V2Promo AKSVMType = "Standard_DS4_v2_Promo"
//	// AKSVMTypeStandardDS5V2 ...
//	AKSVMTypeStandardDS5V2 AKSVMType = "Standard_DS5_v2"
//	// AKSVMTypeStandardDS5V2Promo ...
//	AKSVMTypeStandardDS5V2Promo AKSVMType = "Standard_DS5_v2_Promo"
//	// AKSVMTypeStandardE16sV3 ...
//	AKSVMTypeStandardE16sV3 AKSVMType = "Standard_E16s_v3"
//	// AKSVMTypeStandardE16V3 ...
//	AKSVMTypeStandardE16V3 AKSVMType = "Standard_E16_v3"
//	// AKSVMTypeStandardE2sV3 ...
//	AKSVMTypeStandardE2sV3 AKSVMType = "Standard_E2s_v3"
//	// AKSVMTypeStandardE2V3 ...
//	AKSVMTypeStandardE2V3 AKSVMType = "Standard_E2_v3"
//	// AKSVMTypeStandardE3216sV3 ...
//	AKSVMTypeStandardE3216sV3 AKSVMType = "Standard_E32-16s_v3"
//	// AKSVMTypeStandardE328sV3 ...
//	AKSVMTypeStandardE328sV3 AKSVMType = "Standard_E32-8s_v3"
//	// AKSVMTypeStandardE32sV3 ...
//	AKSVMTypeStandardE32sV3 AKSVMType = "Standard_E32s_v3"
//	// AKSVMTypeStandardE32V3 ...
//	AKSVMTypeStandardE32V3 AKSVMType = "Standard_E32_v3"
//	// AKSVMTypeStandardE4sV3 ...
//	AKSVMTypeStandardE4sV3 AKSVMType = "Standard_E4s_v3"
//	// AKSVMTypeStandardE4V3 ...
//	AKSVMTypeStandardE4V3 AKSVMType = "Standard_E4_v3"
//	// AKSVMTypeStandardE6416sV3 ...
//	AKSVMTypeStandardE6416sV3 AKSVMType = "Standard_E64-16s_v3"
//	// AKSVMTypeStandardE6432sV3 ...
//	AKSVMTypeStandardE6432sV3 AKSVMType = "Standard_E64-32s_v3"
//	// AKSVMTypeStandardE64sV3 ...
//	AKSVMTypeStandardE64sV3 AKSVMType = "Standard_E64s_v3"
//	// AKSVMTypeStandardE64V3 ...
//	AKSVMTypeStandardE64V3 AKSVMType = "Standard_E64_v3"
//	// AKSVMTypeStandardE8sV3 ...
//	AKSVMTypeStandardE8sV3 AKSVMType = "Standard_E8s_v3"
//	// AKSVMTypeStandardE8V3 ...
//	AKSVMTypeStandardE8V3 AKSVMType = "Standard_E8_v3"
//	// AKSVMTypeStandardF1 ...
//	AKSVMTypeStandardF1 AKSVMType = "Standard_F1"
//	// AKSVMTypeStandardF16 ...
//	AKSVMTypeStandardF16 AKSVMType = "Standard_F16"
//	// AKSVMTypeStandardF16s ...
//	AKSVMTypeStandardF16s AKSVMType = "Standard_F16s"
//	// AKSVMTypeStandardF16sV2 ...
//	AKSVMTypeStandardF16sV2 AKSVMType = "Standard_F16s_v2"
//	// AKSVMTypeStandardF1s ...
//	AKSVMTypeStandardF1s AKSVMType = "Standard_F1s"
//	// AKSVMTypeStandardF2 ...
//	AKSVMTypeStandardF2 AKSVMType = "Standard_F2"
//	// AKSVMTypeStandardF2s ...
//	AKSVMTypeStandardF2s AKSVMType = "Standard_F2s"
//	// AKSVMTypeStandardF2sV2 ...
//	AKSVMTypeStandardF2sV2 AKSVMType = "Standard_F2s_v2"
//	// AKSVMTypeStandardF32sV2 ...
//	AKSVMTypeStandardF32sV2 AKSVMType = "Standard_F32s_v2"
//	// AKSVMTypeStandardF4 ...
//	AKSVMTypeStandardF4 AKSVMType = "Standard_F4"
//	// AKSVMTypeStandardF4s ...
//	AKSVMTypeStandardF4s AKSVMType = "Standard_F4s"
//	// AKSVMTypeStandardF4sV2 ...
//	AKSVMTypeStandardF4sV2 AKSVMType = "Standard_F4s_v2"
//	// AKSVMTypeStandardF64sV2 ...
//	AKSVMTypeStandardF64sV2 AKSVMType = "Standard_F64s_v2"
//	// AKSVMTypeStandardF72sV2 ...
//	AKSVMTypeStandardF72sV2 AKSVMType = "Standard_F72s_v2"
//	// AKSVMTypeStandardF8 ...
//	AKSVMTypeStandardF8 AKSVMType = "Standard_F8"
//	// AKSVMTypeStandardF8s ...
//	AKSVMTypeStandardF8s AKSVMType = "Standard_F8s"
//	// AKSVMTypeStandardF8sV2 ...
//	AKSVMTypeStandardF8sV2 AKSVMType = "Standard_F8s_v2"
//	// AKSVMTypeStandardG1 ...
//	AKSVMTypeStandardG1 AKSVMType = "Standard_G1"
//	// AKSVMTypeStandardG2 ...
//	AKSVMTypeStandardG2 AKSVMType = "Standard_G2"
//	// AKSVMTypeStandardG3 ...
//	AKSVMTypeStandardG3 AKSVMType = "Standard_G3"
//	// AKSVMTypeStandardG4 ...
//	AKSVMTypeStandardG4 AKSVMType = "Standard_G4"
//	// AKSVMTypeStandardG5 ...
//	AKSVMTypeStandardG5 AKSVMType = "Standard_G5"
//	// AKSVMTypeStandardGS1 ...
//	AKSVMTypeStandardGS1 AKSVMType = "Standard_GS1"
//	// AKSVMTypeStandardGS2 ...
//	AKSVMTypeStandardGS2 AKSVMType = "Standard_GS2"
//	// AKSVMTypeStandardGS3 ...
//	AKSVMTypeStandardGS3 AKSVMType = "Standard_GS3"
//	// AKSVMTypeStandardGS4 ...
//	AKSVMTypeStandardGS4 AKSVMType = "Standard_GS4"
//	// AKSVMTypeStandardGS44 ...
//	AKSVMTypeStandardGS44 AKSVMType = "Standard_GS4-4"
//	// AKSVMTypeStandardGS48 ...
//	AKSVMTypeStandardGS48 AKSVMType = "Standard_GS4-8"
//	// AKSVMTypeStandardGS5 ...
//	AKSVMTypeStandardGS5 AKSVMType = "Standard_GS5"
//	// AKSVMTypeStandardGS516 ...
//	AKSVMTypeStandardGS516 AKSVMType = "Standard_GS5-16"
//	// AKSVMTypeStandardGS58 ...
//	AKSVMTypeStandardGS58 AKSVMType = "Standard_GS5-8"
//	// AKSVMTypeStandardH16 ...
//	AKSVMTypeStandardH16 AKSVMType = "Standard_H16"
//	// AKSVMTypeStandardH16m ...
//	AKSVMTypeStandardH16m AKSVMType = "Standard_H16m"
//	// AKSVMTypeStandardH16mr ...
//	AKSVMTypeStandardH16mr AKSVMType = "Standard_H16mr"
//	// AKSVMTypeStandardH16r ...
//	AKSVMTypeStandardH16r AKSVMType = "Standard_H16r"
//	// AKSVMTypeStandardH8 ...
//	AKSVMTypeStandardH8 AKSVMType = "Standard_H8"
//	// AKSVMTypeStandardH8m ...
//	AKSVMTypeStandardH8m AKSVMType = "Standard_H8m"
//	// AKSVMTypeStandardL16s ...
//	AKSVMTypeStandardL16s AKSVMType = "Standard_L16s"
//	// AKSVMTypeStandardL32s ...
//	AKSVMTypeStandardL32s AKSVMType = "Standard_L32s"
//	// AKSVMTypeStandardL4s ...
//	AKSVMTypeStandardL4s AKSVMType = "Standard_L4s"
//	// AKSVMTypeStandardL8s ...
//	AKSVMTypeStandardL8s AKSVMType = "Standard_L8s"
//	// AKSVMTypeStandardM12832ms ...
//	AKSVMTypeStandardM12832ms AKSVMType = "Standard_M128-32ms"
//	// AKSVMTypeStandardM12864ms ...
//	AKSVMTypeStandardM12864ms AKSVMType = "Standard_M128-64ms"
//	// AKSVMTypeStandardM128ms ...
//	AKSVMTypeStandardM128ms AKSVMType = "Standard_M128ms"
//	// AKSVMTypeStandardM128s ...
//	AKSVMTypeStandardM128s AKSVMType = "Standard_M128s"
//	// AKSVMTypeStandardM6416ms ...
//	AKSVMTypeStandardM6416ms AKSVMType = "Standard_M64-16ms"
//	// AKSVMTypeStandardM6432ms ...
//	AKSVMTypeStandardM6432ms AKSVMType = "Standard_M64-32ms"
//	// AKSVMTypeStandardM64ms ...
//	AKSVMTypeStandardM64ms AKSVMType = "Standard_M64ms"
//	// AKSVMTypeStandardM64s ...
//	AKSVMTypeStandardM64s AKSVMType = "Standard_M64s"
//	// AKSVMTypeStandardNC12 ...
//	AKSVMTypeStandardNC12 AKSVMType = "Standard_NC12"
//	// AKSVMTypeStandardNC12sV2 ...
//	AKSVMTypeStandardNC12sV2 AKSVMType = "Standard_NC12s_v2"
//	// AKSVMTypeStandardNC12sV3 ...
//	AKSVMTypeStandardNC12sV3 AKSVMType = "Standard_NC12s_v3"
//	// AKSVMTypeStandardNC24 ...
//	AKSVMTypeStandardNC24 AKSVMType = "Standard_NC24"
//	// AKSVMTypeStandardNC24r ...
//	AKSVMTypeStandardNC24r AKSVMType = "Standard_NC24r"
//	// AKSVMTypeStandardNC24rsV2 ...
//	AKSVMTypeStandardNC24rsV2 AKSVMType = "Standard_NC24rs_v2"
//	// AKSVMTypeStandardNC24rsV3 ...
//	AKSVMTypeStandardNC24rsV3 AKSVMType = "Standard_NC24rs_v3"
//	// AKSVMTypeStandardNC24sV2 ...
//	AKSVMTypeStandardNC24sV2 AKSVMType = "Standard_NC24s_v2"
//	// AKSVMTypeStandardNC24sV3 ...
//	AKSVMTypeStandardNC24sV3 AKSVMType = "Standard_NC24s_v3"
//	// AKSVMTypeStandardNC6 ...
//	AKSVMTypeStandardNC6 AKSVMType = "Standard_NC6"
//	// AKSVMTypeStandardNC6sV2 ...
//	AKSVMTypeStandardNC6sV2 AKSVMType = "Standard_NC6s_v2"
//	// AKSVMTypeStandardNC6sV3 ...
//	AKSVMTypeStandardNC6sV3 AKSVMType = "Standard_NC6s_v3"
//	// AKSVMTypeStandardND12s ...
//	AKSVMTypeStandardND12s AKSVMType = "Standard_ND12s"
//	// AKSVMTypeStandardND24rs ...
//	AKSVMTypeStandardND24rs AKSVMType = "Standard_ND24rs"
//	// AKSVMTypeStandardND24s ...
//	AKSVMTypeStandardND24s AKSVMType = "Standard_ND24s"
//	// AKSVMTypeStandardND6s ...
//	AKSVMTypeStandardND6s AKSVMType = "Standard_ND6s"
//	// AKSVMTypeStandardNV12 ...
//	AKSVMTypeStandardNV12 AKSVMType = "Standard_NV12"
//	// AKSVMTypeStandardNV24 ...
//	AKSVMTypeStandardNV24 AKSVMType = "Standard_NV24"
//	// AKSVMTypeStandardNV6 ...
//	AKSVMTypeStandardNV6 AKSVMType = "Standard_NV6"
//)

type CredentialsType string

const (
	Password CredentialsType = "password"
	SSHKey   CredentialsType = "key"
)
const (
	ProjectGetEndpoint = "/raccoon/projects/{projectId}"
	WeaselGetEndpoint  = "/weasel/network/{cloud}/{projectId}"
)
const (
	WoodpeckerEnpoint = "/agent/api/v1/clientconfig"
)
const (
	VaultEndpoint = "/robin/api/v1"

	VaultGetKeyURI     = "/template/sshKey/{cloud}/{region}/{keyName}"
	VaultGetAllKeysURI = "/template/sshKey/{cloud}/{region}"
	VaultGetProfileURI = "/template/{cloud}/credentials/{profileId}"
	VaultCreateKeyURI  = "/template/sshKey"
	VaultDeleteKeyURI  = "/template/sshKey/{cloudType}/{region}/{name}"
)
const (
	IBM_IAM_Endpoint                 = "https://iam.cloud.ibm.com/identity/token"
	IBM_Kube_Cluster_Endpoint        = "https://containers.cloud.ibm.com/global/v2/vpc/createCluster"
	IBM_Kube_GetWorker_Endpoint      = "https://containers.cloud.ibm.com/global/v2/getWorkerPools"
	IBM_Kube_GetNodes_Endpoint       = "https://containers.cloud.ibm.com/global/v2/vpc/getWorkers"
	IBM_Kube_GetCluster_Endpoint     = "https://containers.cloud.ibm.com/global/v2/getCluster"
	IBM_Kube_Delete_Cluster_Endpoint = "https://containers.cloud.ibm.com/global/v1/clusters/"
	IBM_Remove_WorkerPool            = "https://containers.cloud.ibm.com/global/v2/removeWorkerPool"
	IBM_WorkerPool_Endpoint          = "https://containers.cloud.ibm.com/global/v2/vpc/createWorkerPool"
	IBM_Zone                         = "https://containers.cloud.ibm.com/global/v2/vpc/createWorkerPoolZone"
	IBM_All_Instances_Endpoint       = "https://containers.cloud.ibm.com/global/v2/getFlavors?zone="
	IBM_ALL_Kube_Version_Endpoint    = "https://containers.cloud.ibm.com/global/v2/getVersions"
	IBM_Update_Version               = "https://containers.cloud.ibm.com/global/v2/updateMaster"
	IBM_Update_PoolSize              = "https://containers.cloud.ibm.com/global/v2/resizeWorkerPool"
	IBM_Version                      = "?version=2020-01-28&generation=1"
)
const (
	RbacEndpoint = "/security/api/rbac/"

	RbacListURI     = "list"
	RbacAccessURI   = "allowed"
	RbacEvaluateURI = "evaluate"
	RbacInfoURI     = "token/info"
	RbacPolicyURI   = "policy"
	RbacExtractURI  = "token/extract"
)

type Logger string

const (
	//////logging///////////
	Backend_Logging Logger = "backend-logging"
	Audit_Trails    Logger = "audit-trails"

	LOGGING_LEVEL_INFO    = "info"
	LOGGING_LEVEL_ERROR   = "error"
	LOGGING_LEVEL_WARNING = "warning"
)

const (
	LoggingEndpoint = "/elephant/api/v1/"

	BackEndLoggingURI    = "backend/logging"
	FrontEndLoggingURI   = "frontend/logging"
	AuditTrailLoggingURI = "audit/store"
)
const WoodPeckerCertificate = "/agent/api/v1/config/k8s/{profileId}"
const GKEAuthContainerName = "jhgke"
const AKSAuthContainerName = "jhaks"
const EKSAuthContainerName = "jheks"
const DOAuthContainerName = "jhdo"
const IBMKSAuthContainerName = "jhibmks"

type Machine struct {
	InstanceType string `json: "instanceType" `
	Cores        int64  `json: "cores" `
}
type GCPMachine struct {
	InstanceType string  `json: "instanceType" `
	Cores        float64 `json: "cores" `
}

type Limits struct {
	CoreCount      int64 `json: "CoreCount" `
	DeveloperCount int64 `json: "DeveloperCount" `
	MeshCount      int64 `json: "MeshCount" `
	MeshSize       int64 `json: "MeshSize" `
}

const (
	NETWORK_CONTRIBUTOR_GUID = "4d97b98b-1d4f-4787-a291-c67834d212e7"
	VM_CONTRIBUTOR_GUID      = "9980e02c-c2be-4d73-94e8-173b1dc7cf3c"
	STORAGE_CONTRIBUTOR_GUID = "17d1049b-9a84-46fb-8f53-869881c3d3ab"
	AVERE_CONTRIBUTER_GUID   = "4f8fab4f-1852-4a58-a46a-8eaf358af14a"
)

type Region struct {
	Name     string `json: "name" `
	Location string `json: "location" `
}

type AzureRegion struct {
	Region   string
	Location string
}
type AzureZone struct {
	Label      string 	`json:"label" bson:"label" description:"label of the zone"`
	Value      string 	`json:"value" bson:"value" description:"value of the zone"`
}
type GcpRegion struct {
	Name     string `json: "name"`
	Zone     string `json: "zone"`
	Location string `json: "location"`
}

func RemoteRun(user string, addr string, privateKey string, cmd string) (string, error) {
	//clientPem, err := ioutil.ReadFile(privateKey)
	//if err != nil {
	// return "", err
	//}
	clientPem := []byte(privateKey)
	key, err := ssh.ParsePrivateKey(clientPem)
	if err != nil {
		return "", err
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	client, err := ssh.Dial("tcp", net.JoinHostPort(addr, "22"), config)
	if err != nil {
		return "", err
	}
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	var b bytes.Buffer
	session.Stdout = &b
	err = session.Run(cmd)
	return b.String(), err
}
