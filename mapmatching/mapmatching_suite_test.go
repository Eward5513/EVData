package mapmatching_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMapmatching(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mapmatching Suite")
}

var _ = BeforeSuite(func() {
	// 这里写全局初始化逻辑
	println("BeforeSuite: 全局初始化一次")
})
