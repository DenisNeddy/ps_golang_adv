package auth

type SendSMSRequest struct {
	Phone string `json:"phone"`
}

type SendSMSResponse struct {
	SessionID string `json:"sessionId"`
}

type VerifyCodeRequest struct {
	SessionID string `json:"sessionId"`
	Code      int    `json:"code"`
}

type VerifyCodeResponse struct {
	Token string `json:"token"`
}
