package secret

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type Object struct {
	Auths map[string]*Auth `json:"auths"`
}

type Auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Auth     string `json:"auth"`
}

func (o *Object) Encode() []byte {
	for _, auth := range o.Auths {
		au := fmt.Sprintf("%s:%s", auth.Username, auth.Password)
		auth.Auth = base64.StdEncoding.EncodeToString([]byte(au))
	}

	bytes, _ := json.Marshal(o)

	return bytes
}
