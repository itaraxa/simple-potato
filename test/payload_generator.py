import random
import os
import time
from datetime import datetime
from string import ascii_letters

target_path = r"test/sender/new/"

def generate_file(file_name) -> None:
    
    with open(file_name, "wt", encoding="utf-8") as fout:
        fileLen = random.randint(100, 2000) * 1024
        for i in range(fileLen // 64):
            s = "".join([str(random.choice(ascii_letters)) for _ in range(64)]) + "\n"
            fout.write(s)
        print(f"{datetime.utcnow()} File {file_name} ({fileLen} bytes) created")

if __name__ == "__main__":
    os.chdir(target_path)
    k = 1
    while True:
        generate_file(f"{k}.txt")
        k += 1
        time.sleep(random.random() * 10)
