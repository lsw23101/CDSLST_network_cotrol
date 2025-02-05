package main

import (
	"fmt"
	// "time"
	"net"
	// time 패키지 추가
	utils "github.com/CDSL-EncryptedControl/CDSL/utils"
	RGSW "github.com/CDSL-EncryptedControl/CDSL/utils/core/RGSW"
	RLWE "github.com/CDSL-EncryptedControl/CDSL/utils/core/RLWE"
	"github.com/tuneinsight/lattigo/v6/core/rgsw"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"fmt"
	"math"
	"net"
	"time" // time 패키지 추가
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

func main() {
	// 컨트롤러 소켓 설정
	conn, err := net.Dial("tcp", "192.168.0.50:8080") // 라즈베리파이의 IP 주소와 포트
	if err != nil {
		fmt.Println("서버에 연결 실패:", err)
		return
	}
	defer conn.Close()

	// 데이터 수신 버퍼 설정
	chunkSize := 1024

	buf := make([]byte, chunkSize) // 1024 바이트씩 수신
	// buf := make([]byte, 65000)

	// 데이터 수신을 위한 누적된 결과 저장
	var totalData []byte

	for {
		// 데이터 수신 (서버에서 전송한 바이너리 데이터 받기)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("수신 오류:", err)
			break
		}

		// 수신된 데이터 누적
		totalData = append(totalData, buf[:n]...)

		// 만약 전체 데이터를 다 받았으면 종료
		if len(totalData) >= 131406 { // 예시로 131406 크기만큼 받으면 종료
			break
		}
	}

	enc_F = [][]*rgsw.Ciphertext{[[0xc000088500 0xc000088540 0xc000088580 0xc0000885c0]
		 [0xc000088600 0xc000088000 0xc000088100 0xc000088140]
		  [0xc000088180 0xc0000881c0 0xc000088200 0xc000088240]
		   [0xc000088640 0xc000088680 0xc0000886c0 0xc000088700]]
		
	}
	ctF := totalData
	fmt.Println("recieved ctF", ctF)
}
