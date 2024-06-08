package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path/filepath"
	bullet "server/bulletproof/src"
	"server/utils"
	"strconv"
	"sync"

	"github.com/ZZMarquis/gm/sm2"
	"github.com/ZZMarquis/gm/sm3"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

type Contract struct {
	contract *gateway.Contract
}

type Wallet struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
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
	Status int64  `json:"status"`
}
type Proposal struct {
	OrderNum   string `json:"orderNum"`
	GoodId     string `json:"goodId"`     //商品Id(确认当前商品)
	Ciphertext string `json:"ciphertext"` //价格密文Enc(m)
	Sender     string `json:"sender"`
	Reciver    string `json:"reciver"`
	Flag       int64  `json:"flag"`
}

type Signature struct {
	OrderNum string `json:"OrderNum"`
	Enc_B_M  string `json:"enc_b_m"`
	Enc_B_B  string `json:"enc_b_b"`
	Sign     string `json:"sign"` //签名 Enc_B(m)||Enc_B(b)||OrderNum||Add_A
	Address  string `json:"address"`
}

type Commit struct {
	Comm       string `json:"comm"`
	ProposalId string `json:"proposalId"`
	Address    string `json:"address"`
	Sign       string `json:"sign"`
}

type Order struct {
	OrderNum     string `json:"orderNum"`
	GoodId       string `json:"goodId"`
	CommB        string `json:"commB"`        //卖方对价格的承诺
	Sign_CommB   string `json:"sign_commB"`   //对承诺的签名
	Sign_Confirm string `json:"sign_confirm"` //卖方确认签名
	CommA        string `json:"commA"`        //买方对价格的承诺
	Seller_Opt   int64  `json:"seller_opt"`   //卖方对方案的确认
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

var instance *Contract
var once sync.Once

func GetContractInstance() *Contract {
	once.Do(func() {
		instance = &Contract{}
		instance.initialize()
	})
	return instance
}

func (c *Contract) initialize() {
	log.Println("============ application-golang starts ============")

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = c.populateWallet(wallet)
		if err != nil {
			log.Fatalf("failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"go",
		"src",
		"github.com",
		"hyperledger",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		log.Fatalf("failed to get network: %v", err)
	}
	c.contract = network.GetContract("basic")
}

func (c *Contract) populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"..",
		"..",
		"go",
		"src",
		"github.com",
		"hyperledger",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	// certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	certPath := filepath.Join(credPath, "signcerts", "User1@org1.example.com-cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}
	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))
	return wallet.Put("appUser", identity)
}

func (c *Contract) Init() ([]byte, error) {
	result, err := c.contract.SubmitTransaction("InitLedger")
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction: %v", err)
	}
	return result, nil
}

func (c *Contract) Hello() ([]byte, error) {
	result, err := c.contract.SubmitTransaction("Hello")
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction: %v", err)
	}
	return result, nil
}

func (c *Contract) SetWallet(username string, amount int64) ([]byte, error) {
	if err := utils.KeyGen(username); err != nil {
		return nil, err
	}
	address := utils.GetAddress(username)
	pub := utils.ReadPubKey(username)
	ctext, err := utils.EncryptAmount(amount, pub)
	if err != nil {
		return nil, err
	}
	res, err := c.contract.SubmitTransaction("SetWallet", address, ctext)
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction: %v", err)
	}
	return res, nil
}

func (c *Contract) GetWallet(username string) ([]byte, error) {
	address := utils.GetAddress(username)
	result, err := c.contract.EvaluateTransaction("GetWallet", address)
	if err != nil {
		return nil, fmt.Errorf("failed to Evaluate transaction: %v", err)
	}
	var wallet Wallet
	if err := json.Unmarshal(result, &wallet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet %v", err)
	}
	pri := utils.ReadPriKey(username)
	plaintext, err := utils.DecryptAmount(wallet.Balance, pri)
	if err != nil {
		return nil, fmt.Errorf("failed to Decrypt balance %v", err)
	}
	wallet.Balance = plaintext
	walletJSON, err := json.Marshal(wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal wallet %v", err)
	}
	return walletJSON, nil
}

func (c *Contract) GetAllGoods() ([]byte, error) {
	result, err := c.contract.EvaluateTransaction("GetAllGoods")
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate transaction: %v", err)
	}
	return result, nil
}

