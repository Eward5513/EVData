import requests
import json

OUTPUT_FILE = "/home/tzhang174/EVData/traffic_light/frontend/shanghai_new.geojson"
OVERPASS_URL = "https://overpass-api.de/api/interpreter"

QUERY = """
[out:json][timeout:3000];
way["highway"~"motorway|trunk|primary|secondary|tertiary|residential|unclassified"](poly:"31.2036766 121.1236249 31.2036766 121.3644594 31.3642665 121.3644594 31.3642665 121.1236249");
out body;
>;
out skel qt;
""".strip()


def osm_to_geojson(osm_data):
    try:
        from osm2geojson import json2geojson
    except ImportError:
        print("未安装 osm2geojson，请先运行：pip install osm2geojson")
        raise
    return json2geojson(osm_data)


def main():
    print("请求 Overpass API 获取道路数据（GeoJSON）...")
    resp = requests.post(OVERPASS_URL, data=QUERY)
    resp.raise_for_status()
    osm_json = resp.json()
    geojson = osm_to_geojson(osm_json)
    with open(OUTPUT_FILE, "w", encoding="utf-8") as f:
        json.dump(geojson, f, ensure_ascii=False, indent=2)
    print(f"已保存到 {OUTPUT_FILE}")


if __name__ == "__main__":
    main()


