package generatehelpers

import (
	"time"

	"gorm.io/gorm"
)

type CreatedAtAble struct {
	CreatedAt time.Time
}

func (c *CreatedAtAble) BeforeCreate(tx *gorm.DB) (err error) {
	c.CreatedAt = time.Now()
	return
}

type UpdatedAtAble struct {
	UpdatedAt time.Time
}

func (c *UpdatedAtAble) BeforeCreate(tx *gorm.DB) (err error) {
	c.UpdatedAt = time.Now()
	return
}

func (c *UpdatedAtAble) BeforeUpdate(tx *gorm.DB) (err error) {
	c.UpdatedAt = time.Now()
	return
}

type CreatedAtUpdatedAtAble struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c *CreatedAtUpdatedAtAble) BeforeCreate(tx *gorm.DB) (err error) {
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return
}

func (c *CreatedAtUpdatedAtAble) BeforeUpdate(tx *gorm.DB) (err error) {
	c.UpdatedAt = time.Now()
	return
}
