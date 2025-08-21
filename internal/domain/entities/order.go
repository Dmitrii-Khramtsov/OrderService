// github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities/order.go
package entities

type Order struct {
	OrderUID        string    `json:"order_uid"`
	TrackNumber     string    `json:"track_number"`
	Entry           string    `json:"entry"`
	Delivery        Delivery  `json:"delivery"`
	Payment         Payment   `json:"payment"`
	Items           []Item    `json:"items"`
	Locale          string    `json:"locale"`
	InternalSig     string    `json:"internal_signature"`
	CustomerID      string    `json:"customer_id"`
	DeliveryService string    `json:"delivery_service"`
	ShardKey        string    `json:"shardkey"`
	SMID            int       `json:"sm_id"`
	DateCreated     string    `json:"date_created"`
	OOFShard        string    `json:"oof_shard"`
}
