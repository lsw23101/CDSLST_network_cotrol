package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"math"
	"net"
	"time"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

func main() {
	// 컨트롤러 소켓 설정
	conn, err := net.Dial("tcp", "192.168.0.50:8080")
	if err != nil {
		fmt.Println("서버에 연결 실패:", err)
		return
	}
	defer conn.Close()

	// 암호화 준비
	logN := 12
	primeGen := ring.NewNTTFriendlyPrimesGenerator(18, uint64(math.Pow(2, float64(logN)+1)))
	prime, _ := primeGen.NextAlternatingPrime()

	params, _ := bgv.NewParametersFromLiteral(bgv.ParametersLiteral{
		LogN:             logN,
		LogQ:             []int{28, 28},
		LogP:             []int{15},
		PlaintextModulus: prime,
	})

	ct0 := rlwe.NewCiphertext(params, params.MaxLevel())

	// 데이터 수신
	chunkSize := 1024
	buf := make([]byte, chunkSize)
	var totalData []byte

	startTime := time.Now()
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("수신 오류:", err)
			break
		}
		totalData = append(totalData, buf[:n]...)

		// 데이터 수신 완료 조건 (예시: 전체 크기)
		if len(totalData) >= 131406 {
			break
		}
	}
	endTime := time.Now()

	fmt.Printf("통신에 걸린 시간: %v\n", endTime.Sub(startTime))

	// 데이터 압축 해제
	compressedData := bytes.NewReader(totalData)
	reader, err := zlib.NewReader(compressedData)
	if err != nil {
		fmt.Println("압축 해제 실패:", err)
		return
	}
	defer reader.Close()

	var decompressedData bytes.Buffer
	_, err = decompressedData.ReadFrom(reader)
	if err != nil {
		fmt.Println("데이터 읽기 실패:", err)
		return
	}

	// 바이너리 데이터를 Ciphertext 객체로 복원
	err = ct0.UnmarshalBinary(decompressedData.Bytes())
	if err != nil {
		fmt.Println("Ciphertext 역직렬화 실패:", err)
		return
	}
	fmt.Printf("복원된 ct1의 사이즈: %v\n", ct0.BinarySize())
}
