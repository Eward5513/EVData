import requests
import json

# 构造 Overpass QL 查询语句
query = f"""
[out:json][timeout:3000];
// 查询边界内的所有way
way["highway"~"motorway|trunk|primary|secondary|tertiary|residential|unclassified"](poly:"30.6974858 120.8557546 30.6974858 122.0158854 31.8610639 122.0158854 31.8610639 120.8557546");
// 输出结果
out body;
>;
out skel qt;
    """

# Overpass API endpoint
url = "http://overpass-api.de/api/interpreter"

# 发送 POST 请求
print("Requesting data from Overpass API...")
response = requests.post(url, data={"data": query})
response.raise_for_status()

data = response.json()

# 保存为 GeoJSON 格式
def osm_to_geojson(osm_data):
    from osm2geojson import json2geojson
    return json2geojson(osm_data)

try:
    from osm2geojson import json2geojson
    geojson_data = json2geojson(data)
except ImportError:
    print("模块 'osm2geojson' 未安装，请运行：pip install osm2geojson")
    exit(1)

# 写入文件
output_file = "shanghai_roads.geojson"
with open(output_file, "w", encoding="utf-8") as f:
    json.dump(geojson_data, f, ensure_ascii=False, indent=2)

print(f"GeoJSON 文件已保存为 {output_file}")
