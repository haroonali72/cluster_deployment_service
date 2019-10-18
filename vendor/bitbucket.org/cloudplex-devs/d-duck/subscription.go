package d_duck

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"strings"
)

type Init struct {
	Client Client
}

func (e *Init) GetCatalogs() (Catalogs, error) {
	rawCatalog, err := e.Client.GetRawCatalog()
	if err != nil {
		return Catalogs{}, err
	}

	var catalog Catalogs
	err = xml.Unmarshal(rawCatalog, &catalog)
	if err != nil {
		return Catalogs{}, err
	}

	return catalog, err
}

func (e *Init) GetProduct(subscriptionId string) (JsonProduct, error) {
	rawProduct, err := e.Client.GetRawProduct(subscriptionId)
	if err != nil {
		return JsonProduct{}, err
	}

	var product JsonProduct
	err = json.Unmarshal(rawProduct, &product)
	if err != nil {
		return JsonProduct{}, err
	}

	return product, err
}

func (e *Init) GetLimitsWithProductName(productName string) (map[string]int, error) {
	catalogs, err := e.GetCatalogs()
	if err != nil {
		return nil, err
	}

	if len(catalogs.Versions.Versions) > 0 {
		for _, product := range catalogs.Versions.Versions[0].Products.Products {
			if strings.ToLower(product.Name) == strings.ToLower(productName) {
				limits := map[string]int{}
				for _, limit := range product.Limits.Limits {
					limits[limit.Unit] = int(limit.Max)
				}
				return limits, nil
			}
		}
	}

	return map[string]int{}, err
}

func (e *Init) GetLimitsWithSubscriptionId(subscriptionId string) (map[string]int, error) {
	product, err := e.GetProduct(subscriptionId)
	if err != nil {
		return nil, err
	}

	return e.GetLimitsWithProductName(product.Name)
}

func (e *Init) GetSubscriptionId(accountId string) (string, error) {

	rawData, err := e.Client.GetSubscriptionData(accountId)
	if err != nil {
		return "", err
	}
	var accountData []AccountBudles
	err = json.Unmarshal(rawData, &accountData)
	if err != nil {
		return "", err
	}
	if len(accountData) <= 0 {
		return "", errors.New("can not find data against this account")
	}
	if len(accountData[0].Subscriptions) > 0 {
		if accountData[0].Subscriptions[0].SubscriptionId == "" {
			return "", errors.New("can not fid SubscriptionId this account")
		}

		return accountData[0].Subscriptions[0].SubscriptionId, nil
	}
	return "", errors.New("can not find Subscription against this account")

}
