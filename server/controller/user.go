package controller

import (
	"fmt"
	"server/blockchain"

	"github.com/gin-gonic/gin"
	// "server/utils"
)

type UserController struct{}

func (u UserController) Login(c *gin.Context) {
	Success(c, 0, "success", "login", 1)
}

func (u UserController) SetWallet(ctx *gin.Context) {
	//定义匿名结构体，字段与json字段对应
	var body struct {
		Username string `json:"username"`
		Amount   int64  `json:"amount"`
	}

	//绑定json和结构体
	if err := ctx.BindJSON(&body); err != nil {
		return
	}
	//获取json中的key,注意使用 . 访问
	username := body.Username
	amount := body.Amount
	contractInstance := blockchain.GetContractInstance()
	res, err := contractInstance.SetWallet(username, amount)
	if err == nil {
		Success(ctx, 200, "SUCCESS", string(res), 1)
		return
	}
	Error(ctx, 400, fmt.Sprintf("failed to set wallet:%v", err))
}

func (u UserController) GetWallet(ctx *gin.Context) {
	//定义匿名结构体，字段与json字段对应
	var body struct {
		Username string `json:"username"`
	}

	//绑定json和结构体
	if err := ctx.BindJSON(&body); err != nil {
		Error(ctx, 400, fmt.Sprintf("faild to bind body json:%v", err))
		return
	}
	//获取json中的key,注意使用 . 访问
	username := body.Username
	contractInstance := blockchain.GetContractInstance()
	res, err := contractInstance.GetWallet(username)
	if err == nil {
		Success(ctx, 200, "SUCCESS", string(res), 1)
		return
	}
	Error(ctx, 400, fmt.Sprintf("Failed to Submit transaction: %v", err))
}

func (u UserController) SubmitProposal(ctx *gin.Context) {
	var body struct {
		OrderNum string `json:"orderNum"`
		Buyer    string `json:"buyer"`
		Seller   string `json:"seller"`
		Price    string `json:"price"`
	}
	//绑定json和结构体
	if err := ctx.BindJSON(&body); err != nil {
		Error(ctx, 400, fmt.Sprintf("faild to bind body json:%v", err))
		return
	}
	//获取json中的key,注意使用 . 访问
	orderNum := body.OrderNum
	buyer := body.Buyer
	seller := body.Seller
	price := body.Price
	contractInstance := blockchain.GetContractInstance()
	res, err := contractInstance.SubmitProposal(orderNum, buyer, seller, price)
	if err == nil {
		Success(ctx, 200, "SUCCESS", string(res), 1)
		return
	}
	Error(ctx, 400, fmt.Sprintf("Failed to Submit: %v", err))
}

func (u UserController) UpdateOrder(ctx *gin.Context) {
	var body struct {
		OrderNum string `json:"orderNum"`
		Buyer    string `json:"buyer"`
		Seller   string `json:"seller"`
	}
	//绑定json和结构体
	if err := ctx.BindJSON(&body); err != nil {
		Error(ctx, 400, fmt.Sprintf("faild to bind body json:%v", err))
		return
	}
	//获取json中的key,注意使用 . 访问
	orderNum := body.OrderNum
	buyer := body.Buyer
	seller := body.Seller
	contractInstance := blockchain.GetContractInstance()
	res, err := contractInstance.UpdateOrder(orderNum, buyer, seller)
	if err == nil {
		Success(ctx, 200, "SUCCESS", string(res), 1)
		return
	}
	Error(ctx, 400, fmt.Sprintf("Failed to Submit: %v", err))
}
