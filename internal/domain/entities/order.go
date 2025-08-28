// github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities/order.go
package entities

import "strings"

type Order struct {
	OrderUID        string   `json:"order_uid" db:"order_uid"`
	TrackNumber     string   `json:"track_number" db:"track_number"`
	Entry           string   `json:"entry" db:"entry"`
	Delivery        Delivery `json:"delivery"`
	Payment         Payment  `json:"payment"`
	Items           []Item   `json:"items"`
	Locale          string   `json:"locale" db:"locale"`
	InternalSig     string   `json:"internal_signature" db:"internal_signature"`
	CustomerID      string   `json:"customer_id" db:"customer_id"`
	DeliveryService string   `json:"delivery_service" db:"delivery_service"`
	ShardKey        string   `json:"shardkey" db:"shardkey"`
	SMID            int      `json:"sm_id" db:"sm_id"`
	DateCreated     string   `json:"date_created" db:"date_created"`
	OOFShard        string   `json:"oof_shard" db:"oof_shard"`
}

func (o *Order) Equal(other Order) bool {
	return o.basicFieldsEqual(other) &&
		o.deliveryEqual(other) &&
		o.paymentEqual(other) &&
		o.itemsEqual(other)
}

func (o *Order) basicFieldsEqual(other Order) bool {
	return o.OrderUID == other.OrderUID &&
		o.TrackNumber == other.TrackNumber &&
		o.Entry == other.Entry &&
		o.Locale == other.Locale &&
		o.InternalSig == other.InternalSig &&
		o.CustomerID == other.CustomerID &&
		o.ShardKey == other.ShardKey &&
		o.SMID == other.SMID &&
		o.DateCreated == other.DateCreated &&
		o.OOFShard == other.OOFShard
}

func (o *Order) deliveryEqual(other Order) bool {
	return o.Delivery.Name == other.Delivery.Name &&
		o.Delivery.Phone == other.Delivery.Phone &&
		o.Delivery.Zip == other.Delivery.Zip &&
		o.Delivery.City == other.Delivery.City &&
		o.Delivery.Address == other.Delivery.Address &&
		o.Delivery.Region == other.Delivery.Region &&
		o.Delivery.Email == other.Delivery.Email
}

func (o *Order) paymentEqual(other Order) bool {
	return o.Payment.Transaction == other.Payment.Transaction &&
		o.Payment.RequestID == other.Payment.RequestID &&
		o.Payment.Currency == other.Payment.Currency &&
		o.Payment.Provider == other.Payment.Provider &&
		o.Payment.Amount == other.Payment.Amount &&
		o.Payment.PaymentDT == other.Payment.PaymentDT &&
		o.Payment.Bank == other.Payment.Bank &&
		o.Payment.DeliveryCost == other.Payment.DeliveryCost &&
		o.Payment.GoodsTotal == other.Payment.GoodsTotal &&
		o.Payment.CustomFee == other.Payment.CustomFee
}

func (o *Order) itemsEqual(other Order) bool {
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
	switch {
	case o.OrderUID == "":
		return ErrOrderUIDRequired
	case o.TrackNumber == "":
		return ErrTrackNumberRequired
	case len(o.Items) == 0:
		return ErrItemsEmpty
	case o.Payment.Amount < 0:
		return ErrInvalidPaymentAmount
	case o.Delivery.Email != "" && !strings.Contains(o.Delivery.Email, "@"):
		return ErrInvalidEmailFormat
	case o.Delivery.Phone != "" && !strings.HasPrefix(o.Delivery.Phone, "+"):
		return ErrInvalidPhoneFormat
	}
	return nil
}
