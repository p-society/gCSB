package tempModel

import "time"

type PlayerTemp struct {
	Name          string                   `json:"name,omitempty" bson:"name,omitempty"`
	Sport         string                   `json:"sport,omitempty" bson:"sport,omitempty"`
	Team         string                   `json:"team,omitempty" bson:"team,omitempty"`
	Email         string                   `json:"email,omitempty" bson:"email,omitempty"`
	MetaData      []map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
	OTP           int                      `bson:"otp,omitempty" bson:"otp,omitempty"`
	Verified      bool                     `json:"verified" bson:"verified"`
	OTPExpiration time.Time                `json:"otpexpirationtime,omitempty" bson:"otpexpirationtime,omitempty"`
}
