package main

import (
	"fmt"
	"math"
	"net"
	"os"

	"github.com/CDSL-EncryptedControl/CDSL/utils"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

func main() {
	// *****************************************************************
	// *****************************************************************
	// 서버 소켓 설정
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
	// *****************************************************************


	// *****************************************************************
	// ************************* User's choice *************************
	// *****************************************************************
	// ============== Encryption parameters ==============
	// Refer to ``Homomorphic encryption standard''

	// log2 of polynomial degree
	logN := 12
	// Choose the size of plaintext modulus (2^ptSize) // plantext size
	ptSize := uint64(29)
	// Choose the size of ciphertext modulus (2^ctSize)
	ctSize := int(74)

	// ============== Plant model ==============
	A := [][]float64{
		{1, 0.099091315838924, 0.013632351698217, 0.000450408274543},
		{0, 0.981778666086312, 0.278888611257509, -0.013632351698217},
		{0, -0.002318427159561, 1.159794783603433, 0.105273449905749},
		{0, -0.047430035928148, 3.276421050834629, 1.159794783603433},
	}
	B := [][]float64{
		{0.009086841610758},
		{0.182213339136875},
		{0.023184271595607},
		{0.474300359281477},
	}
	C := [][]float64{
		{1, 0, 0, 0},
		{0, 0, 1, 0},
	}
	// Plant initial state
	xp0 := []float64{
		0.05,
		0,
		0.05, // 0.0524, // 3 degree
		0,
	}


	// ARX 폼의 초기값 [-4]~[-1] 까지의 u와y 의 transpose

	// transpose of Yini from conversion.m

	yy0 := [][]float64{
		{0, 0},
		{0, 0},
		{0, 0},
		{0, 0},
	}

	// transpose of Uini from conversion.m
	uu0 := [][]float64{
		{0},
		{0},
		{0},
		{0},
	}

	// ============== Quantization parameters ==============
	r := 0.0001
	s := 0.0001
	fmt.Println("Scaling parameters 1/r:", 1/r, "1/s:", 1/s)
	// *****************************************************************
	// *****************************************************************

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

	// Generate secret key
	kgen := bgv.NewKeyGenerator(params)
	sk := kgen.GenSecretKeyNew()

	encryptor := bgv.NewEncryptor(params, sk) // 다항식을 암호화할떄 쓰는 인코더

	encoder := bgv.NewEncoder(params) // 다항식으로 바꿀때 쓰는 인코더 //



	// ==============  Encryption of controller ==============
	// dimensions
	nx := len(A)
	ny := len(C)
	nu := len(B[0])
	h := int(math.Max(float64(ny), float64(nu)))

	// fmt.Println("A B C의 차원과 y u 중 큰 녀석의 차원:", nx, ny, nu, h)

	// duplicate // Q: 여기서 복제하는게 뭐지
	yy0vec := make([][]float64, nx)
	uu0vec := make([][]float64, nx)
	for i := 0; i < nx; i++ {
		yy0vec[i] = utils.VecDuplicate(yy0[i], nu, h)
		uu0vec[i] = utils.VecDuplicate(uu0[i], nu, h)
	}

	// Plaintext of past inputs and outputs
	ptY := make([]*rlwe.Plaintext, nx)

	// Ciphertext of past inputs and outputs
	ctY := make([]*rlwe.Ciphertext, nx)

	// Quantization - packing - encryption
	for i := 0; i < nx; i++ {
		ptY[i] = bgv.NewPlaintext(params, params.MaxLevel())
		encoder.Encode(utils.ModVecFloat(utils.RoundVec(utils.ScalVecMult(1/r, yy0vec[i])), params.PlaintextModulus()), ptY[i])
		ctY[i], _ = encryptor.EncryptNew(ptY[i])


	}

	// ============== Simulation ==============
	// Number of simulation steps
	iter := 1
	fmt.Printf("Number of iterations: %v\n", iter)

	// Plant 

	// Plant state
	xp := xp0


	for i := 0; i < iter; i++ {
		// **** Sensor ****
		// Plant output
		Y := utils.MatVecMult(C, xp) // [][]float64

		// Quantize and duplicate
		Ysens := utils.ModVecFloat(utils.RoundVec(utils.ScalVecMult(1/r, utils.VecDuplicate(Y, nu, h))), params.PlaintextModulus())
		Ypacked := bgv.NewPlaintext(params, params.MaxLevel())
		encoder.Encode(Ysens, Ypacked)
		Ycin, _ := encryptor.EncryptNew(Ypacked)

		fmt.Printf("Enc Y: %v\n", Ycin)
		fmt.Printf("Type of Ycin: %T\n", Ycin)
		fmt.Printf("Size of Ycin: %d\n", len(Ycin.Value))
	}
}
