package utils

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"github.com/ZZMarquis/gm/sm2"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// WriteLedger 写入账本
func WriteLedger(obj interface{}, ctx contractapi.TransactionContextInterface, objectType string, keys []string) error {
	//创建复合主键
	var key string
	if val, err := ctx.GetStub().CreateCompositeKey(objectType, keys); err != nil {
		return fmt.Errorf("%s-创建复合主键出错: %s", objectType, err)
	} else {
		key = val
	}
	bytes, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("%s-序列化json数据失败: %s", objectType, err)
	}
	//写入区块链账本
	if err := ctx.GetStub().PutState(key, bytes); err != nil {
		return fmt.Errorf("%s-写入区块链账本出错 %s", objectType, err)
	}
	return nil
}

// GetStateByPartialCompositeKeys 根据复合主键查询数据(适合获取全部或指定的数据)
func GetStateByPartialCompositeKeys2(ctx contractapi.TransactionContextInterface, objectType string, keys []string) (results [][]byte, err error) {
	// 通过主键从区块链查找相关的数据，相当于对主键的模糊查询
	resultIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(objectType, keys)
	if err != nil {
		return nil, fmt.Errorf("%s-获取全部数据出错: %s", objectType, err)
	}
	defer resultIterator.Close()
	//检查返回的数据是否为空，不为空则遍历数据，否则返回空数组
	for resultIterator.HasNext() {
		val, err := resultIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("%s-返回的数据出错: %s", objectType, err)
		}
		results = append(results, val.GetValue())
	}
	return results, nil
}

// GetStateByPartialCompositeKeys 根据复合主键查询数据(适合获取全部，多个，单个数据)
// 将keys拆分查询
func GetStateByPartialCompositeKeys(ctx contractapi.TransactionContextInterface, objectType string, keys []string) (results [][]byte, err error) {
	if len(keys) == 0 {
		// 传入的keys长度为0，则查找并返回所有数据
		// 通过主键从区块链查找相关的数据，相当于对主键的模糊查询
		resultIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(objectType, keys)
		if err != nil {
			return nil, fmt.Errorf("%s-获取全部数据出错: %s", objectType, err)
		}
		defer resultIterator.Close()

		//检查返回的数据是否为空，不为空则遍历数据，否则返回空数组
		for resultIterator.HasNext() {
			val, err := resultIterator.Next()
			if err != nil {
				return nil, fmt.Errorf("%s-返回的数据出错: %s", objectType, err)
			}

			results = append(results, val.GetValue())
		}
	} else {
		// 传入的keys长度不为0，查找相应的数据并返回
		for _, v := range keys {
			// 创建组合键
			key, err := ctx.GetStub().CreateCompositeKey(objectType, []string{v})
			if err != nil {
				return nil, fmt.Errorf("%s-创建组合键出错: %s", objectType, err)
			}
			// 从账本中获取数据
			bytes, err := ctx.GetStub().GetState(key)
			if err != nil {
				return nil, fmt.Errorf("%s-获取数据出错: %s", objectType, err)
			}

			if bytes != nil {
				results = append(results, bytes)
			}
		}
	}
	return results, nil
}

func Base64ToPrivateKey(bytes string) []byte {
	// 解码Base64字符串为原始字节
	privateKeyBytes, err := base64.StdEncoding.DecodeString(bytes)
	if err != nil {
		panic(err)
	}
	// 现在privateKeyBytes就是Base64解码后的SM2私钥的[]byte格式
	return privateKeyBytes
}
func Base64ToPublicKey(bytes string) []byte {
	//编码Base64字符串为原始字节
	publicKeyBytes, err := base64.StdEncoding.DecodeString(bytes)
	if err != nil {
		panic(err)
	}
	return publicKeyBytes
}

func DecodePricateKey(pri_str string) (*sm2.PrivateKey, error) {
	pri_bytes := Base64ToPrivateKey(pri_str)
	var pri *sm2.PrivateKey
	if err := json.Unmarshal(pri_bytes, &pri); err != nil {
		log.Fatalf("Failed to decode private_key: %v", err)
	}
	return pri, nil
}

func DecodePublicKey(pub_str string) (*sm2.PublicKey, error) {
	pub_bytes := Base64ToPublicKey(pub_str)
	var pub *sm2.PublicKey
	if err := json.Unmarshal(pub_bytes, &pub); err != nil {
		log.Fatalf("Failed to decode public_key: %v", err)
	}
	return pub, nil
}

func DecodeCipertext(ciphertext string, pri *sm2.PrivateKey) (int64, error) {
	cipherTextByte, err := hex.DecodeString(ciphertext)
	if err != nil {
		log.Fatalf("failed to DecodeString: %v", err)
	}
	balance_plaintext, err := HomoDecrypt(pri, cipherTextByte)
	if err != nil {
		log.Fatal(err)
	}
	// 将[]byte转换回int64
	bigInt := new(big.Int).SetBytes(balance_plaintext)
	balance := bigInt.Int64()
	return balance, nil
}

/*
密文相加
*/
func AddCiperText(ctext1 string, ctext2 string, pub *sm2.PublicKey) (string, error) {
	hexCiperTextByte1, err := hex.DecodeString(ctext1)
	if err != nil {
		return "", fmt.Errorf("failed to DecodeString: %v", err)
	}
	hexCiperTextByte2, err := hex.DecodeString(ctext2)
	if err != nil {
		return "", fmt.Errorf("failed to DecodeString: %v", err)
	}
	cipertext, err := CiperAdd(pub.Curve, hexCiperTextByte1, hexCiperTextByte2)
	if err != nil {
		return "", fmt.Errorf("failed to CiperAdd: %v", err)
	}
	hexCiperText := hex.EncodeToString(cipertext)
	return hexCiperText, nil
}
