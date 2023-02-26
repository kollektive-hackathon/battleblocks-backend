package model

import "time"

type NftPurchaseHistory struct {
	NftId       uint64
	BuyerId     uint64
	PurchasedAt time.Time
}
