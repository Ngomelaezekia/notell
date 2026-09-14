package main

import (
    "fmt"
    "gorm.io/gorm"
)

type ChannelSchemaMigration struct {
    Version int `gorm:"primaryKey"`
    Name string `gorm:"size:120;not null"`
}

func runChannelMigrations(db *gorm.DB) error {
    if err := db.AutoMigrate(&ChannelSchemaMigration{}); err != nil { return err }
    migrations := []struct{ version int; name string; run func(*gorm.DB) error }{
        {1, "initial_channel_schema", func(tx *gorm.DB) error {
            return tx.AutoMigrate(&Channel{},&ChannelMember{},&ChannelPlan{},&ChannelEntitlement{},&ChannelTeamMember{},&ChannelInvitation{},&ChannelProgram{},&ChannelSchedule{},&ChannelContent{},&ChannelSetting{},&LiveSession{},&MediaReference{},&SafetyCase{},&RightsRecord{},&MonetizationEligibility{},&RevenueEvent{},&AdCampaign{},&ChannelMetric{})
        }},
        {2, "membership_index_repair", repairMembershipIndex},
    }
    for _, m := range migrations {
        var applied ChannelSchemaMigration
        err := db.Where("version = ?", m.version).First(&applied).Error
        if err == nil { continue }
        if err != gorm.ErrRecordNotFound { return err }
        if err := db.Transaction(func(tx *gorm.DB) error {
            if err := m.run(tx); err != nil { return err }
            if err := tx.Create(&ChannelSchemaMigration{Version:m.version,Name:m.name}).Error; err != nil { return err }
            return nil
        }); err != nil { return fmt.Errorf("migration %d (%s) failed: %w",m.version,m.name,err) }
    }
    return nil
}
