// 1/6 ToDo: 암호화시켜서 cypertext polynomial 보내기
// 일단 ARX폼으로 바꾸는거 하지말고 그냥 y 두개를 암호화해서 보내기만

package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/CDSL-EncryptedControl/CDSL/utils"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
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
		0,
		0,
		0.05, // 0.0524, // 3 degree
		0,
	}

	yy0 := [][]float64{
		{0, 0},
		{0, 0},
		{0, 0},
		{0, 0},
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
	decryptor := bgv.NewDecryptor(params, sk)
	encoder := bgv.NewEncoder(params) // 다항식으로 바꿀때 쓰는 인코더 //
	eval := bgv.NewEvaluator(params, nil)

	// bredparmas는 머지
	bredparams := ring.GenBRedConstant(params.PlaintextModulus()) 

	// **************************************************//
	// **************************************************//
	// Quantization - packing - encryption 여기는 플랜트니까 결국 이과정을 통해 ctY를 소켓에 담아야됨
	// 밑에 보이는 yy0vec[i] 이녀석이 
	for i := 0; i < 2; i++ {
		ptY[i] = bgv.NewPlaintext(params, params.MaxLevel())
		encoder.Encode(utils.ModVecFloat(utils.RoundVec(utils.ScalVecMult(1/r, yy0vec[i])), params.PlaintextModulus()), ptY[i])
		ctY[i], _ = encryptor.EncryptNew(ptY[i])
	}

	// **************************************************//
	// **************************************************//

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

	// 초기 상태
	x := xp0 // 초기 상태 설정
	temp_x := make([]float64, 4) // x 연산 저장용 변수
	y := make([]float64, 2)      // 출력값 리스트

	// 초기 출력값 계산
	y[0] = C[0][0]*x[0] + C[0][1]*x[1] + C[0][2]*x[2] + C[0][3]*x[3]
	y[1] = C[1][0]*x[0] + C[1][1]*x[1] + C[1][2]*x[2] + C[1][3]*x[3]

	// 초기 출력값 전송
	initialY := fmt.Sprintf("%.15f,%.15f", y[0], y[1])
	_, err = conn.Write([]byte(initialY)) // 리스트 값을 문자열로 전송
	if err != nil {
		fmt.Println("초기 출력값 전송 실패:", err)
		return
	}
	fmt.Printf("초기 출력값 전송: [%s]\n", initialY)

	// 입력값 처리 루프
	for {
		// 입력값 수신
		buf := make([]byte, 200000000)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("입력값 수신 실패:", err)
			break
		}
		uData := string(buf[:n])
		if uData == "" { // 연결 종료 시 루프 탈출
			break
		}

		// 입력값 예외 처리
		u, err := strconv.ParseFloat(strings.TrimSpace(uData), 64)
		if err != nil {
			fmt.Println("잘못된 입력값:", uData)
			continue
		}

		// 플랜트 동역학 계산
		for i := 0; i < 4; i++ {
			temp_x[i] = 0
			for j := 0; j < 4; j++ {
				temp_x[i] += A[i][j] * x[j]
			}
			temp_x[i] += B[i][0] * u
		}

		copy(x, temp_x) // 상태 갱신

		// 출력값 계산
		y[0] = C[0][0]*x[0] + C[0][1]*x[1] + C[0][2]*x[2] + C[0][3]*x[3]
		y[1] = C[1][0]*x[0] + C[1][1]*x[1] + C[1][2]*x[2] + C[1][3]*x[3]

		// 출력값 전송
		outputY := fmt.Sprintf("%.15f,%.15f", y[0], y[1])
		_, err = conn.Write([]byte(outputY)) // 리스트 값을 문자열로 전송
		if err != nil {
			fmt.Println("출력값 전송 실패:", err)
			break
		}

		// 출력값 로그
		fmt.Printf("입력값: %.15f, 출력값: [%s]\n", u, outputY)
	}
}
