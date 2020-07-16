package controller

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"summerCourse/service"
)

func SelectGoods(ctx *gin.Context) {
	goods := service.SelectGoods()
	ctx.JSON(http.StatusOK, gin.H{
		"status": 200,
		"info": "success",
		"data": struct {
			Goods []service.Goods `json:"goods"`
		}{goods},
	})
}

func AddGoods(ctx *gin.Context) {
	var good service.Goods
	err := ctx.ShouldBindJSON(&good)
	if err != nil {
		ctx.JSON(http.StatusOK,gin.H{
			"status" : 1001,
			"info" : "JSON data unavalible.",
		})
		return
	}
	err = service.AddGoods(good)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusOK,gin.H{
			"status" : 1001,
			"info" : "Add goods failed.",
		})
		return
	}else {
		ctx.JSON(http.StatusOK, gin.H{
			"status": 200,
			"info": "success",
		})
	}
}

 

