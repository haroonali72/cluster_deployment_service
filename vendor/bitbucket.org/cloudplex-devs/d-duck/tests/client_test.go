package tests

import (
	"d-duck"
	"encoding/json"
	"encoding/xml"
	"errors"
	"strings"
	"testing"
)

func TestGetRawCatalog(t *testing.T) {
	testClient := d_duck.Client{}
	d_duck.Get = func(url string, accept string) ([]byte, error) {
		return []byte(TestCatalog), nil
	}

	rawCatalog, _ := testClient.GetRawCatalog()
	var catalogs d_duck.Catalogs
	_ = xml.Unmarshal(rawCatalog, &catalogs)

	if len(catalogs.Versions.Versions) == 0 || len(catalogs.Versions.Versions[0].Products.Products) == 0 {
		t.Error("catalog fetching failed")
	}
}

func TestFailingGetRawCatalog(t *testing.T) {
	testClient := d_duck.Client{}
	d_duck.Get = func(url string, accept string) ([]byte, error) {
		return nil, errors.New("failing raw catalog")
	}

	_, err := testClient.GetRawCatalog()
	if err == nil {
		t.Error("failing catalog fetching failed")
	}
}

func TestGetRawProduct(t *testing.T) {
	testClient := d_duck.Client{}
	d_duck.Get = func(url string, accept string) ([]byte, error) {
		return []byte(TestProductName), nil
	}

	rawProduct, _ := testClient.GetRawProduct("subscriptionId")
	var product d_duck.JsonProduct
	_ = json.Unmarshal(rawProduct, &product)

	if strings.ToLower(product.Name) != "gold" {
		t.Error("product fetching failed")
	}
}

func TestFailingGetRawProduct(t *testing.T) {
	testClient := d_duck.Client{}
	d_duck.Get = func(url string, accept string) ([]byte, error) {
		return nil, errors.New("failing raw catalog")
	}

	_, err := testClient.GetRawProduct("subscriptionId")
	if err == nil {
		t.Error("failing product fetching failed")
	}
}
