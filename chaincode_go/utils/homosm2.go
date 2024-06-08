package utils

import (
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/ZZMarquis/gm/sm2"
	"github.com/ZZMarquis/gm/util"
)

var (
	one = new(big.Int).SetInt64(1)
)

const (
	BitSize    = 256
	KeyBytes   = (BitSize + 7) / 8
	UnCompress = 0x04
)

func nextK(rnd io.Reader, max *big.Int) (*big.Int, error) {
	intOne := new(big.Int).SetInt64(1)
	var k *big.Int
	var err error
	for {
		k, err = rand.Int(rnd, max)
		if err != nil {
			return nil, err
		}
		if k.Cmp(intOne) >= 0 {
			return k, err
		}
	}
}

// inverse 计算椭圆曲线上点的逆元
func Inverse(curve elliptic.Curve, px *big.Int, py *big.Int) (*big.Int, *big.Int, error) {
	if util.IsEcPointInfinity(px, py) {
		return nil, nil, errors.New("point at infinity")
	}
	// 假设我们使用的是标准的Weierstrass曲线，逆元可以通过取y坐标的相反数得到
	negY := new(big.Int).Neg(py)
	negY.Mod(negY, curve.Params().P) // 确保y坐标在模P的范围内
	// 确保新点确实在曲线上
	if !curve.IsOnCurve(px, negY) {
		return nil, nil, fmt.Errorf("computed inverse point is not on the curve")
	}
	return px, negY, nil
}

// (1) 计算椭圆曲线点 S = [h]pk, 其中 h = 1, 若 S 是无穷远点, 则报错并退出;
// (2) 选择随机数 k ∈ Z∗q−1, 计算 c1 = [k]G = (x1, y1), [k]pk = (x2, y2);
// (3) 计算 c2 = [m]G + [k]pk;
// (4) c3 = H(x2||m||y2);
// (5) 输出密文 c = (c1, c2, c3).
func HomoEncrypt(pub *sm2.PublicKey, in []byte) ([]byte, error) {
	var c1 []byte
	var kPBx, kPBy *big.Int

	k, err := nextK(rand.Reader, pub.Curve.N)
	if err != nil {
		return nil, err
	}
	kBytes := k.Bytes()
	//c1=k[G]
	c1x, c1y := pub.Curve.ScalarBaseMult(kBytes)
	c1 = elliptic.Marshal(pub.Curve, c1x, c1y)
	kPBx, kPBy = pub.Curve.ScalarMult(pub.X, pub.Y, kBytes)
	//  c2 = [m]G + [k]pk;
	lc2x, lc2y := pub.Curve.ScalarBaseMult(in)
	c2x, c2y := pub.Curve.Add(lc2x, lc2y, kPBx, kPBy)
	c2 := elliptic.Marshal(pub.Curve, c2x, c2y)

	c1Len := len(c1)
	c2Len := len(c2)

	result := make([]byte, c1Len+c2Len)
	copy(result[:c1Len], c1)
	copy(result[c1Len:c1Len+c2Len], c2)
	return result, nil
}

// [sk] c1 = (x2, y2), [m] G = c2 − [sk] c1;
//
//	[m]G 中恢复 m;
//
// 同态运算后的密文, 直接输出明文 m, 解密完成并退出;
func HomoDecrypt(priv *sm2.PrivateKey, cipherText []byte) ([]byte, error) {
	c1Len := ((priv.Curve.BitSize+7)>>3)*2 + 1
	c1 := make([]byte, c1Len)
	copy(c1, cipherText[:c1Len])
	c1x, c1y := elliptic.Unmarshal(priv.Curve, c1)
	if !priv.Curve.IsOnCurve(c1x, c1y) {
		return nil, errors.New("c1 does not satisfy the elliptic curve equation")
	}
	// S=[h]c1
	sx, sy := priv.Curve.ScalarMult(c1x, c1y, one.Bytes())
	if util.IsEcPointInfinity(sx, sy) {
		return nil, errors.New("[h]C1 at infinity")
	}
	c2Len := len(cipherText) - c1Len
	c2 := make([]byte, c2Len)
	copy(c2, cipherText[c1Len:c1Len+c2Len])

	// [sk] c1 = (x2, y2), [m] G = c2 − [sk] c1;
	c2x, c2y := elliptic.Unmarshal(priv.Curve, c2)
	x2, y2 := priv.Curve.ScalarMult(c1x, c1y, priv.D.Bytes())
	mGx, mGy, err := Inverse(priv.Curve, x2, y2)
	if err != nil {
		return nil, err
	}
	mGx, mGy = priv.Curve.Add(c2x, c2y, mGx, mGy)
	if !priv.Curve.IsOnCurve(mGx, mGy) {
		return []byte("error"), nil
	}
	res, flag := babyStepGiantStep(priv.Curve, priv.Curve.Gx, priv.Curve.Gy, mGx, mGy, priv.Curve.N)
	if flag && res != nil {
		return res.Bytes(), nil
	} else {
		return nil, nil
	}
}

