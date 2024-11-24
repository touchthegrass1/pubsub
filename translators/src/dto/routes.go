package dto

import "github.com/gin-gonic/gin"

type Routes struct {
	Method      string
	Pattern     string
	HandlerFunc gin.HandlerFunc
}
