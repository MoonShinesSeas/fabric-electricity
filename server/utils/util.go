package utils

import (
	bullet "server/bulletproof/src"

	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ZZMarquis/gm/sm2"
	"github.com/ZZMarquis/gm/sm3"
)

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func KeyGen(username string) error {
	// 确保 "key" 文件夹存在
	keyDir := filepath.Join(".", "key")
	err := os.MkdirAll(keyDir, 0755) // 0755表示文件所有者有读/写/执行权限，组用户和其他用户有读/执行权限
	if err != nil {
		return fmt.Errorf("failed to create key directory: %v", err)
	}
	priFileName := filepath.Join(keyDir, fmt.Sprintf("%s-pri", username))
	pubFileName := filepath.Join(keyDir, fmt.Sprintf("%s-pub", username))
	flag1, _ := exists(priFileName)
	flag2, _ := exists(pubFileName)
	if flag1 || flag2 {
		return nil
	}
	pri, pub, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate private,public key:%v", err)
	}
	pri_byte, err := json.Marshal(pri)
	if err != nil {
		return fmt.Errorf("failed to marshal private key:%v", err)
	}
	pub_byte, err := json.Marshal(pub)
	if err != nil {
		return fmt.Errorf("failed to marshal public key:%v", err)
	}

	err = ioutil.WriteFile(priFileName, pri_byte, 0644) // 0600表示只有文件所有者有读写权限
	if err != nil {
		return fmt.Errorf("failed to write private file:%v", err)
	}

	err = ioutil.WriteFile(pubFileName, pub_byte, 0644) // 0644表示文件所有者有读写权限，其他用户有读权限
	if err != nil {
		// 如果公钥写入失败，可能需要删除已写入的私钥文件以保持一致性
		os.Remove(priFileName)
		return fmt.Errorf("failed to write public key:%v", err)
	}
	return nil
}

func ReadPubKey(username string) *sm2.PublicKey {
	// keyDir := filepath.Join(".", "key") // 文件存储在 "key" 文件夹下
	pubFileName := filepath.Join(".", "key", fmt.Sprintf("%s-pub", username))
	// 读取公钥文件
	pubData, err := ioutil.ReadFile(pubFileName)
	if err != nil {
		fmt.Printf("Error reading public key file %s: %v\n", pubFileName, err)
		return nil
	}
	var pub *sm2.PublicKey
	if err := json.Unmarshal(pubData, &pub); err != nil {
		log.Fatalf("Failed to decode public key: %v", err)
	}
	return pub
}

func ReadPriKey(username string) *sm2.PrivateKey {
	keyDir := filepath.Join(".", "key") // 文件存储在 "key" 文件夹下
	priFileName := filepath.Join(keyDir, fmt.Sprintf("%s-pri", username))
	// 读取公钥文件
	priData, err := ioutil.ReadFile(priFileName)
	if err != nil {
		fmt.Printf("Error reading private key file %s: %v\n", priFileName, err)
		return nil
	}
	var pri *sm2.PrivateKey
	if err := json.Unmarshal(priData, &pri); err != nil {
		log.Fatalf("Failed to decode private key: %v", err)
	}
	return pri
}

func GetAddress(username string) string {
	keyDir := filepath.Join(".", "key") // 文件存储在 "key" 文件夹下
	pubFileName := filepath.Join(keyDir, fmt.Sprintf("%s-pub", username))
	// 读取公钥文件
	pubData, err := ioutil.ReadFile(pubFileName)
	if err != nil {
		fmt.Printf("Error reading public key file %s: %v\n", pubFileName, err)
		return ""
	}
	h := sm3.New()
	h.Write(pubData)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// amount
func EncryptAmount(num int64, pub *sm2.PublicKey) (string, error) {
	// 使用bytes.Buffer来存储转换后的字节
	var buf bytes.Buffer
	// 将int64类型的num写入buffer中
	if err := binary.Write(&buf, binary.BigEndian, num); err != nil {
		return "", err
	}
	// cipertext,err:=sm2.Encrypt(pub,buf.Bytes(),sm2.C1C2C3)
	plaintext := buf.Bytes()
	cipertext, err := HomoEncrypt(pub, plaintext)
	if err != nil {
		return "", fmt.Errorf("encrypt amount error:%v", err)
	}
	ctext_str := hex.EncodeToString(cipertext)
	return ctext_str, nil
}
func DecryptAmount(ctext string, pri *sm2.PrivateKey) (string, error) {
	cipherTextByte, err := hex.DecodeString(ctext)
	if err != nil {
		return "", fmt.Errorf("failed to DecodeString: %v", err)
	}
	plaintext, err := HomoDecrypt(pri, cipherTextByte)
	if err != nil {
		return "", fmt.Errorf("failed to Decrypt amount: %v", err)
	}
	// 将[]byte转换回int64
	bigInt := new(big.Int).SetBytes(plaintext)
	amount := bigInt.Int64()
	amount_str := fmt.Sprintf("%d", amount)
	return amount_str, nil
}

// ECPointToBytes 将 ECPoint 结构体序列化为 []byte。
func ECPointToBytes(p *bullet.ECPoint) ([]byte, error) {
	var buf bytes.Buffer

	// 写入 X 的字节表示
	xBytes := p.X.Bytes()
	// 写入 X 的长度（为了之后能够反序列化）
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(xBytes))); err != nil {
		return nil, err
	}
	if _, err := buf.Write(xBytes); err != nil {
		return nil, err
	}

	// 写入 Y 的字节表示
	yBytes := p.Y.Bytes()
	// 写入 Y 的长度（为了之后能够反序列化）
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(yBytes))); err != nil {
		return nil, err
	}
	if _, err := buf.Write(yBytes); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ECPointFromBytes 将 []byte 反序列化为 ECPoint 结构体。
