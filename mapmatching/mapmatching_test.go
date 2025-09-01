package mapmatching_test

import (
	"EVdata/common"
	"EVdata/mapmatching"
	"EVdata/proto_struct"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"log"
)

var _ = Describe("mapmatching test", func() {
	//BeforeAll(func() {
	//})

	Context("CandidateSearch", func() {
		mapmatching.DistanceOffset = 40

		graph := mapmatching.BuildGraph("shanghai_new.json")

		log.Println("MIN_LONGITUDE", common.MIN_LONGITUDE)
		log.Println("MIN_LATITUDE", common.MIN_LATITUDE)
		log.Println("MAX_LATITUDE", common.MAX_LATITUDE)
		log.Println("MAX_LONGITUDE", common.MAX_LONGITUDE)

		indexRoot := mapmatching.BuildIndex(graph)
		mmp := common.NewMemoryMap()

		log.Println("Finish building index")

		It("should get right result #1", MustPassRepeatedly(10), func() {
			p := &proto_struct.RawPoint{
				Longitude: 121.129475,
				Latitude:  31.317084,
			}
			cps := mapmatching.CandidateSearch(p, indexRoot, mmp, 0)

			Expect(len(cps)).Should(Equal(1))
			Expect(cps[0].RoadID).Should(Equal(int64(405790958)))
			Expect(len(cps[0].Vertex)).Should(Equal(2))
			ids := []int64{cps[0].Vertex[0].Id, cps[0].Vertex[1].Id}
			Expect(ids).Should(ContainElement(int64(3425116687)))
			Expect(ids).Should(ContainElement(int64(3425116689)))
			Expect(cps[0].Lat).To(BeNumerically("~", 31.317221, 0.00001))
			Expect(cps[0].Lon).To(BeNumerically("~", 121.129740, 0.00001))
			Expect(cps[0].Distance).To(BeNumerically("~", 33, 1))
			Expect(cps[0].Ep).To(BeNumerically("~", 0.99, 0.01))
			Expect(cps[0].Ttype).Should(Equal(0))
		})
	})
})