// babyStepGiantStep 解决椭圆曲线上的离散对数问题 mG = (x, y)
// G 是基点, N 是椭圆曲线的阶, (x, y) 是椭圆曲线上的一个点
func babyStepGiantStep(curve elliptic.Curve, gx *big.Int, gy *big.Int, x *big.Int, y *big.Int, N *big.Int) (*big.Int, bool) {
	if !curve.IsOnCurve(x, y) {
		return big.NewInt(-1), false
	}
	m := new(big.Int).Sqrt(curve.Params().N)
	t := new(big.Int).Sub(new(big.Int).Div(N, m), one)
	// Compute the baby steps and store them in the 'precomputed' hash table.
	table := make(map[string]*big.Int) // 哈希表，存储小步的结果
	// 小步
	for i := one; i.Cmp(m) < 0; i = new(big.Int).Add(i, one) {
		iGx, iGy := curve.ScalarBaseMult(i.Bytes())
		if iGx.Cmp(x) == 0 && iGy.Cmp(y) == 0 {
			return i, true
		}
		table[string(elliptic.Marshal(curve, iGx, iGy))] = i
	}
	// 大步
	for j := one; j.Cmp(t) < 0; j = new(big.Int).Add(j, one) {
		jm := new(big.Int).Mul(j, m)
		jGmx, jGmy := curve.ScalarMult(gx, gy, jm.Bytes())
		jGmx_inv, jGmy_inv, err := Inverse(curve, jGmx, jGmy)
		if err != nil {
			return big.NewInt(-1), false
		}
		qjmGx, qjmGy := curve.Add(jGmx_inv, jGmy_inv, x, y)
		for key, k := range table {
			iGx, iGy := elliptic.Unmarshal(curve, []byte(key))
			if iGx.Cmp(qjmGx) == 0 && iGy.Cmp(qjmGy) == 0 {
				res := new(big.Int).Add(jm, k)
				fmt.Println("res,jm,k", res, jm, k)
				return res, true
			}
		}
	}
	return big.NewInt(-1), false // 如果没有找到解，则返回nil和false
}

