package chaincode

import (
	"chaincode_go/utils"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

/*
钱包
Balance：余额
*/
type Wallet struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
}

/*
公钥环
在用户初始化时将加入环
需要环签名公钥时，从中随机选取指定大小的公钥环
*/
type Ring struct {
	Pubs []string `json:"pubs"`
}

/*
商品结构体
ID、Owner(拥有者)、价格（Price）、Amount（数量）
*/
type Goods struct {
	ID     string `json:"id"`
	Owner  string `json:"owner"`
	Price  int64  `json:"price"`
	Amount int64  `json:"amount"`
	Status int64  `json:"status"` //状态，0无操作；1售卖中；2锁定中
}

/*
Proposal结构体
*/
type Proposal struct {
	OrderNum   string `json:"orderNum"`
	GoodId     string `json:"goodId"`     //商品Id(确认当前商品)
	Ciphertext string `json:"ciphertext"` //价格密文Enc(m)
	Sender     string `json:"sender"`
	Reciver    string `json:"reciver"`
	Flag       int64  `json:"flag"`
}

/*
Signature只保存卖方确认交易的签名
*/
type Signature struct {
	OrderNum string `json:"OrderNum"`
	Enc_B_M  string `json:"enc_b_m"`
	Enc_B_B  string `json:"enc_b_b"`
	Sign     string `json:"sign"` //签名 Enc_B(m)||Enc_B(b)||OrderNum||Add_A
	Address  string `json:"address"`
}

type Order struct {
	OrderNum     string `json:"orderNum"`
	GoodId       string `json:"goodId"`
	CommB        string `json:"commB"`        //卖方对价格的承诺
	Sign_CommB   string `json:"sign_commB"`   //对承诺的签名
	Sign_Confirm string `json:"sign_confirm"` //卖方确认签名
	CommA        string `json:"commA"`        //买方对价格的承诺
	Seller_Opt   int64  `json:"seller_opt"`   //卖方对方案的确认；0未操作；1同意；2拒绝
	Sign_CommA   string `json:"sign_commA"`   //对承诺的签名
	RP_m         string `json:"rp_m"`         //交易金额大于0的承诺
	RP_b         string `json:"rp_b"`         //余额不小于0的承诺
	Link_sign_1  string `json:"link_sign_1"`  //可链接环签名1 Enc_A(m)||Enc_B(m)||Enc_A(b)
	Link_sign_2  string `json:"link_sign_2"`  //可链接环签名2 Add_A||Add_B||OrderNum||Sign_B
	Enc_B_M      string `json:"enc_b_m"`      //卖方公钥加密价格
	Enc_B_B      string `json:"enc_b_b"`      //卖方加密余额
	Enc_A_M      string `json:"enc_a_m"`      //买方公钥加密价格
	Enc_A_B      string `json:"enc_a_b"`      //卖方余额加密
	Enc_S_Add_A  string `json:"enc_s_add_a"`  //用CA公钥加密买方地址
	Enc_S_Add_B  string `json:"enc_s_add_b"`  //用CA公钥加密卖方地址
	Buyer        string `json:"buyer"`        //买方的地址
	Seller       string `json:"seller"`       //卖方的地址
	Pubs         string `json:"pubs"`         //环公钥
	Flag         bool   `json:"flag"`         //订单标志ture已完成 false未完成
}

type Commit struct {
	Comm       string `json:"comm"`
	ProposalId string `json:"proposalId"`
	Address    string `json:"address"`
	Sign       string `json:"sign"`
}

const (
	Proposalkey  = "proposal-key" //复合主键
	Signaturekey = "signature-key"
	CommitKey    = "commit-key"
)

/*
测试连接函数、启动链码成功，进行查询，返回hello
*/
func (s *SmartContract) Hello(ctx contractapi.TransactionContextInterface) string {
	return "hello"
}

