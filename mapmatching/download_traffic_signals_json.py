import requests
import json
import time
import argparse

# 同区域多边形（与现有查询保持一致）
SH_POLY = (
    "31.2036766 121.1236249 "
    "31.2036766 121.3644594 "
    "31.3642665 121.3644594 "
    "31.3642665 121.1236249"
)


def download_traffic_signals(output_file="/home/tzhang174/EVData/traffic_light/backend/shanghai_traffic_lights.json"):
    overpass_url = "https://overpass-api.de/api/interpreter"

    # 仅下载红绿灯节点（highway=traffic_signals）
    query = f"""
[out:json][timeout:3000];
node["highway"="traffic_signals"](poly:"{SH_POLY}");
out body;
""".strip()

    try:
        print("正在从Overpass API下载红绿灯节点...")
        start_time = time.time()

        response = requests.post(overpass_url, data=query)
        response.raise_for_status()

        data = response.json()

        with open(output_file, "w") as f:
            json.dump(data, f, indent=2)

        download_time = time.time() - start_time
        print(f"红绿灯节点下载完成！耗时 {download_time:.2f} 秒")
        print(f"数据已保存到: {output_file}")

        return data

    except requests.exceptions.RequestException as e:
        print(f"下载数据时出错: {e}")
        return None


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="下载指定区域内的红绿灯节点（JSON 原始数据）。")
    parser.add_argument(
        "--output",
        default="/home/tzhang174/EVData/traffic_light/backend/shanghai_traffic_lights.json",
        help="输出文件路径（默认：/home/tzhang174/EVData/traffic_light/backend/shanghai_traffic_lights.json）",
    )
    args = parser.parse_args()
    download_traffic_signals(args.output)


