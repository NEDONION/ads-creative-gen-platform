package service

import "github.com/google/uuid"

// randBucket 将 userKey hash 到 [0,9999]
func randBucket(userKey string) int {
	if userKey == "" {
		userKey = uuid.New().String()
	}
	h := uuid.NewSHA1(uuid.NameSpaceOID, []byte(userKey))
	b := h[:]
	val := int(b[0])<<8 + int(b[1])
	return val % 10000
}