/*
初始化账本，将两个商品加入账本
并初始化公钥环
*/
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) (string, error) {
	ids := [2]string{"10000", "10001"}
	owners := [2]string{"Bob", "Alice"}
	prices := [2]int64{100, 50}
	amounts := [2]int64{20, 30}
	for i, v := range ids {
		good := Goods{
			ID:     v,
			Owner:  owners[i],
			Price:  prices[i],
			Amount: amounts[i],
			Status: 0,
		}
		res, err := json.Marshal(good)
		if err != nil {
			return "", err
		}
		if err := ctx.GetStub().PutState(v, []byte(res)); err != nil {
			return "", err
		}
	}
	ring := Ring{Pubs: []string{"", "eyJYIjo2NzkwOTAwNzUxNTkxMDY5NDczNDUyNzE3OTU2Nzc5NTExMDA1Njg5MDcwODUxODY0Mzc5NTE2MDgwNTAxNDg2ODk2ODM2NTMwMDQ1LCJZIjozNDM2MTc5NzY4NTYxNDc3NjUxOTAzNDYwODU5MjY3MjE5Mzc0MTE1NzI4MTQ3MjczNTM5MzY0MDk5NjUyMDA2NDg5NzAyNjE2MzQ3MiwiQ3VydmUiOnsiUCI6MTE1NzkyMDg5MjEwMzU2MjQ4NzU2NDIwMzQ1MjE0MDIwODkyNzY2MjUwMzUzOTkxOTI0MTkxNDU0NDIxMTkzOTMzMjg5Njg0OTkxOTk5LCJOIjoxMTU3OTIwODkyMTAzNTYyNDg3NTY0MjAzNDUyMTQwMjA4OTI3NjYwNjE2MjM3MjQ5NTc3NDQ1Njc4NDM4MDkzNTYyOTM0MzkwNDU5MjMsIkIiOjE4NTA1OTE5MDIyMjgxODgwMTEzMDcyOTgxODI3OTU1NjM5MjIxNDU4NDQ4NTc4MDEyMDc1MjU0ODU3MzQ2MTk2MTAzMDY5MTc1NDQzLCJHeCI6MjI5NjMxNDY1NDcyMzcwNTA1NTk0Nzk1MzEzNjI1NTAwNzQ1Nzg4MDI1NjcyOTUzNDE2MTY5NzAzNzUxOTQ4NDA2MDQxMzk2MTU0MzEsIkd5Ijo4NTEzMjM2OTIwOTgyODU2ODgyNTYxODk5MDYxNzExMjQ5NjQxMzA4ODM4ODYzMTkwNDUwNTA4MzI4MzUzNjYwNzU4ODg3NzIwMTU2OCwiQml0U2l6ZSI6MjU2LCJOYW1lIjoiU00yLVAtMjU2LVYxIiwiQSI6MTE1NzkyMDg5MjEwMzU2MjQ4NzU2NDIwMzQ1MjE0MDIwODkyNzY2MjUwMzUzOTkxOTI0MTkxNDU0NDIxMTkzOTMzMjg5Njg0OTkxOTk2fX0=", "eyJYIjo1OTA1MjE3MjgxMTkyMTU1MzQ2NjMwNzUzNDQxNDQyMDEwMzc0NjgzOTgwMTAwNzAwMTcwNDgyMzc3MzUyODgxMzExNDU4OTc0OTI4NSwiWSI6MTAyMTU2Nzc4NjU5ODIxNTA3NDg5MDE2MDg3NDE0ODY2NzQ2MzcyNDg2Njk2MzM0MjI3NDM2ODcxNDExOTExMjUwNjkyMzQ5ODAzNTYsIkN1cnZlIjp7IlAiOjExNTc5MjA4OTIxMDM1NjI0ODc1NjQyMDM0NTIxNDAyMDg5Mjc2NjI1MDM1Mzk5MTkyNDE5MTQ1NDQyMTE5MzkzMzI4OTY4NDk5MTk5OSwiTiI6MTE1NzkyMDg5MjEwMzU2MjQ4NzU2NDIwMzQ1MjE0MDIwODkyNzY2MDYxNjIzNzI0OTU3NzQ0NTY3ODQzODA5MzU2MjkzNDM5MDQ1OTIzLCJCIjoxODUwNTkxOTAyMjI4MTg4MDExMzA3Mjk4MTgyNzk1NTYzOTIyMTQ1ODQ0ODU3ODAxMjA3NTI1NDg1NzM0NjE5NjEwMzA2OTE3NTQ0MywiR3giOjIyOTYzMTQ2NTQ3MjM3MDUwNTU5NDc5NTMxMzYyNTUwMDc0NTc4ODAyNTY3Mjk1MzQxNjE2OTcwMzc1MTk0ODQwNjA0MTM5NjE1NDMxLCJHeSI6ODUxMzIzNjkyMDk4Mjg1Njg4MjU2MTg5OTA2MTcxMTI0OTY0MTMwODgzODg2MzE5MDQ1MDUwODMyODM1MzY2MDc1ODg4NzcyMDE1NjgsIkJpdFNpemUiOjI1NiwiTmFtZSI6IlNNMi1QLTI1Ni1WMSIsIkEiOjExNTc5MjA4OTIxMDM1NjI0ODc1NjQyMDM0NTIxNDAyMDg5Mjc2NjI1MDM1Mzk5MTkyNDE5MTQ1NDQyMTE5MzkzMzI4OTY4NDk5MTk5Nn19", "eyJYIjoxNzA2OTExMjM4NTAzMjU0MDE1NzU2NTQwNzY3Njc2NDQwNTIxMzA5MjEyOTY2MzY1MzgwNjgzNTYzOTg2Nzg1MzUxMzcwODAzOTEwMSwiWSI6NDM5MzU2MzY1NTQ5ODI5Nzc1MDEwNDI4NTAwMjc2MTcwNzE2NzM2ODM0ODI4Nzc5NDg3NTM3MjQ4NTYxMTcwMDg2MzA3OTk4MjEwNjksIkN1cnZlIjp7IlAiOjExNTc5MjA4OTIxMDM1NjI0ODc1NjQyMDM0NTIxNDAyMDg5Mjc2NjI1MDM1Mzk5MTkyNDE5MTQ1NDQyMTE5MzkzMzI4OTY4NDk5MTk5OSwiTiI6MTE1NzkyMDg5MjEwMzU2MjQ4NzU2NDIwMzQ1MjE0MDIwODkyNzY2MDYxNjIzNzI0OTU3NzQ0NTY3ODQzODA5MzU2MjkzNDM5MDQ1OTIzLCJCIjoxODUwNTkxOTAyMjI4MTg4MDExMzA3Mjk4MTgyNzk1NTYzOTIyMTQ1ODQ0ODU3ODAxMjA3NTI1NDg1NzM0NjE5NjEwMzA2OTE3NTQ0MywiR3giOjIyOTYzMTQ2NTQ3MjM3MDUwNTU5NDc5NTMxMzYyNTUwMDc0NTc4ODAyNTY3Mjk1MzQxNjE2OTcwMzc1MTk0ODQwNjA0MTM5NjE1NDMxLCJHeSI6ODUxMzIzNjkyMDk4Mjg1Njg4MjU2MTg5OTA2MTcxMTI0OTY0MTMwODgzODg2MzE5MDQ1MDUwODMyODM1MzY2MDc1ODg4NzcyMDE1NjgsIkJpdFNpemUiOjI1NiwiTmFtZSI6IlNNMi1QLTI1Ni1WMSIsIkEiOjExNTc5MjA4OTIxMDM1NjI0ODc1NjQyMDM0NTIxNDAyMDg5Mjc2NjI1MDM1Mzk5MTkyNDE5MTQ1NDQyMTE5MzkzMzI4OTY4NDk5MTk5Nn19", "eyJYIjoxMDk3NzYzNTc1OTUzNTY2NDE5NzkwMTE0MjU5MzE0NTkzODUxMjMyOTAxOTg3OTI2MjA0OTAzODcxMTgwOTM2NjA5NTY0NDc5ODY0NCwiWSI6NzcxNjkzMjgxNTIxMDQ0OTI4Mzg5NjU3MjgzMDY5NzMyODAzNjY1NTgyMDk3NDM2Mzg2Mzc0OTE4MDI4Mjg3MDkwMDAzMTE0MTQ1NjQsIkN1cnZlIjp7IlAiOjExNTc5MjA4OTIxMDM1NjI0ODc1NjQyMDM0NTIxNDAyMDg5Mjc2NjI1MDM1Mzk5MTkyNDE5MTQ1NDQyMTE5MzkzMzI4OTY4NDk5MTk5OSwiTiI6MTE1NzkyMDg5MjEwMzU2MjQ4NzU2NDIwMzQ1MjE0MDIwODkyNzY2MDYxNjIzNzI0OTU3NzQ0NTY3ODQzODA5MzU2MjkzNDM5MDQ1OTIzLCJCIjoxODUwNTkxOTAyMjI4MTg4MDExMzA3Mjk4MTgyNzk1NTYzOTIyMTQ1ODQ0ODU3ODAxMjA3NTI1NDg1NzM0NjE5NjEwMzA2OTE3NTQ0MywiR3giOjIyOTYzMTQ2NTQ3MjM3MDUwNTU5NDc5NTMxMzYyNTUwMDc0NTc4ODAyNTY3Mjk1MzQxNjE2OTcwMzc1MTk0ODQwNjA0MTM5NjE1NDMxLCJHeSI6ODUxMzIzNjkyMDk4Mjg1Njg4MjU2MTg5OTA2MTcxMTI0OTY0MTMwODgzODg2MzE5MDQ1MDUwODMyODM1MzY2MDc1ODg4NzcyMDE1NjgsIkJpdFNpemUiOjI1NiwiTmFtZSI6IlNNMi1QLTI1Ni1WMSIsIkEiOjExNTc5MjA4OTIxMDM1NjI0ODc1NjQyMDM0NTIxNDAyMDg5Mjc2NjI1MDM1Mzk5MTkyNDE5MTQ1NDQyMTE5MzkzMzI4OTY4NDk5MTk5Nn19", "eyJYIjo2MDQzNTk0MzA2OTE3NTM3Njc1ODg2MTE5NTM2OTE3MjQ5NTIzNTk1NTU0NTE4MjQxMDI3NTM5NjA1MTMyNjQ4NzMyMzUwMDUyODI0MiwiWSI6ODk0MDU4OTAzNzU1Mjk3ODgzMzU4NDEwNzgxNDAwMDEzMDQyNDIxNzQwNjI5OTIzMzYzOTQ0MTk1OTE1MjA5NDc1ODkyNDM4MjE4NTAsIkN1cnZlIjp7IlAiOjExNTc5MjA4OTIxMDM1NjI0ODc1NjQyMDM0NTIxNDAyMDg5Mjc2NjI1MDM1Mzk5MTkyNDE5MTQ1NDQyMTE5MzkzMzI4OTY4NDk5MTk5OSwiTiI6MTE1NzkyMDg5MjEwMzU2MjQ4NzU2NDIwMzQ1MjE0MDIwODkyNzY2MDYxNjIzNzI0OTU3NzQ0NTY3ODQzODA5MzU2MjkzNDM5MDQ1OTIzLCJCIjoxODUwNTkxOTAyMjI4MTg4MDExMzA3Mjk4MTgyNzk1NTYzOTIyMTQ1ODQ0ODU3ODAxMjA3NTI1NDg1NzM0NjE5NjEwMzA2OTE3NTQ0MywiR3giOjIyOTYzMTQ2NTQ3MjM3MDUwNTU5NDc5NTMxMzYyNTUwMDc0NTc4ODAyNTY3Mjk1MzQxNjE2OTcwMzc1MTk0ODQwNjA0MTM5NjE1NDMxLCJHeSI6ODUxMzIzNjkyMDk4Mjg1Njg4MjU2MTg5OTA2MTcxMTI0OTY0MTMwODgzODg2MzE5MDQ1MDUwODMyODM1MzY2MDc1ODg4NzcyMDE1NjgsIkJpdFNpemUiOjI1NiwiTmFtZSI6IlNNMi1QLTI1Ni1WMSIsIkEiOjExNTc5MjA4OTIxMDM1NjI0ODc1NjQyMDM0NTIxNDAyMDg5Mjc2NjI1MDM1Mzk5MTkyNDE5MTQ1NDQyMTE5MzkzMzI4OTY4NDk5MTk5Nn19"}}
	// 将 Ring 结构体序列化并保存到状态
	ringBytes, err := json.Marshal(ring)
	if err != nil {
		return "", err
	}
	if err := ctx.GetStub().PutState("ring", ringBytes); err != nil {
		return "", err
	}
	return "init success", nil
}

