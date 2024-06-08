package utils

import (
	"crypto/elliptic"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"strings"

	"github.com/ZZMarquis/gm/sm2"
	"github.com/ZZMarquis/gm/sm3"
)

type ParticipantRandInt func(rand io.Reader, pub *sm2.PublicKey, msg []byte) (*big.Int, error)

func SimpleParticipantRandInt(rand io.Reader, pub *sm2.PublicKey, msg []byte) (*big.Int, error) {
	return randFieldElement(pub.Curve, rand)
}

// randFieldElement returns a random element of the order of the given
// curve using the procedure given in FIPS 186-4, Appendix B.5.2.
func randFieldElement(c elliptic.Curve, rand io.Reader) (k *big.Int, err error) {
	// See randomPoint for notes on the algorithm. This has to match, or s390x
	// signatures will come out different from other architectures, which will
	// break TLS recorded tests.
	for {
		N := c.Params().N
		b := make([]byte, (N.BitLen()+7)/8)
		if _, err = io.ReadFull(rand, b); err != nil {
			return
		}
		if excess := len(b)*8 - N.BitLen(); excess > 0 {
			b[0] >>= excess
		}
		k = new(big.Int).SetBytes(b)
		if k.Sign() != 0 && k.Cmp(N) < 0 {
			return
		}
	}
}

// 完全采用了sm2签名随机数r的生成方式，只是这里我们使用的默认uid
func SM2ParticipantRandInt(rand io.Reader, pub sm2.PublicKey, msg []byte) (*big.Int, error) {
	m, err := calculateSM2Hash(pub, msg, nil)
	if err != nil {
		return nil, err
	}
	e := hashToInt(m, pub.Curve)

	for {
		k, err := randFieldElement(pub.Curve, rand)
		if err != nil {
			return nil, err
		}

		r, _ := pub.Curve.ScalarBaseMult(k.Bytes()) // (x, y) = k*G
		r.Add(r, e)                                 // r = x + e
		r.Mod(r, pub.Curve.Params().N)              // r = (x + e) mod N
		if r.Sign() != 0 {
			s := new(big.Int).Add(r, k)
			if s.Cmp(pub.Curve.Params().N) != 0 { // if r != 0 && (r + k) != N then ok
				return s, nil
			}
		}
	}
}

var defaultUID = []byte{0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38}

func calculateSM2Hash(pub sm2.PublicKey, data, uid []byte) ([]byte, error) {
	if len(uid) == 0 {
		uid = defaultUID
	}
	za, err := CalculateZA(&pub, uid)
	if err != nil {
		panic(fmt.Sprintf("calculateSM2Hash error:%s\n", err))
	}
	md := sm3.New()
	md.Write(za)
	md.Write(data)
	return md.Sum(nil), nil
}

// CalculateZA ZA = H256(ENTLA || IDA || a || b || xG || yG || xA || yA).
// Compliance with GB/T 32918.2-2016 5.5.
//
// This function will not use default UID even the uid argument is empty.
func CalculateZA(pub *sm2.PublicKey, uid []byte) ([]byte, error) {
	uidLen := len(uid)
	if uidLen >= 0x2000 {
		return nil, errors.New("sm2: the uid is too long")
	}
	entla := uint16(uidLen) << 3
	md := sm3.New()
	md.Write([]byte{byte(entla >> 8), byte(entla)})
	if uidLen > 0 {
		md.Write(uid)
	}
	a := new(big.Int).Sub(pub.Curve.Params().P, big.NewInt(3))
	md.Write(toBytes(pub.Curve, a))
	md.Write(toBytes(pub.Curve, pub.Curve.Params().B))
	md.Write(toBytes(pub.Curve, pub.Curve.Params().Gx))
	md.Write(toBytes(pub.Curve, pub.Curve.Params().Gy))
	md.Write(toBytes(pub.Curve, pub.X))
	md.Write(toBytes(pub.Curve, pub.Y))
	return md.Sum(nil), nil
}

