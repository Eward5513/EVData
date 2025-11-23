import argparse
import json
import math
import os
from collections import defaultdict, deque
from pathlib import Path
from typing import Dict, List, Tuple, Any, Set


# Epsilon (cluster radius) in meters for DBSCAN-like grouping
CLUSTER_EPS_METERS = 30.0
# Search radius in meters to associate nearby OSM nodes/ways to a group
WAY_NODE_RADIUS_METERS = 25.0

# Grid cell sizes in degrees (approx near lat ~31°)
# ~ 111_000 m per degree latitude; ~ 111_000 * cos(lat) per degree longitude
DEG_PER_M_LAT = 1.0 / 111_000.0
DEG_PER_M_LON_AT_31 = 1.0 / (111_000.0 * math.cos(math.radians(31.0)))

CELL_DEG_LAT = 0.0003  # ≈ 33 m
CELL_DEG_LON = 0.00035  # ≈ 33 m


def distance_meters(lat1: float, lon1: float, lat2: float, lon2: float) -> float:
    # Equirectangular approximation (good enough for small distances)
    mean_lat_rad = math.radians((lat1 + lat2) * 0.5)
    dx = (lon2 - lon1) * math.cos(mean_lat_rad) * 111_320.0
    dy = (lat2 - lat1) * 110_540.0
    return math.hypot(dx, dy)


def grid_key(lat: float, lon: float) -> Tuple[int, int]:
    return (int(math.floor(lat / CELL_DEG_LAT)), int(math.floor(lon / CELL_DEG_LON)))


def build_grid(points: List[Tuple[float, float]]) -> Dict[Tuple[int, int], List[int]]:
    grid: Dict[Tuple[int, int], List[int]] = defaultdict(list)
    for idx, (lat, lon) in enumerate(points):
        grid[grid_key(lat, lon)].append(idx)
    return grid


def iter_neighbor_cells(cell: Tuple[int, int]):
    ci, cj = cell
    for di in (-1, 0, 1):
        for dj in (-1, 0, 1):
            yield (ci + di, cj + dj)


def cluster_dbscan_connectivity(points: List[Tuple[float, float]], eps_m: float) -> List[List[int]]:
    """Min-samples = 1 connectivity clustering with neighbor threshold eps_m."""
    n = len(points)
    visited = [False] * n
    clusters: List[List[int]] = []
    grid = build_grid(points)
    for i in range(n):
        if visited[i]:
            continue
        queue = deque([i])
        visited[i] = True
        component: List[int] = []
        while queue:
            idx = queue.popleft()
            component.append(idx)
            lat_i, lon_i = points[idx]
            # Candidate neighbors from 3x3 cells
            for nb_cell in iter_neighbor_cells(grid_key(lat_i, lon_i)):
                for j in grid.get(nb_cell, []):
                    if visited[j]:
                        continue
                    lat_j, lon_j = points[j]
                    if distance_meters(lat_i, lon_i, lat_j, lon_j) <= eps_m:
                        visited[j] = True
                        queue.append(j)
        clusters.append(component)
    return clusters


def load_overpass_json(path: Path) -> Dict[str, Any]:
    with path.open("r", encoding="utf-8") as f:
        return json.load(f)


def extract_signal_nodes(signals_json: Dict[str, Any]) -> List[Dict[str, Any]]:
    res: List[Dict[str, Any]] = []
    for el in signals_json.get("elements", []):
        if el.get("type") != "node":
            continue
        res.append(
            {
                "id": el.get("id"),
                "lat": el.get("lat"),
                "lon": el.get("lon"),
                "tags": el.get("tags", {}),
            }
        )
    return res


def build_osm_indexes(roads_json: Dict[str, Any]):
    node_id_to_coord: Dict[int, Tuple[float, float]] = {}
    ways_by_id: Dict[int, Dict[str, Any]] = {}
    node_id_to_way_ids: Dict[int, Set[int]] = defaultdict(set)
    # For proximity search among OSM nodes
    node_points: List[Tuple[float, float]] = []
    node_ids_in_order: List[int] = []

    for el in roads_json.get("elements", []):
        if el.get("type") == "node":
            nid = el.get("id")
            lat = el.get("lat")
            lon = el.get("lon")
            if nid is None or lat is None or lon is None:
                continue
            node_id_to_coord[nid] = (lat, lon)
            node_points.append((lat, lon))
            node_ids_in_order.append(nid)
        elif el.get("type") == "way":
            wid = el.get("id")
            if wid is None:
                continue
            ways_by_id[wid] = el

    # nodes referenced by ways
    for wid, way in ways_by_id.items():
        for nid in way.get("nodes", []):
            node_id_to_way_ids[nid].add(wid)

    node_grid = build_grid(node_points)
    return node_id_to_coord, ways_by_id, node_id_to_way_ids, node_points, node_ids_in_order, node_grid


