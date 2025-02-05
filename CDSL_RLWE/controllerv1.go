package main

import (
	"fmt"
	"math"

	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

func main() {
	// *****************************************************************
	// ************************* User's choice *************************
	// *****************************************************************
	// ============== Encryption parameters ==============
	// Refer to ``Homomorphic encryption standard''

	// log2 of polynomial degree
	logN := 8
	// Choose the size of plaintext modulus (2^ptSize)
	ptSize := uint64(28)
	// Choose the size of ciphertext modulus (2^ctSize)
	ctSize := int(74)

	// ============== Encryption settings ==============
	// Search a proper prime to set plaintext modulus
	primeGen := ring.NewNTTFriendlyPrimesGenerator(ptSize, uint64(math.Pow(2, float64(logN)+1)))
	ptModulus, _ := primeGen.NextAlternatingPrime()
	fmt.Println("Plaintext modulus:", ptModulus)

	// Create a chain of ciphertext modulus
	logQ := []int{int(math.Floor(float64(ctSize) * 0.5)), int(math.Ceil(float64(ctSize) * 0.5))}

	// Parameters satisfying 128-bit security
	// BGV scheme is used
	params, _ := bgv.NewParametersFromLiteral(bgv.ParametersLiteral{
		LogN:             logN,
		LogQ:             logQ,
		PlaintextModulus: ptModulus,
	})
	fmt.Println("Ciphertext modulus:", params.QBigInt())
	fmt.Println("Degree of polynomials:", params.N())

	eval := bgv.NewEvaluator(params, nil)

	// ==============  Encryption of controller ==============

	// 저장된 파일로 받아오기 / ctY, ctU 초기값  ctHu, ctHy
	// 한번 받고 말 부분

	// **** Encrypted controller ****

	// 이 부분이 컨트롤러 연산

	Uout, _ := eval.MulNew(ctHy[0], ctY[0])
	eval.MulThenAdd(ctHu[0], ctU[0], Uout)
	fmt.Println("ctHu", ctHu)
	for j := 1; j < n; j++ {
		eval.MulThenAdd(ctHy[j], ctY[j], Uout)
		eval.MulThenAdd(ctHu[j], ctU[j], Uout)
	}

	// 연산 된 Uout 보내기

	// 방금 보낸거 재암호화 된거 받기

	// 방금 보낸거로 돌아간 플랜트 출력 받기

	// 위에 받은 두개 밀어내기

	// 컨트롤러 state 업데이트
	ctY = append(ctY[1:], Ycin)
	ctU = append(ctU[1:], Ucin)

}
