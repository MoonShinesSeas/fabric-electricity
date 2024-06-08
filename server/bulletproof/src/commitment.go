package src

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// PedersenCommit performs a Pedersen commitment on a given value.
func PedersenCommit(value *big.Int) (ECPoint, *big.Int) {
	// Generate a random value for the blinding factor.
	r, err := rand.Int(rand.Reader, EC.N)
	if err != nil {
		panic(err) // In a real-world scenario, you should handle the error gracefully.
	}

	// Modulo operation to ensure the value is within the curve's order.
	modValue := new(big.Int).Mod(value, EC.N)

	// Compute the Pedersen commitment.
	// This is done by adding the result of scalar multiplication of the base point G with the value
	// to the result of scalar multiplication of the base point H with the blinding factor r.
	x1, y1 := EC.C.ScalarBaseMult(modValue.Bytes())
	x2, y2 := EC.C.ScalarBaseMult(r.Bytes())
	commitment := EC.Zero()
	commitment = commitment.Add(ECPoint{x1, y1}).Add(ECPoint{x2, y2})

	return commitment, r
}
/*
Vector Pedersen Commitment

Given an array of values, we commit the array with different generators
for each element and for each randomness.
*/
func VectorPCommit(value []*big.Int) (ECPoint, []*big.Int) {
	R := make([]*big.Int, EC.V)

	commitment := EC.Zero()

	for i := 0; i < EC.V; i++ {
		r, err := rand.Int(rand.Reader, EC.N)
		check(err)

		R[i] = r

		modValue := new(big.Int).Mod(value[i], EC.N)

		// mG, rH
		lhsX, lhsY := EC.C.ScalarMult(EC.BPG[i].X, EC.BPG[i].Y, modValue.Bytes())
		rhsX, rhsY := EC.C.ScalarMult(EC.BPH[i].X, EC.BPH[i].Y, r.Bytes())

		commitment = commitment.Add(ECPoint{lhsX, lhsY}).Add(ECPoint{rhsX, rhsY})
	}

	return commitment, R
}

/*
Two Vector P Commit

Given an array of values, we commit the array with different generators
for each element and for each randomness.
*/
func TwoVectorPCommit(a []*big.Int, b []*big.Int) ECPoint {
	if len(a) != len(b) {
		fmt.Println("TwoVectorPCommit: Uh oh! Arrays not of the same length")
		fmt.Printf("len(a): %d\n", len(a))
		fmt.Printf("len(b): %d\n", len(b))
	}

	commitment := EC.Zero()

	for i := 0; i < EC.V; i++ {
		commitment = commitment.Add(EC.BPG[i].Mult(a[i])).Add(EC.BPH[i].Mult(b[i]))
	}

	return commitment
}

/*
Vector Pedersen Commitment with Gens

Given an array of values, we commit the array with different generators
for each element and for each randomness.

We also pass in the Generators we want to use
*/
func TwoVectorPCommitWithGens(G, H []ECPoint, a, b []*big.Int) ECPoint {
	if len(G) != len(H) || len(G) != len(a) || len(a) != len(b) {
		fmt.Println("TwoVectorPCommitWithGens: Uh oh! Arrays not of the same length")
		fmt.Printf("len(G): %d\n", len(G))
		fmt.Printf("len(H): %d\n", len(H))
		fmt.Printf("len(a): %d\n", len(a))
		fmt.Printf("len(b): %d\n", len(b))
	}

	commitment := EC.Zero()

	for i := 0; i < len(G); i++ {
		modA := new(big.Int).Mod(a[i], EC.N)
		modB := new(big.Int).Mod(b[i], EC.N)

		commitment = commitment.Add(G[i].Mult(modA)).Add(H[i].Mult(modB))
	}

	return commitment
}
