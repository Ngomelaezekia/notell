package main

import (
	"time"
	"gorm.io/gorm"
)

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

func upsertEntitlement(db *gorm.DB, event EntitlementEvent, planID string) error {
	var row Entitlement
	err := db.Where("user_id = ? AND resource_type = ? AND resource_id = ?", event.UserID, event.ResourceType, event.ResourceID).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return db.Create(&Entitlement{
			ID: newID("ent"),
			UserID: event.UserID,
			ResourceType: event.ResourceType,
			ResourceID: event.ResourceID,
			PlanID: planID,
			Status: event.Status,
			StartsAt: event.EffectiveAt,
		}).Error
	}
	if err != nil { return err }
	row.Status = event.Status
	row.PlanID = planID
	return db.Save(&row).Error
}
