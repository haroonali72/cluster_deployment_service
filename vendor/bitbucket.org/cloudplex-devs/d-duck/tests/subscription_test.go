package tests

import (
	"d-duck"
	"errors"
	"strings"
	"testing"
)

func TestGetCatalog(t *testing.T) {
	testClient := d_duck.Client{}
	d_duck.Get = func(url string, accept string) ([]byte, error) {
		return []byte(TestCatalog), nil
	}

	testSubscription := d_duck.Init{Client: testClient}
	catalogs, _ := testSubscription.GetCatalogs()

	if len(catalogs.Versions.Versions) == 0 || len(catalogs.Versions.Versions[0].Products.Products) == 0 {
		t.Error("catalog fetching failed")
	}
}

func TestFailingGetCatalog(t *testing.T) {
	testClient := d_duck.Client{}
	d_duck.Get = func(url string, accept string) ([]byte, error) {
		return nil, errors.New("failing raw catalog")
	}

	testSubscription := d_duck.Init{Client: testClient}
	_, err := testSubscription.GetCatalogs()

	if err == nil {
		t.Error("failing catalog fetching failed")
	}
}

func TestGetProduct(t *testing.T) {
	testClient := d_duck.Client{}
	d_duck.Get = func(url string, accept string) ([]byte, error) {
		return []byte(TestProductName), nil
	}

	testSubscription := d_duck.Init{Client: testClient}
	product, _ := testSubscription.GetProduct("test")

	if strings.ToLower(product.Name) != "gold" {
		t.Error("product fetching failed")
	}
}

func TestFailingGetProduct(t *testing.T) {
	testClient := d_duck.Client{}
	d_duck.Get = func(url string, accept string) ([]byte, error) {
		return nil, errors.New("failing raw catalog")
	}

	testSubscription := d_duck.Init{Client: testClient}
	_, err := testSubscription.GetProduct("test")

	if err == nil {
		t.Error("failing product fetching failed")
	}
}

func TestGetLimitsWithProductName(t *testing.T) {
	testClient := d_duck.Client{}
	d_duck.Get = func(url string, accept string) ([]byte, error) {
		return []byte(TestCatalog), nil
	}

	testSubscription := d_duck.Init{Client: testClient}
	limits, _ := testSubscription.GetLimitsWithProductName("DEVELOPER")

	if len(limits) != 4 {
		t.Error("failed to get limits for product 'developer'")
	}
	if limits["MeshCount"] != 1 {
		t.Error("MeshCount limit is incorrect for product 'developer'")
	}
	if limits["MeshSize"] != 25 {
		t.Error("MeshSize limit is incorrect for product 'developer'")
	}
	if limits["CoreCount"] != 12 {
		t.Error("CoreCount limit is incorrect for product 'developer'")
	}
	if limits["DeveloperCount"] != 5 {
		t.Error("DeveloperCount limit is incorrect for product 'developer'")
	}
}

func TestGetLimitsWithUnknownProductName(t *testing.T) {
	testClient := d_duck.Client{}
	d_duck.Get = func(url string, accept string) ([]byte, error) {
		return []byte(TestCatalog), nil
	}

	testSubscription := d_duck.Init{Client: testClient}
	limits, _ := testSubscription.GetLimitsWithProductName("DEVELOPER1")

	if limits == nil {
		t.Error("nil pointer returned for unknown product")
	}
	if len(limits) != 0 {
		t.Error("fetched limits for unknown product")
	}
}

func TestFailingGetLimitsWithProductName(t *testing.T) {
	testClient := d_duck.Client{}
	d_duck.Get = func(url string, accept string) ([]byte, error) {
		return nil, errors.New("failing raw catalog")
	}

	testSubscription := d_duck.Init{Client: testClient}
	_, err := testSubscription.GetLimitsWithProductName("DEVELOPER")

	if err == nil {
		t.Error("failing limit fetching failed")
	}
}
