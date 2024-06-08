package controller

import (
	"github.com/gin-gonic/gin"
	"server/blockchain"
	"fmt"
)

type GoodController struct{}

func (g GoodController)GetAllGoods(ctx *gin.Context){
	contractInstance := blockchain.GetContractInstance()
	res, err := contractInstance.GetAllGoods()
	if err == nil {
		Success(ctx, 200, "success", string(res), 1)
		return
	}
	
	Error(ctx,400,fmt.Sprintf("Failed to Submit transaction: %v", err))
	return
}
func (g GoodController)GetGood(ctx *gin.Context){
	//定义匿名结构体，字段与json字段对应
	var body struct {
		ID string `json:"id"`
	}
	//绑定json和结构体
	if err := ctx.BindJSON(&body); err != nil {
		Error(ctx,400,fmt.Sprintf("failed to bind body json: %v", err))
		return
	}
	//获取json中的key,注意使用 . 访问
	id:=body.ID
	contractInstance := blockchain.GetContractInstance()
	res, err := contractInstance.GetGood(id)
	if err == nil {
		Success(ctx, 200, "success", string(res), 1)
		return
	}
	Error(ctx,400,fmt.Sprintf("Failed to Submit transaction: %v", err))
	return
}

func (g GoodController)GetGoodByOwner(ctx *gin.Context){
	//定义匿名结构体，字段与json字段对应
	var body struct {
		Owner string `json:"owner"`
	}
	//绑定json和结构体
	if err := ctx.BindJSON(&body); err != nil {
		Error(ctx,400,fmt.Sprintf("failed to bind body json: %v", err))
		return
	}
	//获取json中的key,注意使用 . 访问
	owner:=body.Owner
	contractInstance := blockchain.GetContractInstance()
	res, err := contractInstance.GetGoodByOwner(owner)
	if err == nil {
		Success(ctx, 200, "SUCCESS", string(res), int64(len(res)))
		return
	}
	Error(ctx,400,fmt.Sprintf("Failed to Submit transaction: %v", err))
	return
}

func (g GoodController)UpdateGoodPrice(ctx *gin.Context){
	//定义匿名结构体，字段与json字段对应
	var body struct {
		Id string `json:"id"`
		Price string `json:"price"`
	}
	//绑定json和结构体
	if err := ctx.BindJSON(&body); err != nil {
		Error(ctx,400,fmt.Sprintf("failed to bind body json: %v", err))
		return
	}
	//获取json中的key,注意使用 . 访问
	id:=body.Id
	price:=body.Price
	contractInstance := blockchain.GetContractInstance()
	res, err := contractInstance.UpdateGoodPrice(id,price)
	if err == nil {
		Success(ctx, 200, "SUCCESS", string(res), int64(len(res)))
		return
	}
	Error(ctx,400,fmt.Sprintf("Failed to Submit transaction: %v", err))
	return
}