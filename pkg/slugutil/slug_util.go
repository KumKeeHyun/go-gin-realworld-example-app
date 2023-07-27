package slugutil

import (
	"encoding/hex"
	"github.com/gosimple/slug"
	"math/rand"
)

func init() {
	slug.MaxLength = 20
}

func Make(title string) string {
	slugString := slug.MakeLang(title, "en")
	b := make([]byte, 4) //equals 8 characters
	rand.Read(b)
	ridString := hex.EncodeToString(b)
	return slugString + "-" + ridString
}
