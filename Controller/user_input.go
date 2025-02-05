package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// ============== Pre-designed controller ==============
var F = [][]float64{
	{0.0410, 0.5213, -2.6980, -0.2872},
	{-0.5376, 1.1004, -13.2632, -1.1307},
	{-0.3812, 0.0836, -15.9162, 0.0089},
	{-2.0722, 0.5246, -94.5827, -0.9060},
}

var G = [][]float64{
	{0.9676, 1.2625},
	{0.5782, 7.4383},
	{0.4150, 16.0674},
	{2.2844, 89.4188},
}

var H = [][]float64{
	{0.0331, 0.1819, -7.8223, -1.4117},
}

var xc0 = []float64{
	0,
	0,
	0,
	0,
}

var u0 = 0.000
var userInput string

func main() {
	// 컨트롤러 소켓 설정
	conn, err := net.Dial("tcp", "192.168.0.50:8080") // 라즈베리파이의 IP 주소와 포트
	if err != nil {
		fmt.Println("서버에 연결 실패:", err)
		return
	}
	defer conn.Close()

	// 컨트롤러에서 상태 벡터 xc의 초기값
	xc := xc0
	temp_x := make([]float64, 4) // x 연산 저장용 변수 설정
	uComputed := u0

	// 입력값 처리 루프
	for {
		// 출력값 수신 (서버에서 y값 받기)
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("출력값 수신 실패:", err)
			break
		}
		yData := string(buf[:n])

		// 출력값 y를 배열로 변환
		yStrings := strings.Split(strings.TrimSpace(yData), ",")
		if len(yStrings) != 2 {
			fmt.Println("출력값 배열 크기 불일치:", yData)
			continue
		}

		// y 배열 초기화 및 변환
		y := make([]float64, 2)
		for i, s := range yStrings {
			y[i], err = strconv.ParseFloat(strings.TrimSpace(s), 64)
			if err != nil {
				fmt.Println("출력값 변환 실패:", err)
				break
			}
		}

		// // 계산된 u 값 전송
		// _, err = conn.Write([]byte(fmt.Sprintf("%.3f", uComputed)))
		// if err != nil {
		// 	fmt.Println("입력값 전송 실패:", err)
		// 	break
		// }
		// 사용자 입력받기
		fmt.Print("보낼 u 값을 입력하 세요 (float 형식): ")

		_, err = fmt.Scanln(&userInput)
		if err != nil {
			fmt.Println("입력 오류:", err)
			continue
		}
		// 사용자 입력을 float64로 변환
		uUser, err := strconv.ParseFloat(strings.TrimSpace(userInput), 64)
		if err != nil {
			fmt.Println("입력값 변환 실패. 유효한 숫자를 입력하세요.")
			continue
		}
		// 서버로 사용자 입력값 전송
		_, err = conn.Write([]byte(fmt.Sprintf("%f", uUser)))
		if err != nil {
			fmt.Println("입력값 전송 실패:", err)
			break
		}
		//
		fmt.Printf("받은 y: [%f]\n", y)

		fmt.Printf("계산전 xc: [%f]\n", xc)

		// 상태 벡터 xc를 업데이트 (플랜트 동역학 계산)

		for i := 0; i < 4; i++ {
			temp_x[i] = F[i][0]*xc[0] + F[i][1]*xc[1] + F[i][2]*xc[2] + F[i][3]*xc[3]
			for j := 0; j < 2; j++ {
				temp_x[i] += G[i][j] * y[j]
			}
		}
		// 출력값 계산 (제어 입력 u 계산)
		uComputed = 0
		for i := 0; i < 4; i++ {
			uComputed += H[0][i] * xc[i]
		}

		// temp_x -> xc 업데이트
		copy(xc, temp_x)

		fmt.Printf("계산후 xc: [%f]\n", xc)

		// 출력값 로그
		fmt.Printf("서버로 보낸 u 값: %.3f\n", uComputed)

		// 0.05초 딜레이 추가 (플랜트의 응답 속도에 맞추기 위해)
		time.Sleep(50 * time.Millisecond)
	}
}
