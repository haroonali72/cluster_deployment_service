package models

type Type string

const (
	Existing Type = "existing"
	New      Type = "new"
)

type Cloud string

const (
	AWS   Cloud = "aws"
	Azure Cloud = "azure"
	GCP   Cloud = "gcp"
)
