package order

import (
	"time"
)

type Order struct {
	OrderUID          string    `json:"order_uid" db:"order_uid"`
	TrackNumber       string    `json:"track_number" db:"track_number" validate:"required"`
	Entry             string    `json:"entry" db:"entry" validate:"required"`
	Locale            string    `json:"locale" db:"locale" validate:"required"`
	InternalSignature string    `json:"internal_signature" db:"internal_signature"`
	CustomerID        string    `json:"customer_id" db:"customer_id" validate:"required"`
	DeliveryService   string    `json:"delivery_service" db:"delivery_service" validate:"required"`
	ShardKey          string    `json:"shardkey" db:"shardkey" validate:"required"`
	SmID              int       `json:"sm_id" db:"sm_id" validate:"required"`
	DateCreated       time.Time `json:"date_created" db:"date_created" `
	OofShard          string    `json:"oof_shard" db:"oof_shard" validate:"required"`
	Delivery          *Delivery `json:"delivery" validate:"required"`
	Payment           *Payment  `json:"payment" validate:"required"`
	Items             []*Item   `json:"items" validate:"required"`
}

type Delivery struct {
	DeliveryID string `json:"delivery_id" db:"delivery_id"`
	Name       string `json:"name" db:"name" validate:"required"`
	Phone      string `json:"phone" db:"phone" validate:"required"`
	Zip        string `json:"zip" db:"zip" validate:"required"`
	City       string `json:"city" db:"city" validate:"required"`
	Address    string `json:"address" db:"address" validate:"required"`
	Region     string `json:"region" db:"region" validate:"required"`
	Email      string `json:"email" db:"email" validate:"required"`
}

type Payment struct {
	PaymentID    string `json:"payment_id" db:"payment_id"`
	Transaction  string `json:"transaction" db:"transaction" validate:"required"`
	RequestID    string `json:"request_id" db:"request_id" `
	Currency     string `json:"currency" db:"currency" validate:"required"`
	Provider     string `json:"provider" db:"provider" validate:"required"`
	Amount       int    `json:"amount" db:"amount" validate:"required"`
	PaymentDT    int    `json:"payment_dt" db:"payment_dt" validate:"required"`
	Bank         string `json:"bank" db:"bank" validate:"required"`
	DeliveryCost int    `json:"delivery_cost" db:"delivery_cost" validate:"required"`
	GoodsTotal   int    `json:"goods_total" db:"goods_total" validate:"required"`
	CustomFee    int    `json:"custom_fee" db:"custom_fee" `
}

type Item struct {
	ItemID      string `json:"item_id" db:"item_id"`
	ChrtID      int    `json:"chrt_id" db:"chrt_id" validate:"required"`
	TrackNumber string `json:"track_number" db:"track_number" validate:"required"`
	Price       int    `json:"price" db:"price" validate:"required"`
	Rid         string `json:"rid" db:"rid" validate:"required"`
	Name        string `json:"name" db:"name" validate:"required"`
	Sale        int    `json:"sale" db:"sale"`
	Size        string `json:"size" db:"size" validate:"required"`
	TotalPrice  int    `json:"total_price" db:"total_price" validate:"required"`
	NmID        int    `json:"nm_id" db:"nm_id" validate:"required"`
	Brand       string `json:"brand" db:"brand" validate:"required"`
	Status      int    `json:"status" db:"status" validate:"required"`
}
