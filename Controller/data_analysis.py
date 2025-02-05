import pandas as pd
import matplotlib.pyplot as plt

# CSV 파일 경로
file_path = "controller_data.csv"

# CSV 파일 읽기
data = pd.read_csv(file_path)

# 마지막 열(u) 데이터 추출
u = data["u"].values

# 이터레이션 (X축)
iterations = data["Iteration"].values

# 데이터 플롯
plt.figure(figsize=(10, 6))
plt.plot(iterations, u, label="Control Input (u)", marker="o", linestyle="-", color="b")

plt.xlabel("Iteration")
plt.ylabel("Control Input (u)")
plt.title("Control Input (u) Over Iterations")
plt.legend()
plt.grid(True)
plt.show()
