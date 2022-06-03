package fidelity_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
)

func TestFidelity(t *testing.T) {
	RegisterFailHandler(Fail)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	RunSpecs(t, "Fidelity Suite")
}