func BytesToEcpoint(data []byte) (*bullet.ECPoint, error) {
	buf := bytes.NewBuffer(data)

	// 读取 X 的长度
	var xLen uint32
	if err := binary.Read(buf, binary.BigEndian, &xLen); err != nil {
		return nil, err
	}

	// 读取 X 的字节表示并转换为 *big.Int
	xBytes := make([]byte, xLen)
	if _, err := buf.Read(xBytes); err != nil {
		return nil, err
	}
	x := new(big.Int).SetBytes(xBytes)

	// 读取 Y 的长度
	var yLen uint32
	if err := binary.Read(buf, binary.BigEndian, &yLen); err != nil {
		return nil, err
	}

	// 读取 Y 的字节表示并转换为 *big.Int
	yBytes := make([]byte, yLen)
	if _, err := buf.Read(yBytes); err != nil {
		return nil, err
	}
	y := new(big.Int).SetBytes(yBytes)

	// 返回重建的 ECPoint 结构体
	return &bullet.ECPoint{X: x, Y: y}, nil
}

func RangeProofToBytes(r *bullet.RangeProof) ([]byte, error) {
	// 使用gob进行序列化
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(r); err != nil {
		return nil, err
	}
	// 现在buf.Bytes()是一个[]byte，它包含了rp的序列化表示
	serializedData := buf.Bytes()
	return serializedData, nil
}

func BytesToRangeProof(b []byte) (*bullet.RangeProof, error) {
	var decodedRp *bullet.RangeProof
	dec := gob.NewDecoder(bytes.NewReader(b))
	if err := dec.Decode(&decodedRp); err != nil {
		return nil, err
	}
	return decodedRp, nil
}
func DecodeKeys(pubs []byte) ([]*sm2.PublicKey, error) {
	var ring []string
	if err := json.Unmarshal(pubs, &ring); err != nil {
		return nil, fmt.Errorf("failed to decode publickeys: %v", err)
	}
	var ring_pubs []*sm2.PublicKey
	for _, v := range ring {
		pub_bytes := base64ToPublicKey(v)
		var pub *sm2.PublicKey
		if err := json.Unmarshal(pub_bytes, &pub); err != nil {
			return nil, fmt.Errorf("unmarshal pubs error: %v", err)
		}
		ring_pubs = append(ring_pubs, pub)
	}
	return ring_pubs, nil
}

func GenerateLinkSign(baseSigner *BaseLinkableSigner, bytes []byte) (string, error) {
	sign, err := baseSigner.Sign(rand.Reader, SimpleParticipantRandInt, bytes)
	if err != nil {
		return "", err
	}
	res_sign := FlodSingature(sign)
	return res_sign, nil
}

func LinkSignVerify(baseVerify *BaseLinkableVerfier, msg []byte, signature string) bool {
	sign := DecodeSignature(signature)
	return baseVerify.Verify(msg, sign)
}

func privateKeyToBase64(b []byte) string {
	base64Str := base64.StdEncoding.EncodeToString(b)
	return base64Str
}
func publicKeyToBase64(b []byte) string {
	base64Str := base64.StdEncoding.EncodeToString(b)
	return base64Str
}

func base64ToPublicKey(decodedBytes string) []byte {
	//编码Base64字符串为原始字节
	publicKeyBytes, err := base64.StdEncoding.DecodeString(decodedBytes)
	if err != nil {
		panic(err)
	}
	return publicKeyBytes
}
