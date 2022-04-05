package flagger

import "context"

const (
	OpGetUserInfo = "get_user_info"
	OpSearchUsers = "search_users"
)

type WebHooker interface {
	GetUserInfo(ctx context.Context, req *GetUserInfoRequest) (*GetUserInfoResponse, error)
	SearchUsers(ctx context.Context, req *SearchUsersRequest) (*SearchUsersResponse, error)
}

type Webhook struct {
	Op          string              `json:"op"`
	GetUserInfo *GetUserInfoRequest `json:"get_user_info,omitempty"`
	SearchUsers *SearchUsersRequest `json:"search_users,omitempty"`
}

type GetUserInfoRequest struct {
	UserIDs []string `json:"user_ids"`
}

type GetUserInfoResponse struct {
	Users map[string]UserInfo `json:"users"`
}

type SearchUsersRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`

	ByName  *string `json:"by_name,omitempty"`
	ByEmail *string `json:"by_email,omitempty"`
	ByPhone *string `json:"by_phone,omitempty"`
	ByID    *string `json:"by_id,omitempty"`
}

type SearchUsersResponse struct {
	Users      []UserInfo `json:"users"`
	TotalCount int        `json:"total_count"`
}
