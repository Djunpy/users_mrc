package entities

type PayloadSendVerifyEmail struct {
	Email    string `json:"email"`
	JWTToken string `json:"jwt_token"`
	LangCode string `json:"lang_code"`
}
