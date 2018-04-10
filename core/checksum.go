package core

import (
	"crypto/md5"
	"fmt"
)

func GetMD5(data []byte) string {
	sum := md5.Sum(data)

	return fmt.Sprintf("%x", sum)
}
