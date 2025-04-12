// // models/company.go
// package models

// import (
// 	"time"

// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )

// type Company struct {
// 	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
// 	UserID       primitive.ObjectID `json:"userId" bson:"userId"` // Reference to the user account
// 	BusinessName string             `json:"businessName" bson:"businessName"`
// 	Category     string             `json:"category" bson:"category"`
// 	SubCategory  string             `json:"subCategory,omitempty" bson:"subCategory,omitempty"`
// 	ReferralCode string             `json:"referralCode,omitempty" bson:"referralCode,omitempty"`
// 	ContactInfo  ContactInfo        `json:"contactInfo" bson:"contactInfo"`
// 	SocialMedia  SocialMedia        `json:"socialMedia,omitempty" bson:"socialMedia,omitempty"`
// 	LogoURL      string             `json:"logoUrl,omitempty" bson:"logoUrl,omitempty"`
// 	Balance      float64            `json:"balance" bson:"balance"`
// 	Branches     []Branch           `json:"branches,omitempty" bson:"branches,omitempty"`
// 	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
// 	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
// }

// type ContactInfo struct {
// 	Phone    string  `json:"phone" bson:"phone"`
// 	WhatsApp string  `json:"whatsapp,omitempty" bson:"whatsapp,omitempty"`
// 	Website  string  `json:"website,omitempty" bson:"website,omitempty"`
// 	Address  Address `json:"address" bson:"address"`
// }

// type SocialMedia struct {
// 	Facebook  string `json:"facebook,omitempty" bson:"facebook,omitempty"`
// 	Instagram string `json:"instagram,omitempty" bson:"instagram,omitempty"`
// }

// type Address struct {
// 	Country    string  `json:"country" bson:"country"`
// 	District   string  `json:"district" bson:"district"`
// 	City       string  `json:"city" bson:"city"`
// 	Street     string  `json:"street" bson:"street"`
// 	PostalCode string  `json:"postalCode" bson:"postalCode"`
// 	Lat        float64 `json:"lat" bson:"lat"`
// 	Lng        float64 `json:"lng" bson:"lng"`
// }

// type Branch struct {
// 	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
// 	Name        string             `json:"name" bson:"name"`
// 	Location    Address            `json:"location" bson:"location"`
// 	Phone       string             `json:"phone" bson:"phone"`
// 	Category    string             `json:"category" bson:"category"`
// 	SubCategory string             `json:"subCategory,omitempty" bson:"subCategory,omitempty"`
// 	Description string             `json:"description,omitempty" bson:"description,omitempty"`
// 	Images      []string           `json:"images" bson:"images"`
// 	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
// 	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
// }