func CiperAdd(curve elliptic.Curve, cipertext1 []byte, cipertext2 []byte) ([]byte, error) {
	cipertext1c1Len := ((curve.Params().BitSize+7)>>3)*2 + 1
	cipertext1c1 := make([]byte, cipertext1c1Len)
	copy(cipertext1c1, cipertext1[:cipertext1c1Len])
	cipertext1c1x, cipertext1c1y := elliptic.Unmarshal(curve, cipertext1c1)
	if !curve.IsOnCurve(cipertext1c1x, cipertext1c1y) {
		return nil, errors.New("c1 does not satisfy the elliptic curve equation")
	}

	cipertext2c1Len := ((curve.Params().BitSize+7)>>3)*2 + 1
	cipertext2c1 := make([]byte, cipertext2c1Len)
	copy(cipertext2c1, cipertext2[:cipertext2c1Len])
	cipertext2c1x, cipertext2c1y := elliptic.Unmarshal(curve, cipertext2c1)
	if !curve.IsOnCurve(cipertext2c1x, cipertext2c1y) {
		return nil, errors.New("c2 does not satisfy the elliptic curve equation")
	}

	cipertext1c2Len := len(cipertext1) - cipertext1c1Len
	cipertext1c2 := make([]byte, cipertext1c2Len)
	copy(cipertext1c2, cipertext1[cipertext1c2Len:cipertext1c1Len+cipertext1c2Len])
	cipertext1c2x, cipertext1c2y := elliptic.Unmarshal(curve, cipertext1c2)

	cipertext2c2Len := len(cipertext2) - cipertext2c1Len
	cipertext2c2 := make([]byte, cipertext2c2Len)
	copy(cipertext2c2, cipertext2[cipertext2c2Len:cipertext2c1Len+cipertext2c2Len])
	cipertext2c2x, cipertext2c2y := elliptic.Unmarshal(curve, cipertext2c2)

	cipertextc1x, cipertextc1y := curve.Add(cipertext1c1x, cipertext1c1y, cipertext2c1x, cipertext2c1y)
	cipertextc2x, cipertextc2y := curve.Add(cipertext1c2x, cipertext1c2y, cipertext2c2x, cipertext2c2y)

	c1 := elliptic.Marshal(curve, cipertextc1x, cipertextc1y)
	c2 := elliptic.Marshal(curve, cipertextc2x, cipertextc2y)

	c1Len := len(c1)
	c2Len := len(c2)

	result := make([]byte, c1Len+c2Len)
	copy(result[:c1Len], c1)
	copy(result[c1Len:c1Len+c2Len], c2)
	return result, nil
}
func CiperSub(curve elliptic.Curve, cipertext1 []byte, cipertext2 []byte) ([]byte, error) {
	cipertext1c1Len := ((curve.Params().BitSize+7)>>3)*2 + 1
	cipertext1c1 := make([]byte, cipertext1c1Len)
	copy(cipertext1c1, cipertext1[:cipertext1c1Len])
	cipertext1c1x, cipertext1c1y := elliptic.Unmarshal(curve, cipertext1c1)
	if !curve.IsOnCurve(cipertext1c1x, cipertext1c1y) {
		return nil, errors.New("c1 does not satisfy the elliptic curve equation")
	}
	cipertext2c1Len := ((curve.Params().BitSize+7)>>3)*2 + 1
	cipertext2c1 := make([]byte, cipertext2c1Len)
	copy(cipertext2c1, cipertext2[:cipertext2c1Len])
	cipertext2c1x, cipertext2c1y := elliptic.Unmarshal(curve, cipertext2c1)
	if !curve.IsOnCurve(cipertext2c1x, cipertext2c1y) {
		return nil, errors.New("c2 does not satisfy the elliptic curve equation")
	}
	cipertext1c2Len := len(cipertext1) - cipertext1c1Len
	cipertext1c2 := make([]byte, cipertext1c2Len)
	copy(cipertext1c2, cipertext1[cipertext1c2Len:cipertext1c1Len+cipertext1c2Len])
	cipertext1c2x, cipertext1c2y := elliptic.Unmarshal(curve, cipertext1c2)

	cipertext2c2Len := len(cipertext2) - cipertext2c1Len
	cipertext2c2 := make([]byte, cipertext2c2Len)
	copy(cipertext2c2, cipertext2[cipertext2c2Len:cipertext2c1Len+cipertext2c2Len])
	cipertext2c2x, cipertext2c2y := elliptic.Unmarshal(curve, cipertext2c2)
	inv_cipertext2c1x, inv_cipertext2c1y, err := Inverse(curve, cipertext2c1x, cipertext2c1y)
	if err != nil {
		return nil, errors.New("inv_c1 has error")
	}
	inv_cipertext2c2x, inv_cipertext2c2y, err := Inverse(curve, cipertext2c2x, cipertext2c2y)
	if err != nil {
		return nil, errors.New("inv_c2 has error")
	}
	cipertextc1x, cipertextc1y := curve.Add(cipertext1c1x, cipertext1c1y, inv_cipertext2c1x, inv_cipertext2c1y)
	cipertextc2x, cipertextc2y := curve.Add(cipertext1c2x, cipertext1c2y, inv_cipertext2c2x, inv_cipertext2c2y)

	c1 := elliptic.Marshal(curve, cipertextc1x, cipertextc1y)
	c2 := elliptic.Marshal(curve, cipertextc2x, cipertextc2y)

	c1Len := len(c1)
	c2Len := len(c2)

	result := make([]byte, c1Len+c2Len)
	copy(result[:c1Len], c1)
	copy(result[c1Len:c1Len+c2Len], c2)
	return result, nil
}
