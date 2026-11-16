package pow

import "fmt"

func buildKey(salt string, timestamp int64, nonce string) string {
	return fmt.Sprintf("%s:%d:%s", salt, timestamp, nonce)
}
