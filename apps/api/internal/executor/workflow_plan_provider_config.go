package executor

type yclientsActionConfig struct {
	OperationKey      string `json:"operation_key"`
	OrderID           string `json:"order_id"`
	OrderItemID       string `json:"order_item_id"`
	InstanceID        string `json:"instance_id"`
	CertificateNumber string `json:"certificate_number"`
	ProductKind       string `json:"product_kind"`
	ProductTitle      string `json:"product_title"`
	AmountRub         string `json:"amount_rub"`
	ServiceIDsJSON    string `json:"service_ids_json"`
}
