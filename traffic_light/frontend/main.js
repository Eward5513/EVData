// Configure Cesium base URL for local assets/workers
window.CESIUM_BASE_URL = "./Cesium-1.127/Build/Cesium/";

// Paths for data
const GROUPS_GEOJSON_URL = "./intersections.geojson";

// Use OpenStreetMap imagery to avoid Cesium ion token requirements

Cesium.Ion.defaultAccessToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI1YjFhYTRjZS0zYzZlLTRmN2ItOTE5NC1mMzEwYjFiZjE3NTUiLCJpZCI6MzEzNjA3LCJpYXQiOjE3NTAzMDY1MjF9.k7exedEe-OwSQ2qgC5NNIMec5tXhTiCEp6of6vdYv0o';

// Create Cesium viewer with OSM basemap and default ellipsoid terrain
const viewer = new Cesium.Viewer('viewer', {
    imageryProvider: new Cesium.OpenStreetMapImageryProvider({
        url: 'https://a.tile.openstreetmap.org/'
    }),
    terrainProvider: new Cesium.EllipsoidTerrainProvider(),
    timeline: false,
    animation: false,
    sceneModePicker: false,
    baseLayerPicker: true, 
    geocoder: false,
    homeButton: false,
    infoBox: true,
    selectionIndicator: true,
    navigationHelpButton: false,
    navigationInstructionsInitiallyVisible: false
});

viewer.scene.globe.depthTestAgainstTerrain = true;

// Deterministic color by group id
function colorForGroupId(groupId) {
  if (!Number.isFinite(groupId)) return Cesium.Color.WHITE;
  const hue = (groupId * 137.508) % 360; // golden angle for separation
  const saturation = 0.75;
  const lightness = 0.5;
  return Cesium.Color.fromHsl(hue / 360, saturation, lightness, 1.0);
}
 
function colorForPointType(type) {
  // Distinguish different point types
  if (type === 'near_node' || type === 'degree_node') return Cesium.Color.YELLOW;
  // default and 'signal'
  return Cesium.Color.RED;
}

// Load prebuilt grouped intersections GeoJSON and style by group
Cesium.GeoJsonDataSource.load(GROUPS_GEOJSON_URL, {
  clampToGround: true
}).then(function(ds) {
  viewer.dataSources.add(ds);
  const entities = ds.entities.values;
  // Build mapping: way_id -> set(group_id)
  const wayGroups = new Map();
  for (let i = 0; i < entities.length; i++) {
    const e = entities[i];
    if (!Cesium.defined(e.polyline)) continue;
    const props = e.properties;
    if (!props) continue;
    const wayId = props.way_id ? props.way_id.getValue() : undefined;
    const gid = props.group_id ? props.group_id.getValue() : undefined;
    if (wayId == null || gid == null) continue;
    if (!wayGroups.has(wayId)) wayGroups.set(wayId, new Set());
    wayGroups.get(wayId).add(gid);
  }

  for (let i = 0; i < entities.length; i++) {
    const e = entities[i];
    const props = e.properties;
    const gid = props && props.group_id ? props.group_id.getValue() : undefined;
    const typeProp = props && props.type ? props.type.getValue() : undefined;
    const lineColor = colorForGroupId(gid).withAlpha(0.95);
    if (Cesium.defined(e.polyline)) {
      e.polyline.material = lineColor;
      e.polyline.width = 3;
      e.polyline.clampToGround = true;
    } else if (Cesium.defined(e.point) || Cesium.defined(e.billboard)) {
      // Replace any billboard with styled point to avoid black box icons
      if (Cesium.defined(e.billboard)) {
        e.billboard = undefined;
      }
      if (!Cesium.defined(e.point)) {
        e.point = new Cesium.PointGraphics();
      }
      const ptColor = colorForPointType(typeProp);
      const size = typeProp === 'near_node' ? 8 : 10;
      e.point.color = ptColor;
      e.point.pixelSize = size;
      e.point.outlineColor = Cesium.Color.BLACK.withAlpha(0.7);
      e.point.outlineWidth = 1.5;
    }
  }
  // Set info box description for ways to include all group_ids this way belongs to
  for (let i = 0; i < entities.length; i++) {
    const e = entities[i];
    if (!Cesium.defined(e.polyline)) continue;
    const props = e.properties;
    if (!props) continue;
    const wayId = props.way_id ? props.way_id.getValue() : undefined;
    if (wayId == null) continue;
    const groupsSet = wayGroups.get(wayId);
    const groupList = groupsSet ? Array.from(groupsSet).sort((a, b) => a - b) : [];
    const groupText = groupList.length ? groupList.join(', ') : 'N/A';
    const name = props.name ? props.name.getValue() : '';
    e.name = name || `Way ${wayId}`;
    e.description = `
      <div style="font-family: sans-serif; font-size: 13px; line-height: 1.6;">
        <div><b>Way ID</b>: ${wayId}</div>
        <div><b>Group IDs</b>: ${groupText}</div>
      </div>
    `;
  }
  return viewer.flyTo(ds, { duration: 1.6 });
}).catch(function(err) {
  console.error("Failed to load intersections GeoJSON:", err);
  alert("加载 intersections.geojson 失败。请先运行后端脚本生成该文件，并通过本地 HTTP 服务访问页面。");
});
 

