package main

import (
	"fmt"
	"net"
	"os"

	utils "github.com/CDSL-EncryptedControl/CDSL/utils"
	RGSW "github.com/CDSL-EncryptedControl/CDSL/utils/core/RGSW"
	RLWE "github.com/CDSL-EncryptedControl/CDSL/utils/core/RLWE"
	"github.com/tuneinsight/lattigo/v6/core/rgsw"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"

)

func main() {

	listen, err := net.Listen("tcp", "192.168.0.50:8080") // 라즈베리파이 IP와 포트
	if err != nil {
		fmt.Println("서버 소켓 설정 실패:", err)
		os.Exit(1)
	}
	defer listen.Close()
	fmt.Println("플랜트 서버 실행 중...")

	// 클라이언트와 연결 수락
	conn, err := listen.Accept()
	if err != nil {
		fmt.Println("연결 수락 실패:", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("컨트롤러와 연결됨:", conn.RemoteAddr())
	// *****************************************************************
	// ************************* User's choice *************************
	// *****************************************************************
	// ============== Encryption parameters ==============
	// Refer to ``Homomorphic encryption standard''
	params, _ := rlwe.NewParametersFromLiteral(rlwe.ParametersLiteral{
		// log2 of polynomial degree
		LogN: 11,
		// Size of ciphertext modulus (Q)
		LogQ: []int{56},
		// Size of plaintext modulus (P)
		LogP:    []int{56},
		NTTFlag: true,
	})
	fmt.Println("Degree of polynomials:", params.N())
	fmt.Println("Ciphertext modulus:", params.QBigInt())
	fmt.Println("Special modulus:", params.PBigInt())
	// Default secret key distribution
	// Each coefficient in the polynomial is uniformly sampled in [-1, 0, 1]
	fmt.Println("Secret key distribution:", params.Xs())
	// Default error distribution
	// Each coefficient in the polynomial is sampled according to a
	// discrete Gaussian distribution with standard deviation 3.2 and bound 19.2
	fmt.Println("Error distribution:", params.Xe())


	// ============== Pre-designed controller ==============
	// F must be an integer matrix
	F := [][]float64{
		{-1, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 2, 0},
		{0, 0, 0, 1},
	}

	// Controller initial state
	x0 := []float64{
		0.5,
		0.02,
		-1,
		0.9,
	}

	// ============== Quantization parameters ==============
	s := 1 / 10000.0
	L := 1 / 1000.0
	r := 1 / 1000.0
	fmt.Printf("Scaling parameters 1/L: %v, 1/s: %v, 1/r: %v \n", 1/L, 1/s, 1/r)
	// *****************************************************************
	// *****************************************************************

	// ============== Encryption settings ==============
	// Set parameters
	levelQ := params.QCount() - 1
	levelP := params.PCount() - 1
	ringQ := params.RingQ()

	// Generate keys
	kgen := rlwe.NewKeyGenerator(params)
	sk := kgen.GenSecretKeyNew()
	rlk := kgen.GenRelinearizationKeyNew(sk)
	// evk := rlwe.NewMemEvaluationKeySet(rlk)

	// Define encryptor and evaluator
	encryptorRLWE := rlwe.NewEncryptor(params, sk)
	//decryptorRLWE := rlwe.NewDecryptor(params, sk)
	encryptorRGSW := rgsw.NewEncryptor(params, sk)
	//evaluator := rgsw.NewEvaluator(params, evk)

	// ==============  Encryption of controller ==============
	// // Quantization
	// GBar := utils.ScalMatMult(1/s, G)
	// RBar := utils.ScalMatMult(1/s, R)
	// HBar := utils.ScalMatMult(1/s, H)

	// Encryption
	ctF := RGSW.Enc(F, encryptorRGSW, levelQ, levelP, params)
	//ctG := RGSW.Enc(GBar, encryptorRGSW, levelQ, levelP, params)
	//ctH := RGSW.Enc(HBar, encryptorRGSW, levelQ, levelP, params)
	//ctR := RGSW.Enc(RBar, encryptorRGSW, levelQ, levelP, params)

	// ============== 클라우드에 보내야 할 암호화된 행렬과 eval key===============
	// rgsw도 바이너리로 보낼 방법이 있을까 ? << ctF는 데이터구조가 2x4 리스트 ?
	fmt.Println("ctF", ctF)
	


	// Controller state encryption
	xBar := utils.ScalVecMult(1/(r*s), x0)
	xCt := RLWE.Enc(xBar, 1/L, *encryptorRLWE, ringQ, params)
	fmt.Println("\nxCt", xCt)


	//============== 암호화된 컨트롤러 행렬, 초기값 보내기=======================

	// _, err = conn.Write(ctF)
	// if err != nil {
	// 	fmt.Println("출력값 전송 실패:", err)
	// 	return
	// }

}
