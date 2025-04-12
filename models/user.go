// models/user.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User model
type User struct {
	ID                  primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	Email               string               `json:"email" bson:"email"`
	Password            string               `json:"password,omitempty" bson:"password"`
	FullName            string               `json:"fullName" bson:"fullName"`
	UserType            string               `json:"userType" bson:"userType"`
	DateOfBirth         string               `json:"dateOfBirth,omitempty" bson:"dateOfBirth,omitempty"`
	Gender              string               `json:"gender,omitempty" bson:"gender,omitempty"`
	Phone               string               `json:"phone,omitempty" bson:"phone,omitempty"`
	ReferralCode        string               `json:"referralCode,omitempty" bson:"referralCode,omitempty"`
	InterestedDeals     []string             `json:"interestedDeals,omitempty" bson:"interestedDeals,omitempty"`
	Location            *Location            `json:"location,omitempty" bson:"location,omitempty"`
	CompanyInfo         *CompanyInfo         `json:"companyInfo,omitempty" bson:"companyInfo,omitempty"`
	ServiceProviderInfo *ServiceProviderInfo `json:"serviceProviderInfo,omitempty" bson:"serviceProviderInfo,omitempty"`
	WholesalerInfo      *WholesalerInfo      `json:"wholesalerInfo,omitempty" bson:"wholesalerInfo,omitempty"`
	LogoPath            string               `json:"logoPath,omitempty" bson:"logoPath,omitempty"`
	OTPInfo             *OTPInfo             `json:"otpInfo,omitempty" bson:"otpInfo,omitempty"`
	ResetPasswordToken  string               `json:"resetPasswordToken,omitempty" bson:"resetPasswordToken,omitempty"`
	ResetTokenExpiresAt time.Time            `json:"resetTokenExpiresAt,omitempty" bson:"resetTokenExpiresAt,omitempty"`
	GoogleUID           string               `bson:"googleUID,omitempty" json:"googleUID,omitempty"`
	ProfilePic          string               `bson:"profilePic,omitempty" json:"profilePic,omitempty"`
	CreatedAt           time.Time            `json:"createdAt" bson:"createdAt"`
	UpdatedAt           time.Time            `json:"updatedAt" bson:"updatedAt"`
}

type OTPInfo struct {
	OTP       string    `json:"otp" bson:"otp"`
	ExpiresAt time.Time `json:"expiresAt" bson:"expiresAt"`
}

// Location model
type Location struct {
	City       string  `json:"city" bson:"city"`
	Country    string  `json:"country" bson:"country"`
	District   string  `json:"district" bson:"district"`
	Street     string  `json:"street" bson:"street"`
	PostalCode string  `json:"postalCode" bson:"postalCode"`
	Lat        float64 `json:"lat" bson:"lat"`
	Lng        float64 `json:"lng" bson:"lng"`
	Allowed    bool    `json:"allowed" bson:"allowed"`
}

// CompanyInfo model
type CompanyInfo struct {
	Name           string   `json:"name" bson:"name"`
	Category       string   `json:"Category" bson:"Category"`
	CustomCategory string   `json:"customCategory,omitempty" bson:"customCategory,omitempty"`
	Logo           string   `json:"logo,omitempty" bson:"logo,omitempty"`
	SubCategory    string   `json:"subCategory,omitempty" bson:"subCategory,omitempty"`
	Branches       []Branch `bson:"branches,omitempty" json:"branches,omitempty"`
	Details        []Detail `bson:"details,omitempty" json:"details,omitempty"`
}

type WholesalerInfo struct {
	BusinessName string `json:"businessName" bson:"businessName"`
	Category     string `json:"Category" bson:"Category"`
	ReferralCode string `json:"referralCode,omitempty" bson:"referralCode,omitempty"`
}

// ServiceProviderInfo model
type ServiceProviderInfo struct {
	ServiceType       string   `json:"serviceType" bson:"serviceType"`
	CustomServiceType string   `json:"customServiceType,omitempty" bson:"customServiceType,omitempty"` // For "Other" service type
	YearsExperience   int      `json:"yearsExperience" bson:"yearsExperience"`
	ProfilePhoto      string   `json:"profilePhoto,omitempty" bson:"profilePhoto,omitempty"`
	AvailableHours    []string `json:"availableHours,omitempty" bson:"availableHours,omitempty"`
	AvailableDays     []string `json:"availableDays,omitempty" bson:"availableDays,omitempty"`
}

// AuthRequest models
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupRequest struct {
	Email               string               `json:"email"`
	Password            string               `json:"password"`
	FullName            string               `json:"fullName"`
	UserType            string               `json:"userType"`
	DateOfBirth         string               `json:"dateOfBirth,omitempty"`
	Gender              string               `json:"gender,omitempty"`
	Phone               string               `json:"phone,omitempty"`
	ReferralCode        string               `json:"referralCode,omitempty"`
	InterestedDeals     []string             `json:"interestedDeals,omitempty"`
	Location            *Location            `json:"location,omitempty"`
	CompanyInfo         *CompanyInfo         `json:"companyInfo,omitempty"`
	ServiceProviderInfo *ServiceProviderInfo `json:"serviceProviderInfo,omitempty"`
	WholesalerInfo      *WholesalerInfo      `json:"wholesalerInfo,omitempty"`
}

type UpdateLocationRequest struct {
	Location *Location `json:"location"`
}

// Response model
type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Branch struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Location    string             `bson:"location" json:"location"`
	Latitude    float64            `bson:"latitude" json:"latitude"`
	Longitude   float64            `bson:"longitude" json:"longitude"`
	Category    string             `bson:"category" json:"category"`
	SubCategory string             `bson:"subCategory" json:"subCategory"`
	Phone       string             `bson:"phone" json:"phone"`
	Description string             `bson:"description" json:"description"`
	Images      []string           `bson:"images" json:"images"`
	Rate        float64            `bson:"rate" json:"rate"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Detail struct {
	Phone     string `json:"phone,omitempty" bson:"phone,omitempty"`
	WhatsApp  string `json:"whatsapp,omitempty" bson:"whatsapp,omitempty"`
	Website   string `json:"website,omitempty" bson:"website,omitempty"`
	Facebook  string `json:"facebook,omitempty" bson:"facebook,omitempty"`
	Instagram string `json:"instagram,omitempty" bson:"instagram,omitempty"`
}