func (c *Contract) GetGood(id string) ([]byte, error) {
	result, err := c.contract.EvaluateTransaction("GetGoods", id)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate transaction: %v", err)
	}
	return result, nil
}

func (c *Contract) GetGoodByOwner(owner string) ([]byte, error) {
	result, err := c.contract.EvaluateTransaction("GetGoodsByOwner", owner)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate transaction: %v", err)
	}
	return result, nil
}

func (c *Contract) UpdateGoodPrice(id string, price_str string) ([]byte, error) {
	result, err := c.contract.SubmitTransaction("UpdateGoodPrice", id, price_str)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate transaction: %v", err)
	}
	return result, nil
}

func (c *Contract) SetProposal(buyer string, seller string, price_str string, goodId string) ([]byte, error) {
	buyer_address := utils.GetAddress(buyer)
	seller_address := utils.GetAddress(seller)
	pub := utils.ReadPubKey(seller)
	price, err := strconv.ParseInt(price_str, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to prase price int64:%v", err)
	}
	ctext, err := utils.EncryptAmount(price, pub)
	args := buyer + seller + price_str
	h := sm3.New()
	h.Write([]byte(args))
	hashStr := base64.StdEncoding.EncodeToString(h.Sum(nil))
	if err != nil {
		return nil, err
	}
	res, err := c.contract.SubmitTransaction("SetProposal", hashStr, buyer_address, seller_address, ctext, goodId)
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction SetProposal: %v", err)
	}
	return res, nil
}

func (c *Contract) GetProposal(orderNum string) ([]byte, error) {
	res, err := c.contract.EvaluateTransaction("GetProposal", orderNum)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate transaction: %v", err)
	}
	return res, nil
}

func (c *Contract) GetProposalByReciver(seller string) ([]byte, error) {
	reciver := utils.GetAddress(seller)
	res, err := c.contract.EvaluateTransaction("GetProposalBySeller", reciver)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate transaction: %v", err)
	}
	return res, nil
}

func (c *Contract) GetProposalBySender(buyer string) ([]byte, error) {
	sender := utils.GetAddress(buyer)
	res, err := c.contract.EvaluateTransaction("GetProposalByBuyer", sender)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate transaction: %v", err)
	}
	return res, nil
}

func (c *Contract) GetProposalByProposalId(orderNum string) ([]byte, error) {
	res, err := c.contract.EvaluateTransaction("GetProposalByOrderNum", orderNum)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate transaction: %v", err)
	}
	return res, nil
}

