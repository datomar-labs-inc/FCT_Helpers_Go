package flagger

const (
	OpGetUserInfo = "get_user_info"
)

type WebHooker interface {
	GetUserInfo(req *GetUserInfoRequest) (*GetUserInfoResponse, error)
}

type Webhook struct {
	Op          string              `json:"op"`
	GetUserInfo *GetUserInfoRequest `json:"get_user_info,omitempty"`
}

type GetUserInfoRequest struct {
	UserIDs []string `json:"user_ids"`
}

type GetUserInfoResponse struct {
	Users map[string]UserInfo `json:"users"`
}
