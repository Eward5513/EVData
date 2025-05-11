import requests
import json
import time

def download_overpass_data(polygon_coords, output_file="shanghai.json"):
    # Overpass API 端点
    overpass_url = "https://overpass-api.de/api/interpreter"
    
    # 构建Overpass QL查询
    query = f"""
[out:json][timeout:3000];
// 查询边界内的所有way
way["highway"~"motorway|trunk|primary|secondary|tertiary|residential|unclassified"](poly:"30.6974858 120.8557546 30.6974858 122.0158854 31.8610639 122.0158854 31.8610639 120.8557546");
out geom;
node(w);
// 输出结果
out body;
>;
out skel qt;
    """
    
    try:
        print("正在从Overpass API下载数据...")
        start_time = time.time()
        
        # 发送POST请求
        response = requests.post(overpass_url, data=query)
        response.raise_for_status()  # 检查请求是否成功
        
        data = response.json()
        
        # 将数据保存为JSON文件
        with open(output_file, 'w') as f:
            json.dump(data, f, indent=2)
        
        download_time = time.time() - start_time
        print(f"数据下载完成！耗时 {download_time:.2f} 秒")
        print(f"数据已保存到: {output_file}")
        
        return data
    
    except requests.exceptions.RequestException as e:
        print(f"下载数据时出错: {e}")
        return None

if __name__ == "__main__":
    # 定义多边形坐标（空格分隔的"lat lon"对）
    polygon_coords = "30.6974858 120.8557546 30.6974858 122.0158854 31.8610639 122.0158854 31.8610639 120.8557546"
    
    # 下载数据
    downloaded_data = download_overpass_data(polygon_coords)
    
    if downloaded_data:
        # 打印一些基本信息
        print(f"下载的元素数量: {len(downloaded_data['elements'])}")