# Backend (Flask)

Flask backend serving intersection query APIs.

## Prerequisites
- Python 3.9+

## Setup
```bash
cd /home/tzhang174/EVData/traffic_light/backend
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

## Prepare data
先下载两份基础 JSON（均已固定输出到 backend 目录）：
```bash
python /home/tzhang174/EVData/traffic_light/backend/download_json.py
python /home/tzhang174/EVData/traffic_light/backend/download_traffic_signals_json.py
```

生成 intersections 文件（无命令行参数，路径写死在文件）：
```bash
python /home/tzhang174/EVData/traffic_light/backend/build_intersections.py
```

## Run
```bash
python app.py
```

The server will start on http://localhost:8000

## API
- `GET /api/intersections` → list of groups (id, centroid, counts)
- `GET /api/intersections/<id>` → full group data