func toBytes(curve elliptic.Curve, value *big.Int) []byte {
	// byteLen := (curve.Params().BitSize + 7) >> 3
	// result := make([]byte, byteLen)
	// value.FillBytes(result)
	// return result
	// 确定字节长度。由于bitSize可能不是8的倍数，所以需要加7然后右移3位来向上取整
	byteLen := (curve.Params().BitSize + 7) >> 3
	// 使用Bytes方法获取value的字节表示
	bytes := value.Bytes()
	// 如果得到的字节比预期的要短，需要在前面填充0
	if len(bytes) < byteLen {
		result := make([]byte, byteLen)
		copy(result[byteLen-len(bytes):], bytes)
		return result
	}
	// 如果得到的字节已经足够长或者更长，直接返回
	return bytes
}

// hashToInt converts a hash value to an integer. Per FIPS 186-4, Section 6.4,
// we use the left-most bits of the hash to match the bit-length of the order of
// the curve. This also performs Step 5 of SEC 1, Version 2.0, Section 4.1.3.
func hashToInt(hash []byte, c elliptic.Curve) *big.Int {
	orderBits := c.Params().N.BitLen()
	orderBytes := (orderBits + 7) / 8
	if len(hash) > orderBytes {
		hash = hash[:orderBytes]
	}

	ret := new(big.Int).SetBytes(hash)
	excess := len(hash)*8 - orderBits
	if excess > 0 {
		ret.Rsh(ret, uint(excess))
	}
	return ret
}

// A invertible implements fast inverse in GF(N).
// type invertible interface {
// 	// Inverse returns the inverse of k mod Params().N.
// 	Inverse(k *big.Int) *big.Int
// }

func getPai(priv *sm2.PrivateKey, pubs []*sm2.PublicKey) (int, error) {
	n := len(pubs)
	if n < 2 {
		return -1, errors.New("require multiple SM2 public keys")
	}
	var pai int = -1
	for i := 0; i < len(pubs); i++ {
		if pubs[i].Curve.A.Cmp(priv.Curve.A) != 0 {
			return -1, errors.New("contains non SM2 public key")
		}
		pub := sm2.CalculatePubKey(priv)
		// if .(pubs[i]) {
		// 	pai = i
		// 	break
		// }
		if pub.X.Cmp(pubs[i].X) == 0 && pub.Y.Cmp(pubs[i].Y) == 0 {
			pai = i
			break
		}
	}
	if pai < 0 {
		return -1, errors.New("does not contain public key of the private key")
	}
	return pai, nil
}

// fermatInverse calculates the inverse of k in GF(P) using Fermat's method
// (exponentiation modulo P - 2, per Euler's theorem). This has better
// constant-time properties than Euclid's method (implemented in
// math/big.Int.ModInverse and FIPS 186-4, Appendix C.1) although math/big
// itself isn't strictly constant-time so it's not perfect.
func fermatInverse(k, N *big.Int) *big.Int {
	two := big.NewInt(2)
	nMinus2 := new(big.Int).Sub(N, two)
	return new(big.Int).Exp(k, nMinus2, N)
}

