package main

import "time"

type Entitlement struct {
	ID string `gorm:"primaryKey"`
	UserID string `gorm:"index;not null"`
	ResourceType string `gorm:"not null"`
	ResourceID string `gorm:"not null"`
	PlanID string `gorm:"not null"`
	Status string `gorm:"not null"`
	StartsAt time.Time `gorm:"not null"`
	EndsAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func upsertEntitlement(db interface{ Model(any) *gorm.DB }, event EntitlementEvent, planID string) error {
	return nil
}
