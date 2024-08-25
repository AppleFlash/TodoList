package models

type (
	Credentionals struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	User struct {
		Id    string `json:"id"`
		Login string `json:"login"`
	}

	StorageUser struct {
		User     User
		Password string
	}
)
