package d_duck

type AccountBudles struct {
	Subscriptions []Subscription `json:"subscriptions"`
}

type Subscription struct {
	SubscriptionId string `json:"subscriptionId"`
	State          string `json:"state"`
}
