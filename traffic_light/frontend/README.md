# Frontend (Cesium)

Static frontend using CesiumJS to display Shanghai GeoJSON.

## Serve locally
Use any static HTTP server. Two simple options:

Option A (Python):
```bash
cd /home/tzhang174/EVData/traffic_light/frontend
python3 -m http.server 5173
```
Open `http://localhost:5173/` in your browser.

Option B (Node):
```bash
npx --yes http-server -p 5173 -c-1 .
```

## Generate grouped intersections GeoJSON
From the backend directory:
```bash
python3 /home/tzhang174/EVData/traffic_light/backend/build_intersections.py
```
This will produce:
- `/home/tzhang174/EVData/traffic_light/backend/intersections.json` (raw groups data)
- `/home/tzhang174/EVData/traffic_light/frontend/intersections.geojson` (for the frontend)

## What the frontend shows
- Loads `intersections.geojson` on startup
- Draws each group's traffic signals (Point) and ways (LineString)
- Uses a distinct color per `group_id`


