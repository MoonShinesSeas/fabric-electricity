package controller

import (
	"encoding/json"
	"fmt"
	"server/blockchain"
	"server/utils"

	"github.com/gin-gonic/gin"
)

type ProposalController struct{}

func (p ProposalController) SetProposal(ctx *gin.Context) {
	//定义匿名结构体，字段与json字段对应
	var body struct {
		Buyer  string `json:"buyer"`
		Seller string `json:"seller"`
		Price  string `json:"price"`
		GoodId string `json:"goodId"`
	}

	//绑定json和结构体
	if err := ctx.BindJSON(&body); err != nil {
		Error(ctx, 400, fmt.Sprintf("faild to bind body json:%v", err))
		return
	}
	//获取json中的key,注意使用 . 访问
	buyer := body.Buyer
	seller := body.Seller
	price := body.Price
	goodId := body.GoodId
	contractInstance := blockchain.GetContractInstance()
	res, err := contractInstance.SetProposal(buyer, seller, price, goodId)
	if err == nil {
		Success(ctx, 200, "SUCCESS", string(res), 1)
		return
	}
	Error(ctx, 400, fmt.Sprintf("Failed to Submit transcation:%v", err))
}

func (p ProposalController) GetProposal(ctx *gin.Context) {
	var body struct {
		OrderNum string `json:"orderNum"`
	}
	//绑定json和结构体
	if err := ctx.BindJSON(&body); err != nil {
		Error(ctx, 400, fmt.Sprintf("faild to bind body json:%v", err))
		return
	}
	//获取json中的key,注意使用 . 访问
	arg := body.OrderNum
	contractInstance := blockchain.GetContractInstance()
	res, err := contractInstance.GetProposal(arg)
	if err == nil {
		Success(ctx, 200, "SUCCESS", string(res), 1)
		return
	}
	Error(ctx, 400, fmt.Sprintf("Failed to Submit transaction: %v", err))
}

func (p ProposalController) GetProposalBySeller(ctx *gin.Context) {
	//定义匿名结构体，字段与json字段对应
	var body struct {
		Seller string `json:"seller"`
	}

	//绑定json和结构体
	if err := ctx.BindJSON(&body); err != nil {
		Error(ctx, 400, fmt.Sprintf("faild to bind body %s json:%v", body.Seller,err))
		return
	}
	//获取json中的key,注意使用 . 访问
	seller := body.Seller
	pri := utils.ReadPriKey(seller)
	contractInstance := blockchain.GetContractInstance()
	res, err := contractInstance.GetProposalByReciver(seller)
	if err != nil {
		Error(ctx, 400, fmt.Sprintf("faild to submit:%v", err))
		return
	}
	//res 是一个 []Proposal 序列化后的 JSON 字节数组
	var orders []blockchain.Order
	if err := json.Unmarshal(res, &orders); err != nil {
		Error(ctx, 400, "Failed to Unmarshal proposals")
		fmt.Printf("%v", err)
		return
	}

	// 遍历提案列表，对每一个提案进行解密操作（如果需要）
	for i, order := range orders {
		plaintext, err := utils.DecryptAmount(order.Enc_B_M, pri)
		if err != nil {
			Error(ctx, 400, fmt.Sprintf("Failed to Decrypt Amount for proposal %s", order.Enc_B_M))
			return
		}
		// 更新提案的明文金额
		orders[i].Enc_B_M = plaintext
	}
	// 将更新后的提案列表重新序列化为 JSON 字节数组
	orderBytes, err := json.Marshal(orders)
	if err != nil {
		Error(ctx, 400, "Failed to Marshal proposals")
		return
	}
	// 返回结果
	Success(ctx, 200, "success", string(orderBytes), int64(len(orderBytes)))
}

func (p ProposalController) GetProposalByBuyer(ctx *gin.Context) {
	//定义匿名结构体，字段与json字段对应
	var body struct {
		Buyer string `json:"buyer"`
		// Seller string `json:"seller"`
	}

	//绑定json和结构体
	if err := ctx.BindJSON(&body); err != nil {
		Error(ctx, 400, "faild")
		return
	}
	//获取json中的key,注意使用 . 访问
	buyer := body.Buyer
	// seller := body.Seller
	// pri := utils.ReadPriKey(seller)
	contractInstance := blockchain.GetContractInstance()
	res, err := contractInstance.GetProposalBySender(buyer)
	if err != nil {
		Error(ctx, 400, fmt.Sprintf("failed to getproposal by Id:%v", err))
		return
	}
	// //res 是一个 []Proposal 序列化后的 JSON 字节数组
	// var proposals []blockchain.Proposal
	// if err := json.Unmarshal(res, &proposals); err != nil {
	// 	Error(ctx, 400, "Failed to Unmarshal proposals")
	// 	fmt.Printf("%v", err)
	// 	return
	// }

	// // 遍历提案列表，对每一个提案进行解密操作（如果需要）
	// for i, proposal := range proposals {
	// 	plaintext, err := utils.DecryptAmount(proposal.Ciphertext, pri)
	// 	if err != nil {
	// 		Error(ctx, 400, "Failed to Decrypt Amount for proposal")
	// 		return
	// 	}
	// 	// 更新提案的明文金额
	// 	proposals[i].Ciphertext = plaintext
	// }
	// // 将更新后的提案列表重新序列化为 JSON 字节数组
	// proposalBytes, err := json.Marshal(proposals)
	if err != nil {
		Error(ctx, 400, "Failed to Marshal proposals")
		return
	}
	// 返回结果
	Success(ctx, 200, "success", string(res), int64(len(res)))
}

func (p ProposalController) GetProposalByOrderNum(ctx *gin.Context) {
	//定义匿名结构体，字段与json字段对应
	var body struct {
		OrderNum string `json:"orderNum"`
	}

	//绑定json和结构体
	if err := ctx.BindJSON(&body); err != nil {
		Error(ctx, 400, "Faild to Bind Body Json")
		return
	}
	//获取json中的key,注意使用 . 访问
	orderNum := body.OrderNum
	contractInstance := blockchain.GetContractInstance()
	res, err := contractInstance.GetProposalByProposalId(orderNum)
	if err != nil {
		Error(ctx, 400, "Failed to GetProposal By ProposalId")
		return
	}
	// 返回结果
	Success(ctx, 200, "success", string(res), int64(len(res)))
}

func (p ProposalController) UpdateProposal(ctx *gin.Context) {
	//定义匿名结构体，字段与json字段对应
	var body struct {
		OrderNum string `json:"orderNum"`
		Seller   string `json:"seller"`
		Flag     string `json:"flag"`
	}

	//绑定json和结构体
	if err := ctx.BindJSON(&body); err != nil {
		Error(ctx, 400, fmt.Sprintf("faild to bind body json:%v", err))
		return
	}
	//获取json中的key,注意使用 . 访问
	orderNum := body.OrderNum
	seller := body.Seller
	flag := body.Flag
	contractInstance := blockchain.GetContractInstance()
	res, err := contractInstance.UpdateProposal(seller, orderNum, flag)
	if err == nil {
		Success(ctx, 200, "success", string(res), 1)
		return
	}
	Error(ctx, 400, fmt.Sprintf("Failed to Submit transaction: %v", err))
}