/*
设置钱包，余额密文、账户地址都存入账本
*/
func (s *SmartContract) SetWallet(ctx contractapi.TransactionContextInterface, address string, ctext string) (*Wallet, error) {
	exist, err := ctx.GetStub().GetState(address)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if exist != nil {
		return nil, fmt.Errorf("the wallet %s already exists", address)
	}
	wallet := Wallet{
		Address: address,
		Balance: ctext,
	}
	walletJSON, err := json.Marshal(wallet)
	if err != nil {
		return nil, fmt.Errorf("marshal wallet  error:%v", err)
	}
	if err := ctx.GetStub().PutState(address, walletJSON); err != nil {
		return nil, fmt.Errorf("put wallet error:%v", err)
	}
	return &wallet, nil
}

/*
获取钱包信息
*/
func (s *SmartContract) GetWallet(ctx contractapi.TransactionContextInterface, address string) (*Wallet, error) {
	walletJSON, err := ctx.GetStub().GetState(address)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if walletJSON == nil {
		return nil, fmt.Errorf("the wallet %s does not exist", address)
	}

	var wallet Wallet
	err = json.Unmarshal(walletJSON, &wallet)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (s *SmartContract) UpdateWallet(ctx contractapi.TransactionContextInterface, address string, ctext string) (*Wallet, error) {
	res, err := ctx.GetStub().GetState(address)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if res == nil {
		return nil, fmt.Errorf("the wallet %s does not exist", address)
	}

	var wallet Wallet
	err = json.Unmarshal(res, &wallet)
	if err != nil {
		return nil, err
	}
	wallet.Balance = ctext
	walletJSON, err := json.Marshal(wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wallet:%v", err)
	}
	if err = ctx.GetStub().PutState(address, walletJSON); err != nil {
		return nil, fmt.Errorf("failed to update wallet:%v", err)
	}
	return &wallet, nil
}

/*
获取所有商品
*/
func (s *SmartContract) GetAllGoods(ctx contractapi.TransactionContextInterface) ([]*Goods, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("10000", "11111")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var goods []*Goods
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var good Goods
		err = json.Unmarshal(queryResponse.Value, &good)
		if err != nil {
			return nil, err
		}
		goods = append(goods, &good)
	}
	return goods, nil
}

