package api

import (
	"encoding/json"
	"fmt"
	"testing"
)

func Test_MicroApp(t *testing.T) {
	ma := MicroApp{
		Addr:      "http://127.0.0.1:80/rs",
		Providers: []string{"a.b", "a.c"},
	}

	jma, _ := json.Marshal(ma)
	fmt.Println(string(jma))
}