def find_osm_nodes_near(
    lat: float,
    lon: float,
    node_points: List[Tuple[float, float]],
    node_ids_in_order: List[int],
    node_grid: Dict[Tuple[int, int], List[int]],
    radius_m: float,
) -> Set[int]:
    result: Set[int] = set()
    cell = grid_key(lat, lon)
    for nb_cell in iter_neighbor_cells(cell):
        for idx in node_grid.get(nb_cell, []):
            nlat, nlon = node_points[idx]
            if distance_meters(lat, lon, nlat, nlon) <= radius_m:
                result.add(node_ids_in_order[idx])
    return result


def way_geometry_lonlat(way: Dict[str, Any], node_id_to_coord: Dict[int, Tuple[float, float]]) -> List[Tuple[float, float]]:
    coords: List[Tuple[float, float]] = []
    if "geometry" in way and way["geometry"]:
        for g in way["geometry"]:
            # Overpass 'geometry' entries are dicts with lat/lon
            coords.append((g["lon"], g["lat"]))
        return coords
    # Fallback reconstruct from 'nodes'
    for nid in way.get("nodes", []):
        latlon = node_id_to_coord.get(nid)
        if latlon is None:
            continue
        lat, lon = latlon
        coords.append((lon, lat))
    return coords


def build_intersections(
    roads_path: Path,
    signals_path: Path,
    output_path: Path,
) -> None:
    print(f"Loading roads: {roads_path}")
    roads_json = load_overpass_json(roads_path)
    print(f"Loading signals: {signals_path}")
    signals_json = load_overpass_json(signals_path)

    signal_nodes = extract_signal_nodes(signals_json)
    signal_points = [(n["lat"], n["lon"]) for n in signal_nodes]

    print(f"Clustering {len(signal_nodes)} signal nodes with eps={CLUSTER_EPS_METERS}m ...")
    clusters = cluster_dbscan_connectivity(signal_points, CLUSTER_EPS_METERS)
    print(f"Formed {len(clusters)} groups")

    print("Building OSM indexes for ways & nodes ...")
    node_id_to_coord, ways_by_id, node_id_to_way_ids, node_points, node_ids_in_order, node_grid = build_osm_indexes(roads_json)

    groups_output: List[Dict[str, Any]] = []
    for gid, comp in enumerate(clusters, start=1):
        # centroid
        if not comp:
            continue
        sum_lat = sum(signal_nodes[i]["lat"] for i in comp)
        sum_lon = sum(signal_nodes[i]["lon"] for i in comp)
        centroid = [sum_lon / len(comp), sum_lat / len(comp)]

        # collect nearby OSM nodes around each signal, then their ways
        nearby_node_ids: Set[int] = set()
        for i in comp:
            lat = signal_nodes[i]["lat"]
            lon = signal_nodes[i]["lon"]
            nearby_node_ids |= find_osm_nodes_near(lat, lon, node_points, node_ids_in_order, node_grid, WAY_NODE_RADIUS_METERS)

        way_ids: Set[int] = set()
        for nid in nearby_node_ids:
            way_ids |= node_id_to_way_ids.get(nid, set())

        ways_payload: List[Dict[str, Any]] = []
        for wid in sorted(way_ids):
            way = ways_by_id.get(wid)
            if not way:
                continue
            coords = way_geometry_lonlat(way, node_id_to_coord)
            if len(coords) < 2:
                continue
            ways_payload.append(
                {
                    "id": wid,
                    "tags": way.get("tags", {}),
                    "nodes": coords,  # [[lon,lat], ...]
                }
            )

        signals_payload = [
            {
                "id": signal_nodes[i]["id"],
                "lat": signal_nodes[i]["lat"],
                "lon": signal_nodes[i]["lon"],
                "tags": signal_nodes[i]["tags"],
            }
            for i in comp
        ]

        groups_output.append(
            {
                "id": gid,
                "centroid": centroid,
                "signal_nodes": signals_payload,
                "ways": ways_payload,
            }
        )

    output_path.parent.mkdir(parents=True, exist_ok=True)
    with output_path.open("w", encoding="utf-8") as f:
        json.dump(groups_output, f, ensure_ascii=False, indent=2)
    print(f"Wrote {len(groups_output)} groups to {output_path}")


def main():
    parser = argparse.ArgumentParser(description="Build intersection groups from signals and way network")
    parser.add_argument(
        "--roads",
        default="/home/tzhang174/EVData/traffic_light/backend/shanghai_new.json",
        help="Path to roads Overpass JSON (with ways and node(w))",
    )
    parser.add_argument(
        "--signals",
        default="/home/tzhang174/EVData/traffic_light/backend/shanghai_traffic_lights.json",
        help="Path to traffic signals Overpass JSON (nodes with highway=traffic_signals)",
    )
    parser.add_argument(
        "--output",
        default="/home/tzhang174/EVData/traffic_light/backend/intersections.json",
        help="Output file path for intersections JSON",
    )
    args = parser.parse_args()

    build_intersections(Path(args.roads), Path(args.signals), Path(args.output))


if __name__ == "__main__":
    main()


