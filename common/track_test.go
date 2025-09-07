package common_test

import (
	"EVdata/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Track", func() {
	Context("CanProjectOntoSegment FootPointAndDistance", func() {
		//A(31.317221,121.129546) B(31.317521, 121.1295856 ) C(31.3170233，121.1298425)
		It("should get right result #1", MustPassRepeatedly(1), func() {
			latB, lonB, latC, lonC, latA, lonA := 31.317521, 121.1295856, 31.3170233, 121.1298425, 31.317100, 121.129546
			onSeg, t := common.CanProjectOntoSegment(latA, lonA, latB, lonB, latC, lonC)
			footLat, footLon, distMeters := common.FootPointAndDistance(latA, lonA, latB, lonB, latC, lonC, t)

			Expect(t).Should(BeNumerically("~", 0.6817461, 1e-7))
			Expect(onSeg).Should(BeTrue())
			Expect(footLat).Should(BeNumerically("~", 31.3171817, 1e-7))
			Expect(footLon).Should(BeNumerically("~", 121.1297607, 1e-7))
			Expect(distMeters).Should(BeNumerically("~", 22.33, 1e-2))
		})
	})

	Context("Distance", func() {
		//A(31.317221,121.129546) B(31.317521, 121.1295856 ) C(31.3170233，121.1298425)
		It("Distance should get right result #1", MustPassRepeatedly(1), func() {
			x1, y1 := 31.317100, 121.129546
			x2, y2 := 31.317181695, 121.129760741
			x3, y3 := 31.3172047, 121.1297489
			dis1 := common.Distance(x1, y1, x2, y2)
			dis2 := common.Distance(x1, y1, x3, y3)

			Expect(dis1).Should(BeNumerically("~", 22.33, 1e-2))
			Expect(dis2).Should(BeNumerically("~", 22.51, 1e-2))
		})
	})
})
