import pandas as pd
import sys

file_path = './data/0198b8b3-e491-7345-b313-332dc01fc298.parquet' 

try:
    df = pd.read_parquet(file_path)

    print("------------------------------------")
    print(df)
    print("\n------------------------------------")

except Exception as e:
    print(f"error: {e}")
    sys.exit(1)
