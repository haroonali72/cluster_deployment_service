# d-duck
A client library for `cloudplex` subscription engine.

# How to install
```
go get -u bitbucket.org/cloudplex-devs/d-duck
```

# How to use
You can get accountId from RBAC api /security/api/rbac/token/extract 
And use that accountId to get subscriptionId using this method GetSubscriptionId
Use the following code to initialize and fetch the subscription details:

```go
package main

import (
	"bitbucket.org/cloudplex-devs/d-duck"
	"encoding/json"
)

func main() {
    subscriptionClient := d_duck.Init{Client: d_duck.Client{
        Host: "122.129.74.5",
        Port: "8080",
        Username: "admin",
        Password: "password",
        ApiKey: "apiKey",
        ApiSecret: "apiSecret",
    }}
    subId, err := subscriptionClient.GetSubscriptionId("testaccounid") 
    if err !=nil{
        println(err.Error())
        return
    }
    limits, err := subscriptionClient.GetLimitsWithSubscriptionId(subId)
    if err != nil {
        println(err.Error())
        return
    }
    
    b, err := json.MarshalIndent(limits, "", "  ")
    if err != nil {
        println(err.Error())
    }
    println(string(b))
}
```

And the result will look like:
```json
{
  "CoreCount": 180,
  "DeveloperCount": 500,
  "MeshCount": 15,
  "MeshSize": 300
}
```