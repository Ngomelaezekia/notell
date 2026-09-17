package main

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func securedMembers(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !hasPermission(db, c, channelID(c), "team") {
            return
        }
        listMembers(db)(c)
    }
}

func securedPrograms(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !hasPermission(db, c, channelID(c), "schedule") {
            return
        }
        programs(db)(c)
    }
}

func securedSchedules(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !hasPermission(db, c, channelID(c), "schedule") {
            return
        }
        schedules(db)(c)
    }
}

func securedContent(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        if !hasPermission(db, c, channelID(c), "content") {
            return
        }
        content(db)(c)
    }
}
