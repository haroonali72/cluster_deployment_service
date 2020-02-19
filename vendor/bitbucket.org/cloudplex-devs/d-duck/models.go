package d_duck

import (
	"encoding/xml"
	"time"
)

type Catalogs struct {
	XMLName  xml.Name `xml:"catalogs"`
	Versions Versions `xml:"versions"`
}

type Versions struct {
	XMLName  xml.Name  `xml:"versions"`
	Versions []Version `xml:"version"`
}

type Version struct {
	XMLName       xml.Name  `xml:"version"`
	EffectiveDate time.Time `xml:"effectiveDate"`
	Products      Products  `xml:"products"`
}

type Products struct {
	XMLName  xml.Name  `xml:"products"`
	Products []Product `xml:"product"`
}

type Product struct {
	XMLName xml.Name `xml:"product"`
	Name    string   `xml:"name,attr"`
	Limits  Limits   `xml:"limits"`
}

type Limits struct {
	XMLName xml.Name `xml:"limits"`
	Limits  []Limit  `xml:"limit"`
}

type Limit struct {
	XMLName xml.Name `xml:"limit"`
	Unit    string   `xml:"unit"`
	Max     float32  `xml:"max"`
}

type JsonProduct struct {
	Name string `json:"name"`
}
