package handlers

import (
	"errors"
	"net/http"
	"translators/src/dto"
	"translators/src/repositories"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SubscribeHandler struct {
	Repo repositories.SubscribeRepository
	Log  *zap.Logger
}

func (handler SubscribeHandler) Get(c *gin.Context) {
	ip := c.Param("ip")
	subs, err := handler.Repo.FindMany(ip)
	if err != nil {
		handler.Log.Error("Error in SubscribeHandler", zap.Error(err))
	}
	c.JSON(http.StatusOK, subs)
}

func (handler SubscribeHandler) Post(c *gin.Context) {
	var sub dto.Subscription
	err := c.Bind(&sub)
	if err != nil {
		handler.Log.Error("Error binding request data to Subscribtion dto", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = handler.Repo.Create(sub)
	if err != nil {
		handler.Log.Error("Error creating subscribtion from dto", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})

}

func (handler SubscribeHandler) Delete(c *gin.Context) {
	subIp := c.Param("ip")
	translatorIp, exists := c.GetQuery("translatorIp")

	if exists {
		err := handler.Repo.DeleteOne(subIp, translatorIp)

		if err != nil {
			handler.Log.Error("Error when deleting subscription", zap.Error(err))
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, "Subscription didnot exist")
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	} else {
		err := handler.Repo.DeleteMany(subIp)
		if err != nil {
			handler.Log.Error("Error when deleting subscription", zap.Error(err))
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, "Subscription didnot exist")
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}

	c.JSON(http.StatusOK, "Deleted")

}

func (handler SubscribeHandler) GetMethods() []dto.Routes {
	return []dto.Routes{
		{
			Method:      http.MethodGet,
			Pattern:     "/api/v1/sub/:ip",
			HandlerFunc: handler.Get,
		},

		{
			Method:      http.MethodDelete,
			Pattern:     "/api/v1/sub/:ip",
			HandlerFunc: handler.Delete,
		},

		{
			Method:      http.MethodPost,
			Pattern:     "/api/v1/sub",
			HandlerFunc: handler.Post,
		},
	}
}
