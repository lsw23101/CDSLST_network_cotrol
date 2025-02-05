package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

func readSingleColumnCSV(filename string) ([]byte, error) {
	// Open the CSV file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create a CSV reader
	reader := csv.NewReader(file)

	// Read all records from the CSV file
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %v", err)
	}

	// Convert the string records to a byte slice
	var data []byte
	for _, record := range records {
		if len(record) != 1 {
			return nil, fmt.Errorf("unexpected CSV format: each row should have exactly 1 column")
		}
		// Convert the string to an integer, then to a byte
		value, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, fmt.Errorf("failed to convert record to integer: %v", err)
		}
		data = append(data, byte(value))
	}

	return data, nil
}

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

	// // ============== Plant model ==============
	// A := [][]float64{
	// 	{0.998406460921939, 0, 0.00417376927758289, 0},
	// 	{0, 0.998893625478993, 0, -0.00332671872292611},
	// 	{0, 0, 0.995822899329324, 0},
	// 	{0, 0, 0, 0.996671438596397},
	// }
	// B := [][]float64{
	// 	{0.00831836513049678, 9.99686131895421e-06},
	// 	{-5.19664522845810e-06, 0.00627777465144397},
	// 	{0, 0.00477571210746992},
	// 	{0.00311667643652227, 0},
	// }
	// C := [][]float64{
	// 	{0.500000000000000, 0, 0, 0},
	// 	{0, 0.500000000000000, 0, 0},
	// }
	// // Plant initial state
	// xp0 := []float64{
	// 	1,
	// 	1,
	// 	1,
	// 	1,
	// }

	// ============== Quantization parameters ==============
	r := 0.00020
	s := 0.00010
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
	// sk를 고정시킬 예정
	// kgen := bgv.NewKeyGenerator(params)

	// CSV 파일 경로
	filename := "sk_bin_data.csv"

	// 1열짜리 CSV 파일 읽기
	// sk 고정시키기
	totalData, _ := readSingleColumnCSV(filename)
	sk := rlwe.NewSecretKey(params)
	_ = sk.UmarshalBinanry(totalData[:4128])

	fmt.Println(sk)

	// //fmt.Println("sk", sk.BinarySize())
	// encryptor := bgv.NewEncryptor(params, sk)
	// decryptor := bgv.NewDecryptor(params, sk)
	// encoder := bgv.NewEncoder(params)

	// bredparams := ring.GenBRedConstant(params.PlaintextModulus())

	// // ==============  Encryption of controller ==============
	// // dimensions
	// n := len(A)
	// l := len(C)
	// m := len(B[0])
	// h := int(math.Max(float64(l), float64(m)))

	// // Plaintext of past inputs and outputs
	// ptY := make([]*rlwe.Plaintext, n)
	// ptU := make([]*rlwe.Plaintext, n)

	// // Ciphertext of past inputs and outputs
	// ctY := make([]*rlwe.Ciphertext, n)
	// ctU := make([]*rlwe.Ciphertext, n)

	// // Plant state
	// xp := xp0

	// // **** Sensor ****
	// // Plant output
	// Y := utils.MatVecMult(C, xp) // [][]float64

	// // Quantize and duplicate << 여기서
	// //fmt.Println("Ypacked", Ypacked) 여기서 초기값 Ycin
	// Ysens := utils.ModVecFloat(utils.RoundVec(utils.ScalVecMult(1/r, utils.VecDuplicate(Y, m, h))), params.PlaintextModulus())
	// fmt.Println("Ysens", Ysens)
	// Ypacked := bgv.NewPlaintext(params, params.MaxLevel())
	// encoder.Encode(Ysens, Ypacked)
	// Ycin, _ := encryptor.EncryptNew(Ypacked)

	// // 여기서 컨트롤러가 보내는 ctU를 받음

	// // **** Actuator ****
	// // Plant input
	// U := make([]float64, m)
	// // Unpacked and re-scaled u at actuator
	// Uact := make([]uint64, params.N())
	// // u after inner sum
	// Usum := make([]uint64, m)
	// encoder.Decode(decryptor.DecryptNew(Uout), Uact)
	// // 여기서 Uact가 Uout을 복호화하고 언패킹
	// fmt.Println("Uact", Uact)

	// // Generate plant input // 아까 복제한 애들을 다시 잘 뭉쳐서 U만들기
	// for k := 0; k < m; k++ {
	// 	Usum[k] = utils.VecSumUint(Uact[k*h:(k+1)*h], params.PlaintextModulus(), bredparams)
	// 	U[k] = float64(r * s * utils.SignFloat(float64(Usum[k]), params.PlaintextModulus()))
	// }

	// fmt.Println("U", U)

	// // Re-encryption
	// Upacked := bgv.NewPlaintext(params, params.MaxLevel())
	// encoder.Encode(utils.ModVecFloat(utils.RoundVec(utils.ScalVecMult(1/r, utils.VecDuplicate(U, m, h))), params.PlaintextModulus()), Upacked)
	// Ucin, _ := encryptor.EncryptNew(Upacked)
	// //fmt.Println("Ucin", Ucin)

	// // 여기서 받은거 재암호화한 Ucin 보내고
	// // 아까 계산한 최신 Ycin 보내기

	// // **** Plant ****
	// // State update
	// xp = utils.VecAdd(utils.MatVecMult(A, xp), utils.MatVecMult(B, U))

}
