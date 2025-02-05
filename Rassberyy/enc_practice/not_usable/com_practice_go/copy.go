package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// 플랜트 모델링
var A = [][]float64{
	{1.0000 ,   0.4740  ,  0.6003  ,  0.0802},
	{0 ,   0.8774  ,  3.7668  ,  0.6003},
	{0  , -0.1021  ,  8.1373 ,   1.4502},
	{0  ,-0.6406 ,  44.9464  ,  8.1373},
}

var B = [][]float64{
	{0.2603},
    {1.2262},
    {1.0209},
    {6.4061},
}

// 일단은 SISO로 밑에 돌림
var C = [][]float64{
	{1, 0, 0, 0},
	{0, 0, 1, 0},
}

// 플랜트 초기 상태
var xp0 = []float64{
	0.000,
	0.000,
	0.01, // 0.0524 // 3 degree
	0.000,
}

func main() {
	// 서버 소켓 설정
	listen, err := net.Listen("tcp", "192.168.0.50:8080") // 라즈베리파이 IP와 포트
	if err != nil {
		fmt.Println("서버 소켓 설정 실패:", err)
		os.Exit(1)
	}
	defer listen.Close()
	fmt.Println("플랜트 서버 실행 중...")

	// 클라이언트와 연결 수락
	conn, addr := listen.Accept()
	if err != nil {
		fmt.Println("연결 수락 실패:", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("컨트롤러와 연결됨:", addr)

	// 초기 상태
	x := xp0 // 초기 상태 설정
	temp_x := []float64{0, 0, 0, 0}; //x 연산 저장용 변수 설정정
	y := C[0][0]*x[0] + C[0][1]*x[1] + C[0][2]*x[2] + C[0][3]*x[3] // 초기 출력값 계산

	// 초기 출력값 전송
	_, err = conn.Write([]byte(fmt.Sprintf("%.3f", y))) // 소수점 3자리로 출력
	if err != nil {
		fmt.Println("초기 출력값 전송 실패:", err)
		return
	}
	fmt.Printf("초기 출력값 전송: %.3f\n", y)

	// 입력값 처리 루프
	for {
		// 입력값 수신 (제어기로부터 입력값을 기다림)
		buf := make([]byte, 1024)
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

		// 플랜트 동역학 계산 (행렬 연산)

		// fmt.Printf("계산전 벡터 x: %+v\n", x)
		// fmt.Printf("A 4,3 * x3: %+v\n", A[3][2]*x[2])
		temp_x [0] = x[0]
		temp_x [1] = x[1]
		temp_x [2] = x[2]
		temp_x [3] = x[3]

		temp_x[0] = A[0][0]*x[0] + A[0][1]*x[1] + A[0][2]*x[2] + A[0][3]*x[3] + B[0][0]*u
		temp_x[1] = A[1][0]*x[0] + A[1][1]*x[1] + A[1][2]*x[2] + A[1][3]*x[3] + B[1][0]*u
		temp_x[2] = A[2][0]*x[0] + A[2][1]*x[1] + A[2][2]*x[2] + A[2][3]*x[3] + B[2][0]*u
		temp_x[3] = A[3][0]*x[0] + A[3][1]*x[1] + A[3][2]*x[2] + A[3][3]*x[3] + B[3][0]*u

		x[0] = temp_x[0]
		x[1] = temp_x[1]
		x[2] = temp_x[2]
		x[3] = temp_x[3]

		// 상태 벡터 x 값 출력 (디버깅)
		fmt.Printf("상태 벡터 x: %+v\n", x)
		fmt.Printf("u: %+v\n", u)

		// 출력값 계산 // C를 SISO로 계산
		y = C[0][0]*x[0] + C[0][1]*x[1] + C[0][2]*x[2] + C[0][3]*x[3]

		// 출력값 전송 (제어기에게)
		_, err = conn.Write([]byte(fmt.Sprintf("%.12f", y))) // 소수점 3자리로 출력
		if err != nil {
			fmt.Println("출력값 전송 실패:", err)
			break
		}

		// 출력값 로그
		fmt.Printf("입력값: %.3f, 출력값: %.3f\n", u, y)
	}
}
