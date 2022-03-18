package flagger

const (
	AuthSystemCustom   = "custom"
	AuthSystemFirebase = "firebase"
	AuthSystemAuth0    = "auth0"
)

const (
	AuthProviderCustom           = "custom"
	AuthProviderUsernamePassword = "username_password"
	AuthProviderGoogle           = "goog"
	AuthProviderFacebook         = "facebook"
)

type UserInfo struct {
	ID           string  `json:"id"`
	Name         *string `json:"name,omitempty"`
	Email        *string `json:"email,omitempty"`
	Phone        *string `json:"phone,omitempty"`
	ImageURL     *string `json:"image_url,omitempty"`
	AuthSystem   *string `json:"auth_system,omitempty"`
	AuthProvider *string `json:"auth_provider,omitempty"`
}