// 这个hash算法没有给出明确定义
func hash(pubs []*sm2.PublicKey, msg []byte, cx, cy *big.Int) *big.Int {
	// var buffer [32]byte
	// h := sm3.New()
	// for _, pub := range pubs {
	// 	pub.X.FillBytes(buffer[:])
	// 	h.Write(buffer[:])
	// 	pub.Y.FillBytes(buffer[:])
	// 	h.Write(buffer[:])
	// }
	// h.Write(msg)
	// cx.FillBytes(buffer[:])
	// h.Write(buffer[:])
	// cy.FillBytes(buffer[:])
	// h.Write(buffer[:])
	// return hashToInt(h.Sum(nil), pubs[0].Curve)
	h := sm3.New()

	for _, pub := range pubs {
		xBytes := pub.X.Bytes()
		padXBytes := padToFixedLength(xBytes, 32) // 假设公钥的X和Y坐标需要填充到32字节
		h.Write(padXBytes)

		yBytes := pub.Y.Bytes()
		padYBytes := padToFixedLength(yBytes, 32) // 假设公钥的X和Y坐标需要填充到32字节
		h.Write(padYBytes)
	}

	h.Write(msg)

	cxBytes := cx.Bytes()
	padCXBytes := padToFixedLength(cxBytes, 32) // 假设cx和cy需要填充到32字节
	h.Write(padCXBytes)

	cyBytes := cy.Bytes()
	padCYBytes := padToFixedLength(cyBytes, 32) // 假设cx和cy需要填充到32字节
	h.Write(padCYBytes)
	return hashToInt(h.Sum(nil), pubs[0].Curve)
}

// padToFixedLength 将字节切片填充到固定长度。如果原始切片比目标长度短，则在前面填充0。
func padToFixedLength(slice []byte, length int) []byte {
	padded := make([]byte, length)
	copy(padded[length-len(slice):], slice)
	return padded
}

// http://www.jcr.cacrnet.org.cn/CN/10.13868/j.cnki.jcr.000472
func Sign(rand io.Reader, participantRandInt ParticipantRandInt, priv *sm2.PrivateKey, pubs []*sm2.PublicKey, msg []byte) ([]*big.Int, error) {
	n := len(pubs)
	pai, err := getPai(priv, pubs)
	if err != nil {
		return nil, err
	}
	// Step 1
	kPai, err := randFieldElement(priv.Curve, rand)
	if err != nil {
		return nil, err
	}
	kPaiGx, kPaiGy := priv.Curve.ScalarBaseMult(kPai.Bytes())
	c := hash(pubs, msg, kPaiGx, kPaiGy)

	results := make([]*big.Int, n+1)
	// Step 2
	// [pai+1, ... n)
	for i := pai + 1; i < n; i++ {
		s, err := participantRandInt(rand, pubs[i], msg)
		if err != nil {
			return nil, err
		}
		results[i+1] = s
		sx, sy := priv.Curve.ScalarBaseMult(s.Bytes())
		c.Add(s, c)
		c.Mod(c, priv.Curve.Params().N)
		cx, cy := priv.Curve.ScalarMult(pubs[i].X, pubs[i].Y, c.Bytes())
		cx, cy = priv.Curve.Add(sx, sy, cx, cy)
		c = hash(pubs, msg, cx, cy)
	}
	results[0] = new(big.Int).Set(c)
	// [0...pai)
	for i := 0; i < pai; i++ {
		s, err := participantRandInt(rand, pubs[i], msg)
		if err != nil {
			return nil, err
		}
		results[i+1] = s
		sx, sy := priv.Curve.ScalarBaseMult(s.Bytes())
		c.Add(s, c)
		c.Mod(c, priv.Curve.Params().N)
		cx, cy := priv.Curve.ScalarMult(pubs[i].X, pubs[i].Y, c.Bytes())
		cx, cy = priv.Curve.Add(sx, sy, cx, cy)
		c = hash(pubs, msg, cx, cy)
	}

	// Step 3: this step is same with SM2 signature scheme
	c.Mul(c, priv.D)
	kPai.Sub(kPai, c)
	dp1 := new(big.Int).Add(priv.D, one)

	dp1Inv := fermatInverse(dp1, priv.Curve.Params().N) // N != 0

	kPai.Mul(kPai, dp1Inv)
	kPai.Mod(kPai, priv.Curve.Params().N) // N != 0

	results[pai+1] = kPai

	return results, nil
}
func FlodSingature(signature []*big.Int) string {
	// 将环签名转换为 JSON 字符串
	signatureJSON, err := json.Marshal(signature)
	if err != nil {
		log.Fatal("JSON marshaling failed:", err)
	}
	// 将 JSON 字符串转换为普通字符串
	return string(signatureJSON)
}

