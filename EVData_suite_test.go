package EVData_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestEVData(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EVData Suite")
}
