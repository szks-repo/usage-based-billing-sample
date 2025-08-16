import pandas as pd
import sys

file_path = './data/0198b15b-938b-7220-a91c-291b0bad0de9.parquet' 

try:
    df = pd.read_parquet(file_path)

    print("------------------------------------")
    print(df.head())
    print("\n------------------------------------")

except Exception as e:
    print(f"error: {e}")
    sys.exit(1)