/*
根据商品ID获取商品
*/
func (s *SmartContract) GetGoods(ctx contractapi.TransactionContextInterface, id string) (*Goods, error) {
	goodJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if goodJSON == nil {
		return nil, fmt.Errorf("the good %s does not exist", id)
	}

	var good Goods
	err = json.Unmarshal(goodJSON, &good)
	if err != nil {
		return nil, err
	}
	return &good, nil
}

/*
根据拥有者获取商品
*/
func (s *SmartContract) GetGoodsByOwner(ctx contractapi.TransactionContextInterface, owner string) ([]*Goods, error) {
	if owner == "" {
		return nil, fmt.Errorf("the owner %s is nil", owner)
	}
	queryString := fmt.Sprintf("{\"selector\":{\"owner\":\"%s\"}}", owner)
	result, err := ctx.GetStub().GetQueryResult(queryString)
	var goodList []*Goods
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state:%v", err)
	}
	defer result.Close()
	// 遍历迭代器
	for result.HasNext() {
		// 获取下一个查询结果
		queryResult, err := result.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over query results: %v", err)
		}

		// 将结果反序列化为 Proposal 结构体
		var good Goods
		err = json.Unmarshal(queryResult.Value, &good)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal proposal: %v", err)
		}
		// 将解析后的 Proposal 添加到列表中
		goodList = append(goodList, &good)
	}
	// 返回 good 列表
	return goodList, nil
}

