import argparse
import json
import sys
from pathlib import Path

import requests


OVERPASS_URL = "http://overpass-api.de/api/interpreter"

# Same polygon as in download_geojson.py
SH_POLY = (
    "31.2036766 121.1236249 "
    "31.2036766 121.3644594 "
    "31.3642665 121.3644594 "
    "31.3642665 121.1236249"
)


def build_overpass_query(polygon_wgs84: str) -> str:
    # Query all nodes tagged as traffic signals within the polygon area
    return f"""
[out:json][timeout:3000];
node["highway"="traffic_signals"](poly:"{polygon_wgs84}");
out body;
>;
out skel qt;
""".strip()


def fetch_overpass(query: str) -> dict:
    print("Requesting traffic signals from Overpass API...")
    response = requests.post(OVERPASS_URL, data={"data": query})
    response.raise_for_status()
    return response.json()


def convert_to_geojson(osm_json: dict) -> dict:
    try:
        from osm2geojson import json2geojson
    except ImportError:
        print("模块 'osm2geojson' 未安装，请运行：pip install osm2geojson")
        sys.exit(1)
    return json2geojson(osm_json)


def write_geojson(geojson_data: dict, output_path: Path) -> None:
    output_path.parent.mkdir(parents=True, exist_ok=True)
    with output_path.open("w", encoding="utf-8") as f:
        json.dump(geojson_data, f, ensure_ascii=False, indent=2)
    print(f"GeoJSON 文件已保存为 {output_path}")


def main() -> None:
    parser = argparse.ArgumentParser(description="Download traffic signal nodes in Shanghai polygon and save as GeoJSON.")
    parser.add_argument(
        "--output",
        default="/home/tzhang174/EVData/traffic_light/frontend/shanghai_traffic_lights.geojson",
        help="Output GeoJSON file path (default: traffic_light/frontend/shanghai_traffic_lights.geojson)",
    )
    args = parser.parse_args()

    query = build_overpass_query(SH_POLY)
    osm_json = fetch_overpass(query)
    geojson_data = convert_to_geojson(osm_json)
    write_geojson(geojson_data, Path(args.output))


if __name__ == "__main__":
    main()