func (c *Contract) UpdateProposal(seller string, orderNum string, flag_str string) ([]byte, error) {
	flag, err := strconv.ParseInt(flag_str, 10, 2)
	if err != nil {
		return nil, fmt.Errorf("strconv parseInt flag_str:%v", err)
	}
	if flag == 1 {
		seller_address := utils.GetAddress(seller)
		_, err := c.contract.SubmitTransaction("UpdateProposal", orderNum, flag_str)
		if err != nil {
			return nil, fmt.Errorf("failed to submit transaction: %v", err)
		}
		orderres, err := c.contract.EvaluateTransaction("GetProposal", orderNum)
		if err != nil {
			return nil, fmt.Errorf("failed to Evaluate transaction: %v", err)
		}
		if orderres == nil {
			return nil, fmt.Errorf("the order %s is not exist", orderNum)
		}
		var order Order
		if err := json.Unmarshal(orderres, &order); err != nil {
			return nil, fmt.Errorf("failed to unmarshal order %v", err)
		}
		result, err := c.contract.EvaluateTransaction("GetWallet", seller_address)
		if err != nil {
			return nil, fmt.Errorf("failed to Evaluate transaction: %v", err)
		}
		var wallet Wallet
		if err := json.Unmarshal(result, &wallet); err != nil {
			return nil, fmt.Errorf("failed to unmarshal wallet %v", err)
		}
		pub_seller := utils.ReadPubKey(seller)
		pri_seller := utils.ReadPriKey(seller)
		seller_balance_bytes, err := hex.DecodeString(wallet.Balance)
		if err != nil {
			return nil, fmt.Errorf("failed to Decode String%v", err)
		}
		Enc_B_M_bytes, err := hex.DecodeString(order.Enc_B_M)
		if err != nil {
			return nil, fmt.Errorf("failed to Decode String%v", err)
		}
		Enc_B_B, err := utils.CiperAdd(pub_seller.Curve, seller_balance_bytes, Enc_B_M_bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to Homo sub ciper text %v", err)
		}
		ENC_B_B_string := hex.EncodeToString(Enc_B_B)
		sign_args := append([]byte(order.Enc_B_M), Enc_B_B...)
		sign_args = append(sign_args, []byte(orderNum)...)
		sign_args = append(sign_args, []byte(order.Buyer)...)
		sign_b, err := sm2.Sign(pri_seller, []byte(seller), sign_args)
		if err != nil {
			return nil, fmt.Errorf("failed to sign Enc(m)||Enc(b)||proposalId||SenderAddress byte %v", err)
		}
		sign_b_string := hex.EncodeToString(sign_b)
		_, err = c.contract.SubmitTransaction("SetSignature", orderNum, sign_b_string, ENC_B_B_string)
		if err != nil {
			return nil, fmt.Errorf("failed to submit transcation:%v", err)
		}
		cipherTextByte, err := hex.DecodeString(order.Enc_B_M)
		if err != nil {
			return nil, fmt.Errorf("failed to DecodeString: %v", err)
		}
		plaintext, err := utils.HomoDecrypt(pri_seller, cipherTextByte)
		if err != nil {
			return nil, fmt.Errorf("failed to Decrypt amount: %v", err)
		}
		// 将[]byte转换回int64
		bigInt := new(big.Int).SetBytes(plaintext)
		amount := bigInt.Int64()

		big_price := big.NewInt(amount)
		comm, _ := bullet.PedersenCommit(big_price)
		comm_bytes, err := utils.ECPointToBytes(&comm)
		if err != nil {
			return nil, fmt.Errorf("failed to Marshamal Ecpoint to Bytes%v", err)
		}
		comm_bytes_string := hex.EncodeToString(comm_bytes)
		sign_comm, err := sm2.Sign(pri_seller, []byte(seller), comm_bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to Sign Comm%v", err)
		}
		sign_comm_string := hex.EncodeToString(sign_comm)
		res, err := c.contract.SubmitTransaction("SellerSetCommit", orderNum, comm_bytes_string, sign_comm_string)
		if err != nil {
			return nil, fmt.Errorf("failed to Submit Transcation SetCommit:%v", err)
		}
		return res, nil
	} else if flag == 2 {
		res, err := c.contract.SubmitTransaction("CanelProposal", orderNum, flag_str)
		if err != nil {
			return nil, fmt.Errorf("failed to submit transaction: %v", err)
		}
		return res, nil
	} else {
		return nil, fmt.Errorf("invaild opration")
	}
}

func (c *Contract) GetCommit(proposalId string, seller string) ([]byte, error) {
	address := utils.GetAddress(seller)
	res, err := c.contract.EvaluateTransaction("GetCommit", proposalId, address)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate transaction: %v", err)
	}
	return res, nil
}

