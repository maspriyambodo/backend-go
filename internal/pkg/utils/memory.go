package utils

import (
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// responseObjectPool reuses gin.H objects to reduce allocations
var responseObjectPool = sync.Pool{
	New: func() interface{} {
		return make(gin.H)
	},
}

// GetResponseObject returns a reusable gin.H object from the pool
func GetResponseObject() gin.H {
	return responseObjectPool.Get().(gin.H)
}

// PutResponseObject returns a gin.H object to the pool for reuse
func PutResponseObject(obj gin.H) {
	// Clear the map before putting it back
	clear(obj) // Go 1.21+
	responseObjectPool.Put(obj)
}

// ResetSliceToZeroLen resets a slice to zero length without deallocating capacity
// This helps reuse the underlying array when growing again
func ResetSliceToZeroLen[T any](slice []T) []T {
	return slice[:0]
}

// JoinStrings efficiently joins strings with a separator
// More memory-efficient than repeated concatenation
func JoinStrings(parts []string, sep string) string {
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	default:
		return strings.Join(parts, sep)
	}
}

// JoinStringSlice is an alias for JoinStrings for backward compatibility
func JoinStringSlice(parts []string, sep string) string {
	return JoinStrings(parts, sep)
}
