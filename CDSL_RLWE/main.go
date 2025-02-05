package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/CDSL-EncryptedControl/CDSL/utils"
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

// Function to save the binary data as CSV
func saveBinaryDataToCSV(ctHu_bin [][]byte, filename string) error {
	// Create or overwrite the CSV file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush() // Ensure all data is written to the file

	// Number of rows is the size of a single binary slice ct.BinarySize()
	numRows := len(ctHu_bin[0])
	if numRows == 0 {
		return fmt.Errorf("binary data slice is empty")
	}

	// Write each row (each row corresponds to the data from each index in ctHu_bin)
	for i := 0; i < numRows; i++ {
		var record []string
		for _, binData := range ctHu_bin {
			// Convert each byte to a string and append it to the record
			record = append(record, fmt.Sprintf("%d", binData[i]))
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record to CSV: %v", err)
		}
	}

	return nil
}

func saveBinaryDataToSingleColumnCSV(ctHu_bin []byte, filename string) error {
	// Create or overwrite the CSV file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush() // Ensure all data is written to the file

	// Write each byte as a new row
	for _, binData := range ctHu_bin {
		record := []string{fmt.Sprintf("%d", binData)}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record to CSV: %v", err)
		}
	}

	return nil
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

	// ============== Plant model ==============
	A := [][]float64{
		{0.998406460921939, 0, 0.00417376927758289, 0},
		{0, 0.998893625478993, 0, -0.00332671872292611},
		{0, 0, 0.995822899329324, 0},
		{0, 0, 0, 0.996671438596397},
	}
	B := [][]float64{
		{0.00831836513049678, 9.99686131895421e-06},
		{-5.19664522845810e-06, 0.00627777465144397},
		{0, 0.00477571210746992},
		{0.00311667643652227, 0},
	}
	C := [][]float64{
		{0.500000000000000, 0, 0, 0},
		{0, 0.500000000000000, 0, 0},
	}
	// Plant initial state
	xp0 := []float64{
		1,
		1,
		1,
		1,
	}

	// ============== Pre-designed controller ==============
	F := [][]float64{
		{0.601084882500204, 0.00130548899463723, 0.00188689266655532, -0.00223157438686797},
		{-0.000970175325053589, 0.603135944526756, -0.00214986824072896, -0.00135615804381827},
		{-0.160263310167643, -0.00376022301501287, 0.994186337810539, 0.00149800905597996},
		{-0.00246363925973350, 0.160453719221470, -0.000855550018163947, 0.995834150465395},
	}
	G := [][]float64{
		{0.781489217538651, -1.65204644795806e-17},
		{3.55937121216121e-18, 0.781627935451296},
		{0.319044281150596, -2.56382834592279e-15},
		{6.71867608432863e-18, -0.319923274222079},
	}
	H := [][]float64{
		{-0.790470011857417, 0.157886813229693, -0.274507166717187, -0.268647756048890},
		{-0.155195618091332, -0.787363838187106, -0.342684291254742, 0.313672395293020},
	}

	// input-output representation of controller obtained by conversion.m
	// transpose of vecHu, vecHy from conversion.m
	Hy := [][]float64{
		{0.334883269997112, -0.0993726952581632, 0.109105860257554, 0.340141173304891},
		{0.340715074862138, -0.101693452659005, 0.111263681570879, 0.346096102431116},
		{0.0212757993084255, -0.00721494759029773, 0.00717571762620109, 0.0215259945842975},
		{-0.705323732730193, 0.209355413587286, -0.230615165512593, -0.715776671026420},
	}
	Hu := [][]float64{
		{-0.285602015399616, -0.000307101965816320, 0.00106747945670671, -0.286337872976116},
		{0.183962668144521, -0.000156850543232820, 0.000585408816047406, 0.183342919294642},
		{0.464731844320360, -0.000717550250832144, 0.000183250207538066, 0.464698956437188},
		{0.631884279880355, -0.00124460838502882, -0.000477508261005455, 0.632382252336539},
	}
	// Controller initial state
	xc0 := []float64{
		0.500000000000000,
		0.0200000000000000,
		-1,
		0.900000000000000,
	}
	// transpose of Yini from conversion.m
	yy0 := [][]float64{
		{-168.915339084001, 152.553129120773},
		{0, 0},
		{0, 0},
		{37.1009230518511, -33.8787596718866},
	}
	// transpose of Uini from conversion.m
	uu0 := [][]float64{
		{0, 0},
		{151.077820919228, -70.2395320362580},
		{90.8566491021641, -42.4186053244263},
		{54.6591007720606, -25.4768092703056},
	}

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
	// kgen := bgv.NewKeyGenerator(params)
	// sk := kgen.GenSecretKeyNew()

	// CSV 파일 경로
	filename := "sk_bin_data.csv"

	// 1열짜리 CSV 파일 읽기
	// sk 고정시키기
	totalData, _ := readSingleColumnCSV(filename)
	sk := rlwe.NewSecretKey(params) // 이거는 빈 sk 만드는 함수

	_ = sk.UnmarshalBinary(totalData[:4128])
	//fmt.Println(sk)

	sk_bin, _ := sk.MarshalBinary()
	//fmt.Println("sk_bin", sk_bin)

	encryptor := bgv.NewEncryptor(params, sk)
	decryptor := bgv.NewDecryptor(params, sk)
	encoder := bgv.NewEncoder(params)
	eval := bgv.NewEvaluator(params, nil)
	//eval_ := bgv.NewEvaluator(params, nil)

	//fmt.Println(*eval, "\n", *eval_)

	bredparams := ring.GenBRedConstant(params.PlaintextModulus())

	// ==============  Encryption of controller ==============
	// dimensions
	n := len(A)
	l := len(C)
	m := len(B[0])
	h := int(math.Max(float64(l), float64(m)))

	// duplicate
	yy0vec := make([][]float64, n)
	uu0vec := make([][]float64, n)
	for i := 0; i < n; i++ {
		yy0vec[i] = utils.VecDuplicate(yy0[i], m, h)
		uu0vec[i] = utils.VecDuplicate(uu0[i], m, h)
	}

	fmt.Println("n", n)
	// Plaintext of past inputs and outputs
	ptY := make([]*rlwe.Plaintext, n)
	ptU := make([]*rlwe.Plaintext, n)
	// Plaintext of control parameters
	ptHy := make([]*rlwe.Plaintext, n)
	ptHu := make([]*rlwe.Plaintext, n)
	// Ciphertext of past inputs and outputs
	ctY := make([]*rlwe.Ciphertext, n)
	ctU := make([]*rlwe.Ciphertext, n)
	// Ciphertext of control parameters
	ctHy := make([]*rlwe.Ciphertext, n)
	ctHu := make([]*rlwe.Ciphertext, n)

	fmt.Println("ctY=", ctY)
	// Quantization - packing - encryption
	for i := 0; i < n; i++ {
		ptY[i] = bgv.NewPlaintext(params, params.MaxLevel())
		encoder.Encode(utils.ModVecFloat(utils.RoundVec(utils.ScalVecMult(1/r, yy0vec[i])), params.PlaintextModulus()), ptY[i])
		ctY[i], _ = encryptor.EncryptNew(ptY[i])

		ptU[i] = bgv.NewPlaintext(params, params.MaxLevel())
		encoder.Encode(utils.ModVecFloat(utils.RoundVec(utils.ScalVecMult(1/r, uu0vec[i])), params.PlaintextModulus()), ptU[i])
		ctU[i], _ = encryptor.EncryptNew(ptU[i])

		ptHy[i] = bgv.NewPlaintext(params, params.MaxLevel())
		encoder.Encode(utils.ModVecFloat(utils.RoundVec(utils.ScalVecMult(1/s, Hy[i])), params.PlaintextModulus()), ptHy[i])
		ctHy[i], _ = encryptor.EncryptNew(ptHy[i])

		ptHu[i] = bgv.NewPlaintext(params, params.MaxLevel())
		encoder.Encode(utils.ModVecFloat(utils.RoundVec(utils.ScalVecMult(1/s, Hu[i])), params.PlaintextModulus()), ptHu[i])
		ctHu[i], _ = encryptor.EncryptNew(ptHu[i])
	}

	fmt.Println("ctHu=", ctHu)
	//fmt.Println("value of ctHu[1]=", ctHu[1])
	//fmt.Println("value of *ctHu[1]=", *ctHu[1])

	fmt.Println("ptHu=", ptHu)

	fmt.Println("ctY=", ctY)
	//fmt.Println("ctY[1]=", ctY[1])

	// ============== Simulation ==============
	// Number of simulation steps
	iter := 1
	fmt.Printf("Number of iterations: %v\n", iter)

	// 1) Plant + unencrypted (original) controller
	// Data storage
	yUnenc := [][]float64{}
	uUnenc := [][]float64{}
	xpUnenc := [][]float64{}
	xcUnenc := [][]float64{}

	xpUnenc = append(xpUnenc, xp0)
	xcUnenc = append(xcUnenc, xc0)

	// Plant state
	xp := xp0
	// Controller state
	xc := xc0

	for i := 0; i < iter; i++ {
		y := utils.MatVecMult(C, xp)
		u := utils.MatVecMult(H, xc)
		xp = utils.VecAdd(utils.MatVecMult(A, xp), utils.MatVecMult(B, u))
		xc = utils.VecAdd(utils.MatVecMult(F, xc), utils.MatVecMult(G, y))

		yUnenc = append(yUnenc, y)
		uUnenc = append(uUnenc, u)
		xpUnenc = append(xpUnenc, xp)
		xcUnenc = append(xcUnenc, xc)
	}

	// 2) Plant + encrypted controller

	// To save data
	yEnc := [][]float64{}
	uEnc := [][]float64{}
	xpEnc := [][]float64{}
	xpEnc = append(xpEnc, xp0)

	// Plant state
	xp = xp0

	// **** Sensor ****
	// Plant output
	Y := utils.MatVecMult(C, xp) // [][]float64
	fmt.Println("Y", Y)

	// Quantize and duplicate << 여기서
	//fmt.Println("Ypacked", Ypacked) 여기서 초기값 Yin
	Ysens := utils.ModVecFloat(utils.RoundVec(utils.ScalVecMult(1/r, utils.VecDuplicate(Y, m, h))), params.PlaintextModulus())
	fmt.Println("Ysens", Ysens)
	Ypacked := bgv.NewPlaintext(params, params.MaxLevel())
	encoder.Encode(Ysens, Ypacked)
	Ycin, _ := encryptor.EncryptNew(Ypacked)

	// 여기서 컨트롤러에 들어가는 Ycin가 나온다 암호문 4개짜리

	// **** Encrypted controller ****
	Uout, _ := eval.MulNew(ctHy[0], ctY[0])
	eval.MulThenAdd(ctHu[0], ctU[0], Uout)
	fmt.Println("ctHu", ctHu)
	for j := 1; j < n; j++ {
		eval.MulThenAdd(ctHy[j], ctY[j], Uout)
		eval.MulThenAdd(ctHu[j], ctU[j], Uout)
	}
	//fmt.Println("Uout=", Uout)

	// 여기서 Uout이 컨트롤러 아웃풋인데

	// **** Actuator ****
	// Plant input
	U := make([]float64, m)
	// Unpacked and re-scaled u at actuator
	Uact := make([]uint64, params.N())
	// u after inner sum
	Usum := make([]uint64, m)
	encoder.Decode(decryptor.DecryptNew(Uout), Uact)
	// 여기서 Uact가 Uout을 복호화하고 언패킹
	// fmt.Println("Uact", Uact) // 얘는 N개만큼 데이터

	// Generate plant input // 아까 복제한 애들을 다시 잘 뭉쳐서 U만들기
	for k := 0; k < m; k++ {
		Usum[k] = utils.VecSumUint(Uact[k*h:(k+1)*h], params.PlaintextModulus(), bredparams)
		U[k] = float64(r * s * utils.SignFloat(float64(Usum[k]), params.PlaintextModulus()))
	}
	fmt.Println("U", U)
	// Re-encryption
	Upacked := bgv.NewPlaintext(params, params.MaxLevel())
	encoder.Encode(utils.ModVecFloat(utils.RoundVec(utils.ScalVecMult(1/r, utils.VecDuplicate(U, m, h))), params.PlaintextModulus()), Upacked)
	Ucin, _ := encryptor.EncryptNew(Upacked)
	//fmt.Println("Ucin", Ucin)
	// **** Encrypted Controller ****
	// State update

	ctY = append(ctY[1:], Ycin)
	ctU = append(ctU[1:], Ucin)
	fmt.Println("ctY", ctY)

	// **** Plant ****
	// State update
	xp = utils.VecAdd(utils.MatVecMult(A, xp), utils.MatVecMult(B, U))

	// Save data
	yEnc = append(yEnc, Y)
	uEnc = append(uEnc, U)
	xpEnc = append(xpEnc, xp)

	// Compare plant input between 1) and 2)
	uDiff := make([][]float64, iter)
	for i := range uDiff {
		uDiff[i] = []float64{utils.Vec2Norm(utils.VecSub(uUnenc[i], uEnc[i]))}
	}

	// 실제 분리 된 상황에서 돌리기 위한 암호화 된 ARX 행렬과 Y,U 초기 시퀀스

	//============== 저장해서 컨트롤러에 넘겨줄 데이터 =========================//

	// Create a 2D slice to hold the binary data for each ctHu element
	var ctHu_bin [][]byte
	var ctHy_bin [][]byte

	// Loop through ctHu and get the binary data for each element
	for i := 1; i <= 4; i++ {
		ctHu_bin_data, err := ctHu[i-1].MarshalBinary() // Get the binary data for each ctHu[i]
		ctHy_bin_data, err := ctHy[i-1].MarshalBinary()

		if err != nil {
			log.Fatalf("Failed to marshal ctHu[%d]: %v", i, err)
		}
		ctHu_bin = append(ctHu_bin, ctHu_bin_data)
		ctHy_bin = append(ctHy_bin, ctHy_bin_data)

	}

	ctY_bin, _ := Ycin.MarshalBinary()
	ctU_bin, _ := Ucin.MarshalBinary()
	saveBinaryDataToSingleColumnCSV(sk_bin, "sk_bin_data.csv")
	saveBinaryDataToCSV(ctHy_bin, "ctHy_bin_data.csv")
	saveBinaryDataToCSV(ctHu_bin, "ctHu_bin_data.csv")
	saveBinaryDataToSingleColumnCSV(ctY_bin, "ctY_bin_data.csv")
	saveBinaryDataToSingleColumnCSV(ctU_bin, "ctU_bin_data.csv")
	// =========== Export data ===========

}
