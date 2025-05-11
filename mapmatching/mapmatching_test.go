package main

import (
	"EVdata/common"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PreComputing", func() {
	var (
		graphNodes map[int64]*GraphNode
		node1      *GraphNode
		node2      *GraphNode
		node3      *GraphNode
		node4      *GraphNode
	)

	BeforeEach(func() {
		graphNodes = make(map[int64]*GraphNode)

		// 创建测试节点
		node1 = &GraphNode{
			id:           1,
			lat:          31.37876,
			lon:          121.5203,
			next:         make([]*Road, 0),
			shortestPath: make(map[*GraphNode]*Path),
		}
		node2 = &GraphNode{
			id:           2,
			lat:          31.37884,
			lon:          121.5198,
			next:         make([]*Road, 0),
			shortestPath: make(map[*GraphNode]*Path),
		}
		node3 = &GraphNode{
			id:           3,
			lat:          31.37889,
			lon:          121.5195,
			next:         make([]*Road, 0),
			shortestPath: make(map[*GraphNode]*Path),
		}
		node4 = &GraphNode{
			id:           4,
			lat:          31.37401,
			lon:          121.5253,
			next:         make([]*Road, 0),
			shortestPath: make(map[*GraphNode]*Path),
		}

		// 添加边（道路）
		node1.next = append(node1.next, &Road{ID: 1, Node: node2})
		node2.next = append(node2.next, &Road{ID: 1, Node: node1})
		node2.next = append(node2.next, &Road{ID: 2, Node: node3})
		node3.next = append(node3.next, &Road{ID: 2, Node: node2})
		node2.next = append(node2.next, &Road{ID: 3, Node: node4})
		node4.next = append(node4.next, &Road{ID: 3, Node: node2})

		// 将节点添加到图中
		graphNodes[1] = node1
		graphNodes[2] = node2
		graphNodes[3] = node3
		graphNodes[4] = node4
	})

	Context("当计算最短路径时", func() {
		BeforeEach(func() {
			PreComputing(graphNodes)
		})

		It("应该为每个节点创建到自身的路径", func() {
			for _, node := range graphNodes {
				path := node.shortestPath[node]
				Expect(path).NotTo(BeNil())
				Expect(path.distance).To(BeZero())
				Expect(path.points).To(HaveLen(1))
				Expect(path.points[0]).To(Equal(node))
			}
		})

		It("应该正确计算直接相连节点间的最短路径", func() {
			path := node1.shortestPath[node2]
			Expect(path).NotTo(BeNil())
			Expect(path.points).To(HaveLen(2))
			Expect(path.points[0]).To(Equal(node1))
			Expect(path.points[1]).To(Equal(node2))

			expectedDist := common.MAGIC_NUM * common.Distance(node1.lat, node1.lon, node2.lat, node2.lon)
			Expect(path.distance).To(BeNumerically("~", expectedDist, 0.0001))
		})

		It("应该正确计算间接相连节点间的最短路径", func() {
			path := node1.shortestPath[node3]
			Expect(path).NotTo(BeNil())
			Expect(path.points).To(HaveLen(3))
			Expect(path.points[0]).To(Equal(node1))
			Expect(path.points[1]).To(Equal(node2))
			Expect(path.points[2]).To(Equal(node3))
		})

		It("应该保持路径的对称性", func() {
			path1 := node1.shortestPath[node2]
			path2 := node2.shortestPath[node1]
			Expect(path1.distance).To(BeNumerically("~", path2.distance, 0.0001))
		})

		It("每个节点的最短路径数量不应超过100", func() {
			for _, node := range graphNodes {
				Expect(len(node.shortestPath)).To(BeNumerically("<=", 100))
			}
		})

		It("应该找到所有可达节点的路径", func() {
			for _, startNode := range graphNodes {
				for _, endNode := range graphNodes {
					path := startNode.shortestPath[endNode]
					Expect(path).NotTo(BeNil())
				}
			}
		})
	})

	Context("当处理边界情况时", func() {
		It("应该正确处理空图", func() {
			emptyGraph := make(map[int64]*GraphNode)
			PreComputing(emptyGraph)
			Expect(emptyGraph).To(BeEmpty())
		})

		It("应该正确处理只有一个节点的图", func() {
			singleNodeGraph := map[int64]*GraphNode{
				1: {
					id:           1,
					lat:          31.37876,
					lon:          121.5203,
					next:         make([]*Road, 0),
					shortestPath: make(map[*GraphNode]*Path),
				},
			}
			PreComputing(singleNodeGraph)
			node := singleNodeGraph[1]
			Expect(node.shortestPath).To(HaveLen(1))
			Expect(node.shortestPath[node].distance).To(BeZero())
		})
	})
})