func (s *SmartContract) UpdateGoodStatus(ctx contractapi.TransactionContextInterface, id string) (*Goods, error) {
	if id == "" {
		return nil, fmt.Errorf("the args id is null")
	}
	good, err := s.GetGoods(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state:%v", err)
	}
	good.Status = 2
	goodJSON, err := json.Marshal(good)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proposal:%v", err)
	}
	if err := ctx.GetStub().PutState(id, goodJSON); err != nil {
		return nil, fmt.Errorf("failed to put state:%v", err)
	}
	return good, nil
}

func (s *SmartContract) UpdateGoodPrice(ctx contractapi.TransactionContextInterface, id string, price_str string) (*Goods, error) {
	if id == "" {
		return nil, fmt.Errorf("the args id is null")
	}
	good, err := s.GetGoods(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state:%v", err)
	}
	price, err := strconv.ParseInt(price_str, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to prase price int64:%v", err)
	}
	good.Price = price
	good.Status = 1
	goodJSON, err := json.Marshal(good)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proposal:%v", err)
	}
	if err := ctx.GetStub().PutState(id, goodJSON); err != nil {
		return nil, fmt.Errorf("failed to put state:%v", err)
	}
	return good, nil
}

/*
构造初始的交易提案，供交易接收方确认
proposalId：提案ID
sender: 发送者钱包hash地址
reciver：接收者钱包hash地址
ctext：使用接收者公钥加密的密文
*/
func (s *SmartContract) SetProposal(ctx contractapi.TransactionContextInterface, orderNum string, buyer string, seller string, ctext string, goodid string) (*Order, error) {
	exist, err := ctx.GetStub().GetState(orderNum)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if exist != nil {
		return nil, fmt.Errorf("the proposal %s is exists", orderNum)
	}
	order := Order{
		OrderNum:   orderNum,
		GoodId:     goodid,
		Enc_B_M:    ctext,
		Buyer:      buyer,
		Seller:     seller,
		Seller_Opt: 0,
	}
	_, err = s.UpdateGoodStatus(ctx, goodid)
	if err != nil {
		return nil, fmt.Errorf("update good status:%v", err)
	}
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proposal:%v", err)
	}
	// 写入账本
	if err := ctx.GetStub().PutState(order.OrderNum, orderJSON); err != nil {
		return nil, fmt.Errorf("failed to put proposal in state:%v", err)
	}
	return &order, nil
}

/*
根据ProposalId获取
*/
func (s *SmartContract) GetProposal(ctx contractapi.TransactionContextInterface, orderNum string) (*Order, error) {
	res, err := ctx.GetStub().GetState(orderNum)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if res == nil {
		return nil, fmt.Errorf("the proposal %s does not exist", orderNum)
	}

	var order Order
	err = json.Unmarshal(res, &order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

/*
sender
proposalId：提案Id
获取提案，返回前端进行确认
*/
func (s *SmartContract) GetProposalByBuyer(ctx contractapi.TransactionContextInterface, buyer string) ([]*Order, error) {
	queryString := fmt.Sprintf("{\"selector\":{\"buyer\":\"%s\"}}", buyer)
	result, err := ctx.GetStub().GetQueryResult(queryString) //必须是CouchDB才行
	var orders []*Order
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	defer result.Close()
	// 遍历迭代器
	for result.HasNext() {
		// 获取下一个查询结果
		queryResult, err := result.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over query results: %v", err)
		}

		// 将结果反序列化为 Proposal 结构体
		var order Order
		err = json.Unmarshal(queryResult.Value, &order)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal proposal: %v", err)
		}
		// 将解析后的 Proposal 添加到列表中
		orders = append(orders, &order)
	}
	// 返回 order 列表
	return orders, nil
}

