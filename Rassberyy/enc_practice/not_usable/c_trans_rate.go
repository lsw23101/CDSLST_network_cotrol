// package main

// import (
// 	"bytes"
// 	"compress/zlib"
// 	"fmt"
// 	"math"
// 	"math/rand"
// 	"net"
// 	"os"

// 	"github.com/tuneinsight/lattigo/v6/core/rlwe"
// 	"github.com/tuneinsight/lattigo/v6/ring"
// 	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
// )

// func main() {
// 	// 서버 소켓 설정
// 	listen, err := net.Listen("tcp", "192.168.0.50:8080")
// 	if err != nil {
// 		fmt.Println("서버 소켓 설정 실패:", err)
// 		os.Exit(1)
// 	}
// 	defer listen.Close()
// 	fmt.Println("플랜트 서버 실행 중...")

// 	// 클라이언트와 연결 수락
// 	conn, err := listen.Accept()
// 	if err != nil {
// 		fmt.Println("연결 수락 실패:", err)
// 		os.Exit(1)
// 	}
// 	defer conn.Close()
// 	fmt.Println("컨트롤러와 연결됨:", conn.RemoteAddr())

// 	// 암호화 준비
// 	slots := 10
// 	logN := 12
// 	primeGen := ring.NewNTTFriendlyPrimesGenerator(18, uint64(math.Pow(2, float64(logN)+1)))
// 	prime, _ := primeGen.NextAlternatingPrime()

// 	m1 := make([]uint64, slots)
// 	for i := 0; i < slots; i++ {
// 		m1[i] = uint64(rand.Intn(1000))
// 	}

// 	params, _ := bgv.NewParametersFromLiteral(bgv.ParametersLiteral{
// 		LogN:             logN,
// 		LogQ:             []int{28, 28},
// 		LogP:             []int{15},
// 		PlaintextModulus: prime,
// 	})

// 	kgen := rlwe.NewKeyGenerator(params)
// 	sk := kgen.GenSecretKeyNew()
// 	ecd := bgv.NewEncoder(params)
// 	enc := rlwe.NewEncryptor(params, sk)

// 	pt1 := bgv.NewPlaintext(params, params.MaxLevel())
// 	ecd.Encode(m1, pt1)
// 	ct1, _ := enc.EncryptNew(pt1)

// 	// ct1 바이너리 직렬화
// 	binCT1, err := ct1.MarshalBinary()
// 	if err != nil {
// 		fmt.Println("바이너리 직렬화 실패:", err)
// 		return
// 	}

// 	// 데이터 압축
// 	var compressedData bytes.Buffer
// 	writer := zlib.NewWriter(&compressedData)
// 	_, err = writer.Write(binCT1)
// 	if err != nil {
// 		fmt.Println("데이터 압축 실패:", err)
// 		return
// 	}
// 	writer.Close()

// 	// 데이터 전송
// 	_, err = conn.Write(compressedData.Bytes())
// 	if err != nil {
// 		fmt.Println("데이터 전송 실패:", err)
// 		return
// 	}
// 	fmt.Println("압축된 ct1 데이터 전송 완료.")
// }