func (c *Contract) SubmitProposal(OrderNum string, buyer string, seller string, price_str string) ([]byte, error) {
	buyer_address := utils.GetAddress(buyer)
	walletres, err := c.contract.EvaluateTransaction("GetWallet", buyer_address)
	if err != nil {
		return nil, fmt.Errorf("failed to Evaluate transaction: %v", err)
	}
	var wallet Wallet
	if err := json.Unmarshal(walletres, &wallet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet %v", err)
	}
	orderres, err := c.contract.EvaluateTransaction("GetProposal", OrderNum)
	if err != nil {
		return nil, fmt.Errorf("failed to Evaluate transaction GetProposal: %v", err)
	}
	if orderres == nil {
		return nil, fmt.Errorf("the order %s is not exist", OrderNum)
	}
	seller_address := utils.GetAddress(seller)
	var order Order
	if err := json.Unmarshal(orderres, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet %v", err)
	}
	signres := order.Sign_Confirm
	pub := utils.ReadPubKey(buyer)
	pri := utils.ReadPriKey(buyer)

	pub1 := utils.ReadPubKey(seller)

	sign, err := hex.DecodeString(signres)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature")
	}
	Enc_B_B_string := order.Enc_B_B
	Enc_B_B, err := hex.DecodeString(Enc_B_B_string)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Enc_B_B_string:%v", err)
	}

	b := append([]byte(order.Enc_B_M), Enc_B_B...)
	b = append(b, []byte(OrderNum)...)
	b = append(b, []byte(order.Buyer)...)
	//verify signature
	v := sm2.Verify(pub1, []byte(seller), b, sign)
	if !v {
		return nil, fmt.Errorf("failed to verify signature")
	}
	buyer_balance, err := hex.DecodeString(wallet.Balance)
	if err != nil {
		return nil, fmt.Errorf("failed to Decode String%v", err)
	}

	Enc_B_M, err := hex.DecodeString(order.Enc_B_M)
	if err != nil {
		return nil, fmt.Errorf("failed to Decode String%v", err)
	}
	price, err := strconv.ParseInt(price_str, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price to int:%v", err)
	}
	Enc_A_M_string, err := utils.EncryptAmount(price, pub)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt price:%v", err)
	}
	Enc_A_M, err := hex.DecodeString(Enc_A_M_string)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Enc_A_M_string:%v", err)
	}
	Enc_A_B, err := utils.CiperSub(pub.Curve, buyer_balance, Enc_A_M)
	if err != nil {
		return nil, fmt.Errorf("failed to sub ciper text:%v", err)
	}
	Enc_A_B_string := hex.EncodeToString(Enc_A_B)

	big_price := big.NewInt(price)
	comm, _ := bullet.PedersenCommit(big_price)
	comm_bytes, err := utils.ECPointToBytes(&comm)
	if err != nil {
		return nil, fmt.Errorf("failed to Marshamal Ecpoint to Bytes%v", err)
	}
	comm_bytes_string := hex.EncodeToString(comm_bytes)
	sign_commA, err := sm2.Sign(pri, []byte(buyer), comm_bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to Sign Comm%v", err)
	}
	sign_commA_string := hex.EncodeToString(sign_commA)

	_, err = c.contract.SubmitTransaction("BuyerSetCommit", OrderNum, comm_bytes_string, sign_commA_string)
	if err != nil {
		return nil, fmt.Errorf("failed to submit transcation BuyerSetComit:%v", err)
	}
	//生成交易金额大于零零知识证明证据RP(m)
	bullet.EC = bullet.NewECPrimeGroupKey(8)
	prove := bullet.RPProve(big.NewInt(price))
	prove_bytes, err := utils.RangeProofToBytes(&prove)
	if err != nil {
		return nil, fmt.Errorf("failed to encode prove_bytes%v", err)
	}
	prove_bytes_string := hex.EncodeToString(prove_bytes)
	//生成转账方交易余额大于零零知识证明证据RP(b)
	cipherTextByte, err := hex.DecodeString(wallet.Balance)
	if err != nil {
		return nil, fmt.Errorf("failed to DecodeString: %v", err)
	}
	plaintext, err := utils.HomoDecrypt(pri, cipherTextByte)
	if err != nil {
		return nil, fmt.Errorf("failed to Decrypt amount: %v", err)
	}
	// 将[]byte转换回int64
	bigInt := new(big.Int).SetBytes(plaintext)
	amount := bigInt.Int64()

	bullet.EC = bullet.NewECPrimeGroupKey(64)
	prove1 := bullet.RPProve(big.NewInt(amount - price))

	prove1_bytes, err := utils.RangeProofToBytes(&prove1)
	if err != nil {
		return nil, fmt.Errorf("failed to encode prove1_bytes%v", err)
	}
	prove1_bytes_string := hex.EncodeToString(prove1_bytes)

	pubs, err := c.contract.EvaluateTransaction("GetRingPublicKeys")
	if err != nil {
		return nil, fmt.Errorf("failed to Evaluate Transcation GetRingPublicKeys: %v", err)
	}
	pubs_string := hex.EncodeToString(pubs)
	ring_pubs, err := utils.DecodeKeys(pubs)
	ring_pubs = append(ring_pubs, pub)

	baseSigner := utils.NewBaseLinkableSigner(pri, ring_pubs)
	if err != nil {
		return nil, err
	}
	//Enc_A(m)||Enc_B(m)||Enc_A(b)
	sign1_args := append([]byte(Enc_A_M_string), Enc_B_M...)
	sign1_args = append(sign1_args, Enc_A_B...)

	link_sign1, err := utils.GenerateLinkSign(baseSigner, sign1_args)
	if err != nil {
		return nil, fmt.Errorf("failed to GenerateSign: %v", err)
	}

	sign2_args := append([]byte(order.Buyer), []byte(order.Seller)...)
	sign2_args = append(sign2_args, []byte(order.OrderNum)...)
	sign2_args = append(sign2_args, []byte(order.Sign_Confirm)...)
	link_sign2, err := utils.GenerateLinkSign(baseSigner, sign2_args)
	if err != nil {
		return nil, fmt.Errorf("failed to GenerateSign: %v", err)
	}
	pub_CA := utils.ReadPubKey("CA")

	Enc_Add_A, err := sm2.Encrypt(pub_CA, []byte(buyer_address), sm2.C1C3C2)
	if err != nil {
		return nil, fmt.Errorf("failed to Encrypt: %v", err)
	}
	Enc_Add_A_string := hex.EncodeToString(Enc_Add_A)

	Enc_Add_B, err := sm2.Encrypt(pub_CA, []byte(seller_address), sm2.C1C3C2)
	if err != nil {
		return nil, fmt.Errorf("failed to Encrypt: %v", err)
	}
	Enc_Add_B_string := hex.EncodeToString(Enc_Add_B)
	// OrderNum , Enc_A_B , Enc_A_M , RP_m , RP_b , Link_sign_1 , Link_sign_2 , Enc_S_Add_B , Enc_S_Add_A , ring_string
	res, err := c.contract.SubmitTransaction("SetOrder", OrderNum, Enc_A_B_string, Enc_A_M_string, prove_bytes_string, prove1_bytes_string, link_sign1, link_sign2, Enc_Add_B_string, Enc_Add_A_string, pubs_string)
	if err != nil {
		return nil, fmt.Errorf("failed to Submit Transcation SetSignature: %v", err)
	}
	return res, nil
}

