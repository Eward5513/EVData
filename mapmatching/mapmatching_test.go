package mapmatching_test

import (
	"EVdata/common"
	"EVdata/mapmatching"
	"EVdata/proto_struct"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"log"
	"sync"
)

var _ = Describe("mapmatching test", func() {
	mapmatching.DistanceOffset = 40

	graph := mapmatching.BuildGraph("shanghai_new.json")

	log.Println("MIN_LONGITUDE", common.MIN_LONGITUDE)
	log.Println("MIN_LATITUDE", common.MIN_LATITUDE)
	log.Println("MAX_LATITUDE", common.MAX_LATITUDE)
	log.Println("MAX_LONGITUDE", common.MAX_LONGITUDE)

	indexRoot := mapmatching.BuildIndex(graph)
	mmp := common.NewMemoryMap()

	log.Println("Finish building index")

	Context("CandidateSearch", func() {
		XIt("should get right result #1", MustPassRepeatedly(200), func() {
			p := &proto_struct.RawPoint{
				Longitude: 121.129475,
				Latitude:  31.317084,
			}
			cps := mapmatching.CandidateSearch(p, indexRoot, mmp, 0)

			printCandidatePoint(cps[0])
			printCandidatePoint(cps[1])

			Expect(len(cps)).Should(Equal(2))

			Expect(cps).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Lat":      BeNumerically("~", 31.3170233, 1e-7),
					"Lon":      BeNumerically("~", 121.1298425, 1e-7),
					"Distance": BeNumerically("~", 35.56, 0.01),
					"Ep":       BeNumerically("~", 0.996, 1e-3),
					//"TT":       BeNumerically("~", -0.43, 0.01), // 起始点不同t不同
					"RoadID": Equal(int64(405790958)),
					"Ttype":  Equal(common.NORMAL),
					"Vertex": ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							//31.3170233, 121.1298425
							"Lat": BeNumerically("~", 31.3170233, 1e-7),
							"Lon": BeNumerically("~", 121.1298425, 1e-7),
							"Id":  Equal(int64(3425116687)),
						})),
						PointTo(MatchFields(IgnoreExtras, Fields{
							//31.3166365 / 121.1300384
							"Lat": BeNumerically("~", 31.3166365, 1e-7),
							"Lon": BeNumerically("~", 121.1300384, 1e-7),
							"Id":  Equal(int64(3425116351)),
						})),
					),
					"OriginalPoint": PointTo(MatchFields(IgnoreExtras, Fields{
						"Longitude": BeNumerically("~", 121.129475, 1e-6),
						"Latitude":  BeNumerically("~", 31.317084, 1e-6),
					})),
				})),
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Lat":      BeNumerically("~", 31.317190898575586, 1e-7),
					"Lon":      BeNumerically("~", 121.12975598990542, 1e-7),
					"Distance": BeNumerically("~", 29.22, 0.01),
					"Ep":       BeNumerically("~", 0.997, 1e-3),
					//"TT":       BeNumerically("~", 0.66, 0.01),
					"RoadID": Equal(int64(405790958)),
					"Ttype":  Equal(common.NORMAL),
					"Vertex": ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							//31.3170233, 121.1298425
							"Lat": BeNumerically("~", 31.3170233, 1e-7),
							"Lon": BeNumerically("~", 121.1298425, 1e-7),
							"Id":  Equal(int64(3425116687)),
						})),
						PointTo(MatchFields(IgnoreExtras, Fields{
							//31.317521 / 121.1295856
							"Lat": BeNumerically("~", 31.317521, 1e-7),
							"Lon": BeNumerically("~", 121.1295856, 1e-7),
							"Id":  Equal(int64(3425116689)),
						})),
					),
					"OriginalPoint": PointTo(MatchFields(IgnoreExtras, Fields{
						"Longitude": BeNumerically("~", 121.129475, 1e-6),
						"Latitude":  BeNumerically("~", 31.317084, 1e-6),
					})),
				})),
			))
		})
		It("should get right result #2", MustPassRepeatedly(1), func() {
			p := &proto_struct.RawPoint{
				Longitude: 121.12984,
				Latitude:  31.316975,
			}
			cps := mapmatching.CandidateSearch(p, indexRoot, mmp, 0)

			printCandidatePoint(cps[0])
			printCandidatePoint(cps[1])

			Expect(len(cps)).Should(Equal(2))

			//Expect(cps).To(ConsistOf(
			//	PointTo(MatchFields(IgnoreExtras, Fields{
			//		"Lat":      BeNumerically("~", 31.3170233, 1e-7),
			//		"Lon":      BeNumerically("~", 121.1298425, 1e-7),
			//		"Distance": BeNumerically("~", 35.56, 0.01),
			//		"Ep":       BeNumerically("~", 0.996, 1e-3),
			//		//"TT":       BeNumerically("~", -0.43, 0.01), // 起始点不同t不同
			//		"RoadID": Equal(int64(405790958)),
			//		"Ttype":  Equal(common.NORMAL),
			//		"Vertex": ConsistOf(
			//			PointTo(MatchFields(IgnoreExtras, Fields{
			//				//31.3170233, 121.1298425
			//				"Lat": BeNumerically("~", 31.3170233, 1e-7),
			//				"Lon": BeNumerically("~", 121.1298425, 1e-7),
			//				"Id":  Equal(int64(3425116687)),
			//			})),
			//			PointTo(MatchFields(IgnoreExtras, Fields{
			//				//31.3166365 / 121.1300384
			//				"Lat": BeNumerically("~", 31.3166365, 1e-7),
			//				"Lon": BeNumerically("~", 121.1300384, 1e-7),
			//				"Id":  Equal(int64(3425116351)),
			//			})),
			//		),
			//		"OriginalPoint": PointTo(MatchFields(IgnoreExtras, Fields{
			//			"Longitude": BeNumerically("~", 121.129475, 1e-6),
			//			"Latitude":  BeNumerically("~", 31.317084, 1e-6),
			//		})),
			//	})),
			//	PointTo(MatchFields(IgnoreExtras, Fields{
			//		"Lat":      BeNumerically("~", 31.317190898575586, 1e-7),
			//		"Lon":      BeNumerically("~", 121.12975598990542, 1e-7),
			//		"Distance": BeNumerically("~", 29.22, 0.01),
			//		"Ep":       BeNumerically("~", 0.997, 1e-3),
			//		//"TT":       BeNumerically("~", 0.66, 0.01),
			//		"RoadID": Equal(int64(405790958)),
			//		"Ttype":  Equal(common.NORMAL),
			//		"Vertex": ConsistOf(
			//			PointTo(MatchFields(IgnoreExtras, Fields{
			//				//31.3170233, 121.1298425
			//				"Lat": BeNumerically("~", 31.3170233, 1e-7),
			//				"Lon": BeNumerically("~", 121.1298425, 1e-7),
			//				"Id":  Equal(int64(3425116687)),
			//			})),
			//			PointTo(MatchFields(IgnoreExtras, Fields{
			//				//31.317521 / 121.1295856
			//				"Lat": BeNumerically("~", 31.317521, 1e-7),
			//				"Lon": BeNumerically("~", 121.1295856, 1e-7),
			//				"Id":  Equal(int64(3425116689)),
			//			})),
			//		),
			//		"OriginalPoint": PointTo(MatchFields(IgnoreExtras, Fields{
			//			"Longitude": BeNumerically("~", 121.129475, 1e-6),
			//			"Latitude":  BeNumerically("~", 31.317084, 1e-6),
			//		})),
			//	})),
			//))
		})
		XIt("should get right result for discrete point", MustPassRepeatedly(200), func() {
			p := &proto_struct.RawPoint{
				Longitude: 121.129333,
				Latitude:  31.31704,
			}

			cps := mapmatching.CandidateSearch(p, indexRoot, mmp, 0)

			//printCandidatePoint(cps[0])

			Expect(len(cps)).Should(Equal(1))

			Expect(cps).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Lat":      BeNumerically("~", 31.317199275458876, 1e-7),
					"Lon":      BeNumerically("~", 121.1297516659727, 1e-7),
					"Distance": BeNumerically("~", 43.53, 0.01),
					//"Ep":       BeNumerically("~", 0.997, 1e-3),
					//"TT":       BeNumerically("~", 0.66, 0.01),
					"RoadID": Equal(int64(405790958)),
					"Ttype":  Equal(common.DISCRETE),
					"Vertex": ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							//31.3170233, 121.1298425
							"Lat": BeNumerically("~", 31.3170233, 1e-7),
							"Lon": BeNumerically("~", 121.1298425, 1e-7),
							"Id":  Equal(int64(3425116687)),
						})),
						PointTo(MatchFields(IgnoreExtras, Fields{
							//31.317521 / 121.1295856
							"Lat": BeNumerically("~", 31.317521, 1e-7),
							"Lon": BeNumerically("~", 121.1295856, 1e-7),
							"Id":  Equal(int64(3425116689)),
						})),
					),
					"OriginalPoint": PointTo(MatchFields(IgnoreExtras, Fields{
						"Longitude": BeNumerically("~", 121.129333, 1e-6),
						"Latitude":  BeNumerically("~", 31.31704, 1e-5),
					})),
				})),
			))
		})
	})
	XContext("Worker test", func() {
		It("should get right result at beginning #1", MustPassRepeatedly(1), func() {
			msg := []*proto_struct.RawPoint{
				{
					Time:      "7:18:45",
					Vin:       1,
					Speed:     35.9,
					Longitude: 121.129235,
					Latitude:  31.317013,
				},
				{
					Time:      "7:18:46",
					Vin:       1,
					Speed:     31.4,
					Longitude: 121.129333,
					Latitude:  31.31704,
				},
				{
					Time:      "7:18:47",
					Vin:       1,
					Speed:     27.3,
					Longitude: 121.129413,
					Latitude:  31.317062,
				},
				{
					Time:      "7:18:48",
					Vin:       1,
					Speed:     24,
					Longitude: 121.129475,
					Latitude:  31.317084,
				},
				{
					Time:      "7:18:49",
					Vin:       1,
					Speed:     20.7,
					Longitude: 121.129546,
					Latitude:  31.3171,
				},
				{
					Time:      "7:18:50",
					Vin:       1,
					Speed:     17.7,
					Longitude: 121.129591,
					Latitude:  31.317115,
				},
				{
					Time:      "7:18:51",
					Vin:       1,
					Speed:     15.1,
					Longitude: 121.129635,
					Latitude:  31.317128,
				},
				{
					Time:      "7:18:52",
					Vin:       1,
					Speed:     13.2,
					Longitude: 121.12968,
					Latitude:  31.317137,
				},
				{
					Time:      "7:18:53",
					Vin:       1,
					Speed:     12.6,
					Longitude: 121.129706,
					Latitude:  31.317142,
				},
				{
					Time:      "7:18:54",
					Vin:       1,
					Speed:     12.4,
					Longitude: 121.129742,
					Latitude:  31.317140,
				},
				{
					Time:      "7:18:55",
					Vin:       1,
					Speed:     14.9,
					Longitude: 121.129760,
					Latitude:  31.317122,
				},
				{
					Time:      "7:18:56",
					Vin:       1,
					Speed:     19.4,
					Longitude: 121.129786,
					Latitude:  31.317084,
				},
				{
					Time:      "7:18:57",
					Vin:       1,
					Speed:     23.3,
					Longitude: 121.129804,
					Latitude:  31.317035,
				},
				{
					Time:      "7:18:58",
					Vin:       1,
					Speed:     27.2,
					Longitude: 121.129840,
					Latitude:  31.316975,
				},
				{
					Time:      "7:18:59",
					Vin:       1,
					Speed:     27.2,
					Longitude: 121.12986,
					Latitude:  31.316906,
				},
			}

			//expectMps := []*proto_struct.MatchingPoint{
			//	{
			//		OriginalLon: 121.129333,
			//		OriginalLat: 31.31704,
			//		IsBad:       int32(1),
			//	},
			//	{
			//		OriginalLon: 121.129413,
			//		OriginalLat: 31.317062,
			//		IsBad:       int32(1),
			//	},
			//	{
			//		OriginalLon: 121.129475,
			//		OriginalLat: 31.317084,
			//		IsBad:       int32(0),
			//	},
			//	{
			//		OriginalLon: 121.129546,
			//		OriginalLat: 31.3171,
			//		IsBad:       int32(0),
			//	},
			//	{
			//		OriginalLon: 121.129591,
			//		OriginalLat: 31.317115,
			//		IsBad:       int32(0),
			//	},
			//	{
			//		OriginalLon: 121.129635,
			//		OriginalLat: 31.317128,
			//		IsBad:       int32(0),
			//	},
			//	{
			//		OriginalLon: 121.12968,
			//		OriginalLat: 31.317137,
			//		IsBad:       int32(0),
			//	},
			//	{
			//		OriginalLon: 121.129706,
			//		OriginalLat: 31.317142,
			//		IsBad:       int32(0),
			//	},
			//	{
			//		OriginalLon: 121.129742,
			//		OriginalLat: 31.317140,
			//		IsBad:       int32(0),
			//	},
			//	{
			//		OriginalLon: 121.129760,
			//		OriginalLat: 31.317122,
			//		IsBad:       int32(0),
			//	},
			//	{
			//		OriginalLon: 121.129786,
			//		OriginalLat: 31.317084,
			//		IsBad:       int32(0),
			//	},
			//}

			rch := make(chan []*proto_struct.RawPoint, 1)
			wch := make(chan *proto_struct.Track, 1)
			wg := &sync.WaitGroup{}
			wg.Add(1)
			go mapmatching.Worker(rch, wch, indexRoot, wg)

			rch <- msg
			close(rch)

			res := <-wch

			wg.Wait()
			Expect(res.Vin).To(Equal(int32(1)))
			Expect(res.Tid).To(Equal(int32(0)))

			for i := range res.Rps {
				Expect(res.Rps[i]).To(BeIdenticalTo(msg[i]))
			}

			for _, mp := range res.Mps {
				log.Println(mp.OriginalLon, mp.OriginalLat, mp.MatchedLon, mp.MatchedLat, mp.RoadId, mp.IsBad)
			}
		})
	})
	XContext("MergePoints test", func() {
		It("should get right result #1", MustPassRepeatedly(1), func() {
			rps := []*proto_struct.RawPoint{
				{
					Time:      "7:18:45",
					Vin:       1,
					Speed:     35.9,
					Longitude: 121.129235,
					Latitude:  31.317013,
				},
				{
					Time:      "7:18:46",
					Vin:       1,
					Speed:     31.4,
					Longitude: 121.129333,
					Latitude:  31.31704,
				},
				{
					Time:      "7:18:47",
					Vin:       1,
					Speed:     27.3,
					Longitude: 121.129413,
					Latitude:  31.317062,
				},
				{
					Time:      "7:18:48",
					Vin:       1,
					Speed:     24,
					Longitude: 121.129475,
					Latitude:  31.317084,
				},
				{
					Time:      "7:18:49",
					Vin:       1,
					Speed:     20.7,
					Longitude: 121.129546,
					Latitude:  31.3171,
				},
				{
					Time:      "7:18:50",
					Vin:       1,
					Speed:     17.7,
					Longitude: 121.129591,
					Latitude:  31.317115,
				},
				{
					Time:      "7:18:51",
					Vin:       1,
					Speed:     15.1,
					Longitude: 121.129635,
					Latitude:  31.317128,
				},
				{
					Time:      "7:18:52",
					Vin:       1,
					Speed:     13.2,
					Longitude: 121.12968,
					Latitude:  31.317137,
				},
				{
					Time:      "7:18:53",
					Vin:       1,
					Speed:     12.6,
					Longitude: 121.129706,
					Latitude:  31.317142,
				},
				{
					Time:      "7:18:54",
					Vin:       1,
					Speed:     12.4,
					Longitude: 121.129742,
					Latitude:  31.317140,
				},
				{
					Time:      "7:18:55",
					Vin:       1,
					Speed:     14.9,
					Longitude: 121.129760,
					Latitude:  31.317122,
				},
				{
					Time:      "7:18:56",
					Vin:       1,
					Speed:     19.4,
					Longitude: 121.129786,
					Latitude:  31.317084,
				},
				{
					Time:      "7:18:57",
					Vin:       1,
					Speed:     23.3,
					Longitude: 121.129804,
					Latitude:  31.317035,
				},
				{
					Time:      "7:18:58",
					Vin:       1,
					Speed:     27.2,
					Longitude: 121.129840,
					Latitude:  31.316975,
				},
				{
					Time:      "7:18:59",
					Vin:       1,
					Speed:     27.2,
					Longitude: 121.12986,
					Latitude:  31.316906,
				},
			}

			mps := []*proto_struct.MatchingPoint{
				{
					OriginalLon: 121.129235,
					OriginalLat: 31.317013,
					MatchedLon:  121.12974722950413,
					MatchedLat:  31.3172078703612,
					IsBad:       int32(1),
				},
				{
					OriginalLon: 121.129333,
					OriginalLat: 31.31704,
					MatchedLon:  121.129333,
					MatchedLat:  31.31704,
					IsBad:       int32(1),
				},
				{
					OriginalLon: 121.129413,
					OriginalLat: 31.317062,
					MatchedLon:  121.1297553051899,
					MatchedLat:  31.31719222509531,
					IsBad:       int32(0),
				},
				{
					OriginalLon: 121.129475,
					OriginalLat: 31.317084,
					MatchedLon:  121.12975598990542,
					MatchedLat:  31.317190898575586,
					IsBad:       int32(0),
				},
				{
					OriginalLon: 121.129546,
					OriginalLat: 31.3171,
					MatchedLon:  121.12976074057215,
					MatchedLat:  31.317181694967825,
					IsBad:       int32(0),
				},
				{
					OriginalLon: 121.129591,
					OriginalLat: 31.317115,
					MatchedLon:  121.12976165507543,
					MatchedLat:  31.317179923273493,
					IsBad:       int32(0),
				},
				{
					OriginalLon: 121.129635,
					OriginalLat: 31.317128,
					MatchedLon:  121.12976326833983,
					MatchedLat:  31.31717679784845,
					IsBad:       int32(0),
				},
				{
					OriginalLon: 121.12968,
					OriginalLat: 31.317137,
					MatchedLon:  121.12976677154346,
					MatchedLat:  31.317170010988004,
					IsBad:       int32(0),
				},
				{
					OriginalLon: 121.129706,
					OriginalLat: 31.317142,
					MatchedLon:  121.12976888190668,
					MatchedLat:  31.317165922518654,
					IsBad:       int32(0),
				},
				{
					OriginalLon: 121.129742,
					OriginalLat: 31.317140,
					MatchedLon:  121.12977565381016,
					MatchedLat:  31.317152803108932,
					IsBad:       int32(0),
				},
				{
					OriginalLon: 121.129760,
					OriginalLat: 31.317122,
					MatchedLon:  121.12978637441293,
					MatchedLat:  31.317132033766764,
					IsBad:       int32(0),
				},
				{
					OriginalLon: 121.129786,
					OriginalLat: 31.317084,
					MatchedLon:  121.12980703712878,
					MatchedLat:  31.31709200327364,
					IsBad:       int32(0),
				},
				{
					OriginalLon: 121.129804,
					OriginalLat: 31.317035,
					MatchedLon:  121.12983113268348,
					MatchedLat:  31.31704532223991,
					IsBad:       int32(0),
				},
				{
					OriginalLon: 121.12984,
					OriginalLat: 31.316975,
					MatchedLon:  121.1298425,
					MatchedLat:  31.3170233,
					IsBad:       int32(0),
				},
				{
					OriginalLon: 121.12986,
					OriginalLat: 31.316906,
					MatchedLon:  121.1298425,
					MatchedLat:  31.3170233,
					IsBad:       int32(0),
				},
			}

			res := mapmatching.MergePoints(mps, rps)

			for i := range res {
				log.Println(res[i])
			}
		})
	})
})

func printCandidatePoint(p *common.CandidatePoint) {
	log.Println(p.Lat, p.Lon, p.Distance, p.RoadID, p.TT, p.Ttype, p.Ep, p.OriginalPoint.Latitude, p.OriginalPoint.Longitude)
}
