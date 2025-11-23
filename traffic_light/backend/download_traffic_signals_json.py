import requests
import json
import time

# 固定输出文件到 backend 目录
OUTPUT_FILE = "/home/tzhang174/EVData/traffic_light/backend/shanghai_traffic_lights.json"
OVERPASS_URL = "https://overpass-api.de/api/interpreter"

# 同区域多边形（与道路查询一致）
SH_POLY = (
    "31.2036766 121.1236249 "
    "31.2036766 121.3644594 "
    "31.3642665 121.3644594 "
    "31.3642665 121.1236249"
)

def download_traffic_signals():
    # 仅下载红绿灯节点（highway=traffic_signals）
    query = f"""
[out:json][timeout:3000];
node["highway"="traffic_signals"](poly:"{SH_POLY}");
out body;
""".strip()
    try:
        print("正在从Overpass API下载红绿灯节点...")
        start_time = time.time()
        # 与已验证脚本保持一致：直接 data=query
        response = requests.post(OVERPASS_URL, data=query)
        response.raise_for_status()
        data = response.json()
        with open(OUTPUT_FILE, "w") as f:
            json.dump(data, f, indent=2)
        download_time = time.time() - start_time
        print(f"红绿灯节点下载完成！耗时 {download_time:.2f} 秒")
        print(f"数据已保存到: {OUTPUT_FILE}")
        return data
    except requests.exceptions.RequestException as e:
        print(f"下载数据时出错: {e}")
        return None

if __name__ == "__main__":
    download_traffic_signals()


