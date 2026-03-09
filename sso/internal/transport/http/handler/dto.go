package handler

type (
	CreateUserReq struct {
		Name     string `json:"name" example:"name"`
		LastName string `json:"last_name" example:"lastName"`
		Login    string `json:"login" example:"login"`
		Email    string `json:"email" example:"email@example.com"`
		Password string `json:"password" example:"password"`
	}

	CreateUserResp struct {
		ID int64 `json:"id" example:"1"`
	}
)

type (
	LoginReq struct {
		IsEmail    bool   `json:"is_email" example:"true"`
		Identifier string `json:"identifier" example:"email@example.com"`
		Password   string `json:"password" example:"password"`
	}

	LoginResp struct {
		AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
		UserID      int64  `json:"user_id" example:"1"`
		Login       string `json:"login" example:"login"`
		Role        string `json:"role" example:"user"`
	}
)

type AuthResp struct {
	UserID int64  `json:"user_id" example:"1"`
	Role   string `json:"role" example:"user"`
}

type LoginByPhoneReq struct {
	Phone    string `json:"phone" example:"+380501234567"`
	Password string `json:"password" example:"password"`
}

type (
	SendOTPByPhoneReq struct {
		Phone string `json:"phone" example:"+380501234567"`
	}

	ConfirmOTPByPhoneReq struct {
		Phone string `json:"phone" example:"+380501234567"`
		Code  int64  `json:"code" example:"1234"`
	}
)

type (
	ResetPhoneReq struct {
		Phone string `json:"phone" example:"+380501234567"`
	}

	ResetPhoneConfirmReq struct {
		Phone       string `json:"phone" example:"+380501234567"`
		Code        int64  `json:"code" example:"1234"`
		NewPassword string `json:"new_password" example:"newStrongPassword"`
	}
)

type (
	ResetEmailReq struct {
		Email string `json:"email" example:"email@example.com"`
	}

	ResetEmailConfirmReq struct {
		Email       string `json:"email" example:"email@example.com"`
		Code        int64  `json:"code" example:"1234"`
		NewPassword string `json:"new_password" example:"newStrongPassword"`
	}
)
