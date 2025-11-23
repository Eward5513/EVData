import requests
import json
import time

# 固定输出文件到 backend 目录
OUTPUT_FILE = "/home/tzhang174/EVData/traffic_light/backend/shanghai_new.json"
OVERPASS_URL = "https://overpass-api.de/api/interpreter"

def download_overpass_data():
    # 构建 Overpass QL 查询
    query = f"""
[out:json][timeout:3000];
// 查询边界内的所有way
way["highway"~"motorway|trunk|primary|secondary|tertiary|residential|unclassified"](poly:"31.2036766 121.1236249 31.2036766 121.3644594 31.3642665 121.3644594 31.3642665 121.1236249");
out geom;
node(w);
// 输出结果
out body;
>;
out skel qt;
    """
    try:
        print("正在从Overpass API下载道路数据...")
        start_time = time.time()
        # 发送POST请求（与现有可用脚本一致，直接 data=query）
        response = requests.post(OVERPASS_URL, data=query)
        response.raise_for_status()
        data = response.json()
        with open(OUTPUT_FILE, 'w') as f:
            json.dump(data, f, indent=2)
        download_time = time.time() - start_time
        print(f"数据下载完成！耗时 {download_time:.2f} 秒")
        print(f"数据已保存到: {OUTPUT_FILE}")
        return data
    except requests.exceptions.RequestException as e:
        print(f"下载数据时出错: {e}")
        return None

if __name__ == "__main__":
    download_overpass_data()


