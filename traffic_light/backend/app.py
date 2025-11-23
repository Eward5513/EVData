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

    return app


app = create_app()


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8000, debug=True)