func DecodeSignature(sign string) []*big.Int {
	// 在需要时，你可以将字符串解析为 []*big.Int 类型的环签名
	var parsedSignature []*big.Int
	if err := json.NewDecoder(strings.NewReader(sign)).Decode(&parsedSignature); err != nil {
		log.Fatal("JSON unmarshaling failed:", err)
	}
	return parsedSignature
}

func publicKeyToBytes(pub *sm2.PublicKey) ([]byte, error) {
	// sm2.PublicKey 的 X 和 Y 坐标都是 *big.Int 类型
	// 首先，我们需要将它们转换为字节切片
	xBytes := pub.X.Bytes()
	yBytes := pub.Y.Bytes()

	// 为了确保字节序一致，我们需要按照网络字节序（大端序）来处理
	xBytes = append(make([]byte, (sm2.BitSize-(len(xBytes)%sm2.BitSize))/2), xBytes...)
	yBytes = append(make([]byte, (sm2.BitSize-(len(yBytes)%sm2.BitSize))/2), yBytes...)

	// 现在我们可以创建一个包含X和Y坐标的字节切片
	pubBytes := make([]byte, 0, 2*sm2.BitSize)
	pubBytes = append(pubBytes, xBytes...)
	pubBytes = append(pubBytes, yBytes...)

	return pubBytes, nil
}

func PrivateKeyToBytes(pri *sm2.PrivateKey) ([]byte, error) {
	// sm2.PublicKey 的 X 和 Y 坐标都是 *big.Int 类型
	// 首先，我们需要将它们转换为字节切片
	// xBytes := pub.X.Bytes()
	// yBytes := pub.Y.Bytes()
	dBytes := pri.D.Bytes()
	// 为了确保字节序一致，我们需要按照网络字节序（大端序）来处理
	// xBytes = append(make([]byte, (sm2.BitSize-(len(xBytes)%sm2.BitSize))/2), xBytes...)
	// yBytes = append(make([]byte, (sm2.BitSize-(len(yBytes)%sm2.BitSize))/2), yBytes...)
	dBytes = append(make([]byte, (sm2.BitSize-(len(dBytes)%sm2.BitSize))/2), dBytes...)
	// 现在我们可以创建一个包含X和Y坐标的字节切片
	priBytes := make([]byte, 0, sm2.BitSize)
	// pubBytes = append(pubBytes, xBytes...)
	// pubBytes = append(pubBytes, yBytes...)
	priBytes = append(priBytes, dBytes...)
	return priBytes, nil
}

func PublicKeysToBytes(pubKeys []*sm2.PublicKey) ([][]byte, error) {
	var serializedPubKeys [][]byte
	for _, pub := range pubKeys {
		pubBytes, err := publicKeyToBytes(pub)
		if err != nil {
			return nil, err
		}
		serializedPubKeys = append(serializedPubKeys, pubBytes)
	}
	return serializedPubKeys, nil
}

func KeyToString(privateKey *sm2.PrivateKey) (string, error) {
	privateKeyBytes, err := json.Marshal(privateKey)
	if err != nil {
		return "", err
	}
	return string(privateKeyBytes), nil
}
func Verify(pubs []*sm2.PublicKey, msg []byte, signature []*big.Int) bool {
	if len(pubs)+1 != len(signature) {
		return false
	}
	c := new(big.Int).Set(signature[0])
	for i := 0; i < len(pubs); i++ {
		pub := pubs[i]
		s := signature[i+1]
		sx, sy := pub.Curve.ScalarBaseMult(s.Bytes())
		c.Add(s, c)
		c.Mod(c, pub.Curve.Params().N)
		cx, cy := pub.Curve.ScalarMult(pubs[i].X, pubs[i].Y, c.Bytes())
		cx, cy = pub.Curve.Add(sx, sy, cx, cy)
		c = hash(pubs, msg, cx, cy)
	}
	return c.Cmp(signature[0]) == 0
}