/*
Reciver
proposalId：提案Id
获取提案，返回前端进行确认
*/
func (s *SmartContract) GetProposalBySeller(ctx contractapi.TransactionContextInterface, seller string) ([]*Order, error) {
	queryString := fmt.Sprintf("{\"selector\":{\"seller\":\"%s\"}}", seller)
	result, err := ctx.GetStub().GetQueryResult(queryString) //必须是CouchDB才行
	var orders []*Order
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	defer result.Close()
	// 遍历迭代器
	for result.HasNext() {
		// 获取下一个查询结果
		queryResult, err := result.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over query results: %v", err)
		}

		// 将结果反序列化为 Proposal 结构体
		var order Order
		err = json.Unmarshal(queryResult.Value, &order)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal proposal: %v", err)
		}
		// 将解析后的 Proposal 添加到列表中
		orders = append(orders, &order)
	}
	// 返回 order 列表
	return orders, nil
}

func (s *SmartContract) BuyerSetCommit(ctx contractapi.TransactionContextInterface, orderNum string, comm string, sign string) (*Order, error) {
	exist, err := ctx.GetStub().GetState(orderNum)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if exist == nil {
		return nil, fmt.Errorf("the order %s is not exists", orderNum)
	}
	var order Order
	if err := json.Unmarshal(exist, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order:%v", err)
	}
	order.CommA = comm
	order.Sign_CommA = sign
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order:%v", err)
	}
	if err := ctx.GetStub().PutState(orderNum, orderJSON); err != nil {
		return nil, fmt.Errorf("failed to put order: %v", err)
	}
	return &order, nil
}

func (s *SmartContract) SellerSetCommit(ctx contractapi.TransactionContextInterface, orderNum string, comm string, sign string) (*Order, error) {
	exist, err := ctx.GetStub().GetState(orderNum)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if exist == nil {
		return nil, fmt.Errorf("the order %s is not exists", orderNum)
	}
	var order Order
	if err := json.Unmarshal(exist, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order:%v", err)
	}
	order.CommB = comm
	order.Sign_CommB = sign
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order:%v", err)
	}
	if err := ctx.GetStub().PutState(orderNum, orderJSON); err != nil {
		return nil, fmt.Errorf("failed to put order: %v", err)
	}
	return &order, nil
}

func (s *SmartContract) GetCommit(ctx contractapi.TransactionContextInterface, proposalId string, address string) (*Commit, error) {
	results, err := utils.GetStateByPartialCompositeKeys2(ctx, CommitKey, []string{proposalId, address})
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	if len(results) != 1 {
		return nil, fmt.Errorf("the commit is not exists")
	}
	var comm Commit
	if err := json.Unmarshal(results[0], &comm); err != nil {
		return nil, fmt.Errorf("failed to unmarshal commit")
	}
	return &comm, nil
}

/*
Proposal
proposalId：提案Id
获取提案，返回前端进行确认
*/
func (s *SmartContract) GetProposalByProposalId(ctx contractapi.TransactionContextInterface, OrderNum string) ([]*Order, error) {
	queryString := fmt.Sprintf("{\"selector\":{\"orderNum\":\"%s\"}}", OrderNum)
	result, err := ctx.GetStub().GetQueryResult(queryString) //必须是CouchDB才行
	var orders []*Order
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	defer result.Close()
	// 遍历迭代器
	for result.HasNext() {
		// 获取下一个查询结果
		queryResult, err := result.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over query results: %v", err)
		}
		// 将结果反序列化为 Proposal 结构体
		var order Order
		err = json.Unmarshal(queryResult.Value, &order)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal proposal: %v", err)
		}
		// 将解析后的 Proposal 添加到列表中
		orders = append(orders, &order)
	}
	// 返回 order 列表
	return orders, nil
}
func (s *SmartContract) CancelProposal(ctx contractapi.TransactionContextInterface, orderNum string, flag_str string) (*Order, error) {
	flag, err := strconv.ParseInt(flag_str, 10, 2)
	if err != nil {
		return nil, fmt.Errorf("strconv parseInt flag_str:%v", err)
	}
	res, err := ctx.GetStub().GetState(orderNum)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if res == nil {
		return nil, fmt.Errorf("the proposal %s does not exist", orderNum)
	}
	var order Order
	err = json.Unmarshal(res, &order)
	if err != nil {
		return nil, err
	}
	order.Seller_Opt = flag
	goodId := order.GoodId
	res1, err := ctx.GetStub().GetState(goodId)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if res1 == nil {
		return nil, fmt.Errorf("the good %s does not exist", goodId)
	}
	var good Goods
	err = json.Unmarshal(res1, &good)
	if err != nil {
		return nil, err
	}
	good.Status = 1
	goodJSON, err := json.Marshal(good)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proposal:%v", err)
	}
	if err := ctx.GetStub().PutState(goodId, goodJSON); err != nil {
		return nil, fmt.Errorf("failed to put state:%v", err)
	}
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal Proposal: %v", err)
	}
	if err := ctx.GetStub().PutState(order.OrderNum, orderJSON); err != nil {
		return nil, fmt.Errorf("put proposal error:%v", err)
	}
	return &order, nil

}

