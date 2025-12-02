from pathlib import Path
import json
from typing import Any, Dict, List

from flask import Flask, jsonify, abort
from flask_cors import CORS


def create_app() -> Flask:
    app = Flask(__name__)
    CORS(app)

    data_path = Path("/home/tzhang174/EVData/traffic_light/backend/intersections.json")
    intersections: List[Dict[str, Any]] = []
    intersections_index: Dict[int, Dict[str, Any]] = {}

    def group_to_feature_collection(group: Dict[str, Any]) -> Dict[str, Any]:
        gid = group.get("id")
        feature_collection: Dict[str, Any] = {"type": "FeatureCollection", "features": []}
        # Signals
        for sn in group.get("signal_nodes", []):
            lat = sn.get("lat")
            lon = sn.get("lon")
            if lat is None or lon is None:
                continue
            feature_collection["features"].append(
                {
                    "type": "Feature",
                    "geometry": {"type": "Point", "coordinates": [lon, lat]},
                    "properties": {
                        "type": "signal",
                        "group_id": gid,
                        "signal_id": sn.get("id"),
                        "tags": sn.get("tags", {}),
                    },
                }
            )
        # Near nodes
        for nn in group.get("near_nodes", []):
            lat = nn.get("lat")
            lon = nn.get("lon")
            if lat is None or lon is None:
                continue
            feature_collection["features"].append(
                {
                    "type": "Feature",
                    "geometry": {"type": "Point", "coordinates": [lon, lat]},
                    "properties": {
                        "type": "near_node",
                        "group_id": gid,
                        "node_id": nn.get("id"),
                    },
                }
            )
        # Ways
        for w in group.get("ways", []):
            coords = w.get("nodes", [])
            if not coords or len(coords) < 2:
                continue
            feature_collection["features"].append(
                {
                    "type": "Feature",
                    "geometry": {"type": "LineString", "coordinates": coords},
                    "properties": {
                        "type": "way",
                        "group_id": gid,
                        "way_id": w.get("id"),
                        "tags": w.get("tags", {}),
                    },
                }
            )
        return feature_collection

    def load_data() -> None:
        nonlocal intersections, intersections_index
        if not data_path.exists():
            intersections = []
            intersections_index = {}
            return
        with data_path.open("r", encoding="utf-8") as f:
            intersections = json.load(f)
        intersections_index = {int(item["id"]): item for item in intersections if "id" in item}

    # Load once at startup
    load_data()

    @app.get("/api/intersections")
    def get_intersections():
        if not intersections:
            load_data()
        items = []
        for it in intersections:
            items.append(
                {
                    "id": it.get("id"),
                    "centroid": it.get("centroid"),
                    "signals_count": len(it.get("signal_nodes", [])),
                    "ways_count": len(it.get("ways", [])),
                }
            )
        return jsonify(items)

    @app.get("/api/intersections/<int:group_id>")
    def get_intersection(group_id: int):
        if not intersections_index:
            load_data()
        item = intersections_index.get(group_id)
        if not item:
            abort(404, description=f"Group {group_id} not found")
        return jsonify(item)

    @app.get("/api/intersections/<int:group_id>/geojson")
    def get_intersection_geojson(group_id: int):
        if not intersections_index:
            load_data()
        item = intersections_index.get(group_id)
        if not item:
            abort(404, description=f"Group {group_id} not found")
        return jsonify(group_to_feature_collection(item))

    @app.get("/api/intersections/features")
    def get_intersections_features():
        """
        Query param:
          - ids: comma-separated list of group ids, e.g. ?ids=1,2,3
        Returns a GeoJSON FeatureCollection that merges the requested groups.
        """
        if not intersections_index:
            load_data()
        from flask import request

        ids_param = request.args.get("ids", "")
        if not ids_param.strip():
            abort(400, description="Missing query parameter 'ids'")
        raw_ids = [s.strip() for s in ids_param.split(",") if s.strip()]
        group_ids: List[int] = []
        for s in raw_ids:
            try:
                n = int(s)
                group_ids.append(n)
            except ValueError:
                continue
        if not group_ids:
            abort(400, description="No valid group ids found in 'ids'")

        features: List[Dict[str, Any]] = []
        for gid in group_ids:
            item = intersections_index.get(gid)
            if not item:
                # skip missing ids rather than fail entire request
                continue
            fc = group_to_feature_collection(item)
            features.extend(fc.get("features", []))
        return jsonify({"type": "FeatureCollection", "features": features})

    return app


app = create_app()


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8000, debug=True)