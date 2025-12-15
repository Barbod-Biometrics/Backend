package entity

import "github.com/shopspring/decimal"

type Service struct {
	ServiceID   uint64          `gorm:"primaryKey;autoIncrement" json:"service_id"`
	ServiceName string          `gorm:"type:varchar(225);unique;not null" json:"service_name"`
	CurrentCost decimal.Decimal `gorm:"type:decimal(10,2);not null;default:0" json:"current_cost"`
	IsAvailable bool            `gorm:"not null;default:true" json:"is_available"`
}

func (Service) TableName() string {
	return "services"
}