/*
proposalId：提案Id
flag：要执行的操作
修改Proposal的状态是同意
*/
func (s *SmartContract) UpdateProposal(ctx contractapi.TransactionContextInterface, orderNum string, flag_str string) (*Order, error) {
	flag, err := strconv.ParseInt(flag_str, 10, 2)
	if err != nil {
		return nil, fmt.Errorf("strconv parseInt flag_str:%v", err)
	}
	res, err := ctx.GetStub().GetState(orderNum)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if res == nil {
		return nil, fmt.Errorf("the proposal %s does not exist", orderNum)
	}

	var order Order
	err = json.Unmarshal(res, &order)
	if err != nil {
		return nil, err
	}
	order.Seller_Opt = flag
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal Proposal: %v", err)
	}
	if err := ctx.GetStub().PutState(order.OrderNum, orderJSON); err != nil {
		return nil, fmt.Errorf("put proposal error:%v", err)
	}
	return &order, nil
}

/*
交易确认后
s：生成的签名，address：接收人的钱包地址
*/
func (s *SmartContract) SetSignature(ctx contractapi.TransactionContextInterface, orderNum string, signature string, encb string) (*Order, error) {
	exist, err := ctx.GetStub().GetState(orderNum)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if exist == nil {
		return nil, fmt.Errorf("the order %s is not exists", orderNum)
	}
	var order Order
	if err := json.Unmarshal(exist, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order:%v", err)
	}
	order.Enc_B_B = encb
	order.Sign_Confirm = signature
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order:%v", err)
	}
	if err := ctx.GetStub().PutState(orderNum, orderJSON); err != nil {
		return nil, fmt.Errorf("failed to put order: %v", err)
	}
	return &order, nil
}

func (s *SmartContract) GetSignature(ctx contractapi.TransactionContextInterface, proposalId string, address string) (string, error) {
	results, err := utils.GetStateByPartialCompositeKeys2(ctx, Signaturekey, []string{proposalId, address})
	if err != nil {
		return "", fmt.Errorf("%s", err)
	}
	if len(results) != 1 {
		return "", fmt.Errorf("the signature is not exists")
	}
	return string(results[0]), nil
}

func (s *SmartContract) SetOrder(ctx contractapi.TransactionContextInterface, OrderNum string, Enc_A_B string, Enc_A_M string, RP_m string, RP_b string, Link_sign_1 string, Link_sign_2 string, Enc_S_Add_B string, Enc_S_Add_A string, ring_string string) (*Order, error) {
	exist, err := ctx.GetStub().GetState(OrderNum)
	if err != nil {
		return nil, fmt.Errorf("failed to read state: %v", err)
	}
	if exist == nil {
		return nil, fmt.Errorf("the order %s is not exist", OrderNum)
	}
	// ring_byte, err := hex.DecodeString(ring_string)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to decode ring string: %v", err)
	// }
	// var ring []string
	// if err := json.Unmarshal(ring_byte, &ring); err != nil {
	// 	return nil, fmt.Errorf("failed to decode ring_bytes: %v", err)
	// }
	var order Order
	if err := json.Unmarshal(exist, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order:%v", err)
	}
	order.Enc_A_B = Enc_A_B
	order.Enc_A_M = Enc_A_M
	order.RP_m = RP_m
	order.RP_b = RP_b
	order.Link_sign_1 = Link_sign_1
	order.Link_sign_2 = Link_sign_2
	order.Enc_S_Add_A = Enc_S_Add_A
	order.Enc_S_Add_B = Enc_S_Add_B
	order.Pubs = ring_string
	order.Flag = false
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal order")
	}
	if err = ctx.GetStub().PutState(OrderNum, orderJSON); err != nil {
		return nil, fmt.Errorf("failed to put state:%v", err)
	}
	return &order, nil
}

