// github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities/order.go
package entities

import (
	"errors"
	"strings"
)

type Order struct {
	OrderUID        string   `json:"order_uid"`
	TrackNumber     string   `json:"track_number"`
	Entry           string   `json:"entry"`
	Delivery        Delivery `json:"delivery"`
	Payment         Payment  `json:"payment"`
	Items           []Item   `json:"items"`
	Locale          string   `json:"locale"`
	InternalSig     string   `json:"internal_signature"`
	CustomerID      string   `json:"customer_id"`
	DeliveryService string   `json:"delivery_service"`
	ShardKey        string   `json:"shardkey"`
	SMID            int      `json:"sm_id"`
	DateCreated     string   `json:"date_created"`
	OOFShard        string   `json:"oof_shard"`
}

func (o *Order) Equal(other Order) bool {
	if o.OrderUID != other.OrderUID ||
		o.TrackNumber != other.TrackNumber ||
		o.Entry != other.Entry ||

		o.Delivery.Name != other.Delivery.Name ||
		o.Delivery.Phone != other.Delivery.Phone ||
		o.Delivery.Zip != other.Delivery.Zip ||
		o.Delivery.City != other.Delivery.City ||
		o.Delivery.Address != other.Delivery.Address ||
		o.Delivery.Region != other.Delivery.Region ||
		o.Delivery.Email != other.Delivery.Email ||

		o.Payment.Transaction != other.Payment.Transaction ||
		o.Payment.RequestID != other.Payment.RequestID ||
		o.Payment.Currency != other.Payment.Currency ||
		o.Payment.Provider != other.Payment.Provider ||
		o.Payment.Amount != other.Payment.Amount ||
		o.Payment.PaymentDT != other.Payment.PaymentDT ||
		o.Payment.Bank != other.Payment.Bank ||
		o.Payment.DeliveryCost != other.Payment.DeliveryCost ||
		o.Payment.GoodsTotal != other.Payment.GoodsTotal ||
		o.Payment.CustomFee != other.Payment.CustomFee ||

		o.Locale != other.Locale ||
		o.InternalSig != other.InternalSig ||
		o.CustomerID != other.CustomerID ||
		o.ShardKey != other.ShardKey ||
		o.SMID != other.SMID ||
		o.DateCreated != other.DateCreated ||
		o.OOFShard != other.OOFShard {
		return false
	}

	if len(o.Items) != len(other.Items) {
		return false
	}

	for i := range o.Items {
		if !o.Items[i].Equal(other.Items[i]) {
			return false
		}
	}

	return true
}

func (o Order) Validate() error {
	if o.OrderUID == "" {
		return errors.New("order_uid is required")
	}
	if o.TrackNumber == "" {
		return errors.New("track_number is required")
	}
	if len(o.Items) == 0 {
		return errors.New("items cannot be empty")
	}
	if o.Payment.Amount <= 0 {
		return errors.New("payment amount must be > 0")
	}
	if o.Delivery.Email != "" && !strings.Contains(o.Delivery.Email, "@") {
		return errors.New("invalid email format")
	}
	if !strings.HasPrefix(o.Delivery.Phone, "+") {
		return errors.New("phone must start with +")
	}
	return nil
}
