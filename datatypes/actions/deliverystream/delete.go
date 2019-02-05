package deliverystream

type DeleteInput struct {
	Name string `json:"DeliveryStreamName"`
}

func (i DeleteInput) Validate() error {
	return validateName(i.Name)
}