func (c *Contract) UpdateOrder(OrderNum string, buyer string, seller string) ([]byte, error) {
	res, err := c.contract.EvaluateTransaction("GetOrder", OrderNum)
	if err != nil {
		return nil, fmt.Errorf("failed to read state:%v", err)
	}
	var order Order
	if err := json.Unmarshal(res, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order:%v", err)
	}
	pri_CA := utils.ReadPriKey("CA")
	pub_seller := utils.ReadPubKey(seller)
	pub_buyer := utils.ReadPubKey(buyer)
	add_A_bytes, err := hex.DecodeString(order.Enc_S_Add_A)
	if err != nil {
		return nil, fmt.Errorf("failed to decoding add_A:%v", err)
	}
	add_A, err := sm2.Decrypt(pri_CA, add_A_bytes, sm2.C1C3C2)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt add_A:%v", err)
	}

	add_B_bytes, err := hex.DecodeString(order.Enc_S_Add_B)
	if err != nil {
		return nil, fmt.Errorf("failed to decoding add_A:%v", err)
	}
	add_B, err := sm2.Decrypt(pri_CA, add_B_bytes, sm2.C1C3C2)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt add_B:%v", err)
	}
	buyer_address := utils.GetAddress(buyer)
	seller_address := utils.GetAddress(seller)
	buyer_wallet, err := c.contract.EvaluateTransaction("GetWallet", buyer_address)
	if err != nil {
		return nil, fmt.Errorf("failed to Evaluate transaction: %v", err)
	}
	var wallet1 Wallet
	if err := json.Unmarshal(buyer_wallet, &wallet1); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet %v", err)
	}
	enc_a_b_string := wallet1.Balance
	enc_a_b_bytes, err := hex.DecodeString(enc_a_b_string)
	if err != nil {
		return nil, fmt.Errorf("failed to decoding enc_a_a:%v", err)
	}
	seller_wallet, err := c.contract.EvaluateTransaction("GetWallet", seller_address)
	if err != nil {
		return nil, fmt.Errorf("failed to Evaluate transaction: %v", err)
	}
	var wallet2 Wallet
	if err := json.Unmarshal(seller_wallet, &wallet2); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet %v", err)
	}
	enc_b_b_string := wallet2.Balance
	enc_b_b_bytes, err := hex.DecodeString(enc_b_b_string)
	if err != nil {
		return nil, fmt.Errorf("failed to decoding enc_b_b:%v", err)
	}
	pubs, err := hex.DecodeString(order.Pubs)
	if err != nil {
		return nil, fmt.Errorf("failed to decode pub string:%v", err)
	}
	ring_pubs, err := utils.DecodeKeys(pubs)
	if err != nil {
		return nil, err
	}
	ring_pubs = append(ring_pubs, pub_buyer)

	baseVerifyer := utils.NewBaseLinkableVerfier(ring_pubs)
	Enc_B_M_bytes, err := hex.DecodeString(order.Enc_B_M)
	if err != nil {
		return nil, fmt.Errorf("failed to Enc_B_M String%v", err)
	}
	Enc_A_M_bytes, err := hex.DecodeString(order.Enc_A_M)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Enc_A_M string:%v", err)
	}
	Enc_A_B_bytes, err := hex.DecodeString(order.Enc_A_B)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Enc_A_B string:%v", err)
	}
	// sign1_args := append([]byte(Enc_A_M), Enc_B_M...)
	// sign1_args = append(sign1_args, Enc_A_B...)

	// link_sign1, err := utils.GenerateLinkSign(ring_pubs, pri, sign1_args)

	sign1_args := append([]byte(order.Enc_A_M), Enc_B_M_bytes...)
	sign1_args = append(sign1_args, Enc_A_B_bytes...)
	flag1 := utils.LinkSignVerify(baseVerifyer, sign1_args, order.Link_sign_1)
	if !flag1 {
		return nil, fmt.Errorf("failed to verify link sign1")
	}
	// sign2_args := append([]byte(order.Buyer), []byte(order.Seller)...)
	// sign2_args = append(sign2_args, []byte(order.OrderNum)...)
	// sign2_args = append(sign2_args, []byte(order.Sign_Confirm)...)
	// link_sign2, err := utils.GenerateLinkSign(ring_pubs, pri, sign2_args)
	sign2_args := append([]byte(order.Buyer), []byte(order.Seller)...)
	sign2_args = append(sign2_args, []byte(OrderNum)...)
	sign2_args = append(sign2_args, []byte(order.Sign_Confirm)...)
	flag2 := utils.LinkSignVerify(baseVerifyer, sign2_args, order.Link_sign_2)
	if !flag2 {
		return nil, fmt.Errorf("failed to verify link sign2")
	}
	link_sing1 := utils.DecodeSignature(order.Link_sign_1)
	link_sing2 := utils.DecodeSignature(order.Link_sign_2)
	if !utils.Linkable(link_sing1, link_sing2) {
		return nil, fmt.Errorf("signature linkable failure")
	}
	Enc_A_B_, err := utils.CiperSub(pub_buyer.Curve, enc_a_b_bytes, Enc_A_M_bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to homo sub cipertext:%v", err)
	}
	if !bytes.Equal(Enc_A_B_, Enc_A_B_bytes) {
		return nil, fmt.Errorf("failed to Verify amount equal")
	}

	Enc_B_B, err := hex.DecodeString(order.Enc_B_B)
	if err != nil {
		return nil, fmt.Errorf("failed to decode enc_b_b:%v", err)
	}
	// sign_args := append([]byte(order.Enc_B_M), Enc_B_B...)
	// sign_args = append(sign_args, []byte(orderNum)...)
	// sign_args = append(sign_args, []byte(order.Buyer)...)
	// sign_b, err := sm2.Sign(pri_seller, []byte(seller), sign_args)
	sign_args := append([]byte(order.Enc_B_M), Enc_B_B...)
	sign_args = append(sign_args, []byte(order.OrderNum)...)
	sign_args = append(sign_args, []byte(order.Buyer)...)
	sign, err := hex.DecodeString(order.Sign_Confirm)
	if err != nil {
		return nil, fmt.Errorf("failed to decode enc_b_b:%v", err)
	}
	v := sm2.Verify(pub_seller, []byte(seller), sign_args, sign)
	if !v {
		return nil, fmt.Errorf("failed to verify sign_confirm")
	}
	Enc_B_B_, err := utils.CiperAdd(pub_seller.Curve, enc_b_b_bytes, Enc_B_M_bytes)
	if err != nil {
		return nil, fmt.Errorf("ciper add failure:%v", err)
	}
	if !bytes.Equal(Enc_B_B_, Enc_B_B) {
		return nil, fmt.Errorf("failed to Verify B amount equal")
	}
	// big_price := big.NewInt(price)
	// comm, _ := bullet.PedersenCommit(big_price)
	// comm_bytes, err := utils.ECPointToBytes(&comm)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to Marshamal Ecpoint to Bytes%v", err)
	// }
	// comm_bytes_string := hex.EncodeToString(comm_bytes)
	// sign_commA, err := sm2.Sign(pri, []byte(buyer), comm_bytes)
	commA_bytes, err := hex.DecodeString(order.CommA)
	if err != nil {
		return nil, fmt.Errorf("failed to decode comm_a:%v", err)
	}
	sign_commA_bytes, err := hex.DecodeString(order.Sign_CommA)
	if err != nil {
		return nil, fmt.Errorf("failed to decode sign_comm_a:%v", err)
	}
	if !sm2.Verify(pub_buyer, []byte(buyer), commA_bytes, sign_commA_bytes) {
		return nil, fmt.Errorf("failed to verify sign_commA")
	}
	// big_price := big.NewInt(amount)
	// comm, _ := bullet.PedersenCommit(big_price)
	// comm_bytes, err := utils.ECPointToBytes(&comm)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to Marshamal Ecpoint to Bytes%v", err)
	// }
	// comm_bytes_string := hex.EncodeToString(comm_bytes)
	// sign_comm, err := sm2.Sign(pri_seller, []byte(seller), comm_bytes)
	commB_bytes, err := hex.DecodeString(order.CommB)
	if err != nil {
		return nil, fmt.Errorf("failed to decode comm_b:%v", err)
	}
	sign_commB_bytes, err := hex.DecodeString(order.Sign_CommB)
	if err != nil {
		return nil, fmt.Errorf("failed to decode sign_comm_b:%v", err)
	}
	if !sm2.Verify(pub_seller, []byte(seller), commB_bytes, sign_commB_bytes) {
		return nil, fmt.Errorf("failed to verify sign_commB")
	}

	CommA, err := utils.BytesToEcpoint(commA_bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to return commA:%v", err)
	}
	CommB, err := utils.BytesToEcpoint(commB_bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to return commB:%v", err)
	}
	inv_commA := CommA.Neg()
	inv_commB := CommB.Neg()
	if !inv_commA.Add(*CommB).Equal(inv_commB.Add(*CommA)) {
		return nil, fmt.Errorf("commA!=commB")
	}

	bullet.EC = bullet.NewECPrimeGroupKey(8)
	prove_bytes, err := hex.DecodeString(order.RP_m)
	if err != nil {
		return nil, fmt.Errorf("failed to decode rp_m:%v", err)
	}
	prove, err := utils.BytesToRangeProof(prove_bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to return prove:%v", err)
	}
	if !bullet.RPVerify(*prove) {
		return nil, fmt.Errorf("rp_m range proof failure")
	}
	bullet.EC = bullet.NewECPrimeGroupKey(64)
	prove1_bytes, err := hex.DecodeString(order.RP_b)
	if err != nil {
		return nil, fmt.Errorf("failed to decode rp_b:%v", err)
	}
	prove1, err := utils.BytesToRangeProof(prove1_bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to return prove1:%v", err)
	}
	if !bullet.RPVerify(*prove1) {
		return nil, fmt.Errorf("rp_b range proof failure")
	}
	result1, err := c.contract.SubmitTransaction("UpdateWallet", string(add_A), order.Enc_A_B)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet:%v", err)
	}
	result2, err := c.contract.SubmitTransaction("UpdateWallet", string(add_B), order.Enc_B_B)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet:%v", err)
	}
	_, err = c.contract.SubmitTransaction("UpdateOrder", OrderNum)
	if err != nil {
		return nil, fmt.Errorf("failed to UpdateOrder:%v", err)
	}
	return append(result1, result2...), nil
}
