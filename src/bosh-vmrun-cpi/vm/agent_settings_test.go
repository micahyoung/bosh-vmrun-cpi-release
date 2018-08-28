package vm_test

import (
	"bytes"

	"github.com/cppforlife/bosh-cpi-go/apiv1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakelogger "github.com/cloudfoundry/bosh-utils/logger/loggerfakes"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"

	"bosh-vmrun-cpi/vm"
)

var _ = Describe("AgentSettings", func() {
	var (
		logger          *fakelogger.FakeLogger
		fs              *fakesys.FakeFileSystem
		agentEnvFactory apiv1.AgentEnvFactory
		agentSettings   vm.AgentSettings
	)

	BeforeEach(func() {
		logger = &fakelogger.FakeLogger{}
		fs = fakesys.NewFakeFileSystem()
		agentEnvFactory = apiv1.NewAgentEnvFactory()

		agentSettings = vm.NewAgentSettings(fs, logger, agentEnvFactory)
	})

	Describe("GenerateAgentEnvIso", func() {
		It("returns path to the generated agent config iso", func() {
			agentEnv, err := apiv1.AgentEnvFactory{}.FromBytes([]byte("{}"))

			isoPath, err := agentSettings.GenerateAgentEnvIso(agentEnv)
			Expect(err).ToNot(HaveOccurred())
			Expect(isoPath).To(ContainSubstring("/env.iso"))

			fileStats := fs.GetFileTestStat(isoPath)
			Expect(err).ToNot(HaveOccurred())

			agentEnvBytes, _ := agentEnv.AsBytes()
			Expect(bytes.Contains(fileStats.Content, agentEnvBytes)).To(BeTrue())
			Expect(len(fileStats.Content)).To(Equal(4096))
		})
	})

	Describe("AgentEnvBytesFromFile", func() {
		It("returns AgentEnv from the env iso", func() {
			isoPath := "../test/fixtures/env.iso"
			agentEnv, err := agentSettings.GetIsoAgentEnv(isoPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(agentEnv).ToNot(BeNil())

			// VM VMSpec `json:"vm"`
			//
			// Mbus string   `json:"mbus"`
			// NTP  []string `json:"ntp"`
			//
			// Blobstore BlobstoreSpec `json:"blobstore"`
			//
			// Networks NetworksSpec `json:"networks"`
			//
			// Disks DisksSpec `json:"disks"`
			//
			// Env EnvSpec `json:"env"`

		})

		It("returns error when file does not exist", func() {
			agentEnv, err := agentSettings.GetIsoAgentEnv("does/not/exist")
			Expect(err).To(HaveOccurred())
			Expect(agentEnv).To(BeNil())
		})
	})
})