func (s *SmartContract) UpdateOrder(ctx contractapi.TransactionContextInterface, OrderNum string) (*Order, error) {
	res, err := ctx.GetStub().GetState(OrderNum)
	if err != nil {
		return nil, fmt.Errorf("failed to read state:%v", err)
	}
	var order Order
	if err := json.Unmarshal(res, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order:%v", err)
	}
	order.Flag = true
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order:%v", err)
	}
	if err = ctx.GetStub().PutState(OrderNum, orderJSON); err != nil {
		return nil, fmt.Errorf("failed to update order:%v", err)
	}
	return &order, nil
}

func (s *SmartContract) GetOrder(ctx contractapi.TransactionContextInterface, OrderNum string) (*Order, error) {
	res, err := ctx.GetStub().GetState(OrderNum)
	if err != nil {
		return nil, fmt.Errorf("failed to read state:%v", err)
	}
	var order Order
	if err := json.Unmarshal(res, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order:%v", err)
	}
	return &order, nil
}

// 获取环签名公钥（直接返回五个公钥）
func (s *SmartContract) GetRingPublicKeys(ctx contractapi.TransactionContextInterface) (string, error) {
	pubsBytes, err := ctx.GetStub().GetState("ring")
	if err != nil {
		return "", err
	}

	var ring Ring
	if err := json.Unmarshal(pubsBytes, &ring); err != nil {
		return "", err
	}

	// 确保Pubs数组至少有五个元素，并且跳过第一个元素
	if len(ring.Pubs) > 5 {
		// 只需要前五个公钥（不包括第一个）
		selectedPubs := ring.Pubs[1:6]

		// 将选定的公钥转换为JSON字符串并返回
		selectedPubsBytes, err := json.Marshal(selectedPubs)
		if err != nil {
			return "json marshal public keys error", err
		}
		return string(selectedPubsBytes), nil
	}

	// 如果Pubs数组的元素少于五个（不包括第一个），或者为空，则返回一个错误或空数组
	if len(ring.Pubs) > 1 {
		// 返回除了第一个公钥之外的所有公钥
		remainingPubs := ring.Pubs[1:]

		// 将剩余的公钥转换为JSON字符串并返回
		remainingPubsBytes, err := json.Marshal(remainingPubs)
		if err != nil {
			return "json marshal public keys error", err
		}
		return string(remainingPubsBytes), nil
	}

	// 如果Pubs数组只有一个元素（不包括第一个），或者为空，则返回空数组
	return "[]", nil
}

// // 获取环签名公钥
// func (s *SmartContract) GetRingPublicKeys(ctx contractapi.TransactionContextInterface) (string, error) {
// 	pubsBytes, err := ctx.GetStub().GetState("ring")
// 	if err != nil {
// 		return "", err
// 	}

// 	var ring Ring
// 	if err := json.Unmarshal(pubsBytes, &ring); err != nil {
// 		return "", err
// 	}
// 	// 确保Pubs数组至少有一个元素，并且跳过第一个元素
// 	if len(ring.Pubs) > 1 {
// 		// 排除第一个公钥后的剩余公钥数量
// 		remainingPubs := ring.Pubs[1:]

// 		// Seed the random number generator
// 		rand.Seed(time.Now().UnixNano())

// 		// 如果需要返回固定数量的公钥，例如5个，但不超过剩余公钥的数量
// 		var numKeysToReturn int
// 		if len(remainingPubs) >= 5 {
// 			numKeysToReturn = 5
// 		} else {
// 			numKeysToReturn = len(remainingPubs)
// 		}
// 		// 生成随机索引并返回对应的公钥
// 		randomIndices := generateRandomIndices(len(remainingPubs), numKeysToReturn)
// 		randomPubs := make([]string, 0, numKeysToReturn)
// 		for _, index := range randomIndices {
// 			randomPubs = append(randomPubs, remainingPubs[index])
// 		}
// 		// 将随机公钥转换为JSON字符串并返回
// 		randomPubsBytes, err := json.Marshal(randomPubs)
// 		if err != nil {
// 			return "json marshal public keys error", err
// 		}
// 		return string(randomPubsBytes), nil
// 	}

// 	// 如果Pubs数组只有一个元素（不包括第一个），或者为空，则返回空字符串或错误
// 	return "[]", nil // 或者可以返回一个错误，表示没有足够的公钥可以返回
// }

// // Function to generate 'n' unique random indices between 0 and 'max'
// func generateRandomIndices(max, n int) []int {
// 	// Create a map to store unique indices
// 	indexMap := make(map[int]bool)
// 	for len(indexMap) < n {
// 		index := rand.Intn(max) // Generate a random index
// 		indexMap[index] = true  // Add the index to the map
// 	}

// 	// Convert map keys to slice
// 	indices := make([]int, 0, n)
// 	for index := range indexMap {
// 		indices = append(indices, index)
// 	}
// 	return indices
// }

