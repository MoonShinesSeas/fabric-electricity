package controller

import (
	"github.com/gin-gonic/gin"
)

type JsonStruct struct{
	Code int `json:"code"`
	Msg interface{} `json:"msg"`
	Data interface{} `json:"data"`
	Count int64 `json:"count"`
}

type JsonErrorStruct struct{
	Code int `json:"code"`
	Msg interface{} `json:"msg"`
}

func Success(c *gin.Context,code int,msg interface{},data interface{},count int64){
	json:=JsonStruct{
		Code:code,
		Msg:msg,
		Data:data,
		Count:count,
	}
	c.JSON(200,json)
}
func Error(c *gin.Context,code int,msg interface{}){
	json:=JsonStruct{
		Code:code,
		Msg:msg,
	}
	c.JSON(200,json)
}