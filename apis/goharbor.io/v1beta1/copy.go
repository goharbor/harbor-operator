package v1beta1

import "encoding/json"

func CopyViaJSON(src, dest interface{}) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}
