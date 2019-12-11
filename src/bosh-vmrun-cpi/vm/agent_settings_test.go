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
			agentEnv := &apiv1.AgentEnvImpl{}

			isoPath, err := agentSettings.GenerateAgentEnvIso(agentEnv)
			Expect(err).ToNot(HaveOccurred())
			Expect(isoPath).To(ContainSubstring("env.iso"))

			fileStats := fs.GetFileTestStat(isoPath)
			Expect(err).ToNot(HaveOccurred())

			expectedEnvBytes, _ := agentEnv.AsBytes()
			Expect(bytes.Contains(fileStats.Content, expectedEnvBytes)).To(BeTrue())
			Expect(len(fileStats.Content)).To(Equal(4096))
		})
	})

	Describe("AgentEnvBytesFromFile", func() {
		It("returns AgentEnv from the env iso", func() {
			isoPath := "../test/fixtures/env.iso"
			actualAgentEnv, err := agentSettings.GetIsoAgentEnv(isoPath)
			Expect(err).ToNot(HaveOccurred())

			expectedAgentEnv, _ := agentEnvFactory.FromBytes([]byte(`
				{
					"agent_id":"c292e270-d067-45a9-553d-645c9c7c4fd9",
					"vm":{"name":"vm-d70f1f0d-b9b8-4ad9-b9c0-9cd45853349f","id":"vm-54443"},
					"mbus":"https://mbus:mbus-password@0.0.0.0:6868",
					"ntp":["0.pool.ntp.org","1.pool.ntp.org"],
					"blobstore":{"provider":"local","options":{"blobstore_path":"/var/vcap/micro_bosh/data/cache"}},
					"networks":{
						"private":{"type":"manual","ip":"10.85.57.200","netmask":"255.255.255.0","gateway":"10.85.57.1","dns":["8.8.8.8"],"default":["dns","gateway"],"mac":"00:50:56:9a:20:b2","preconfigured":false}
					},
					"disks":{"system":"0","ephemeral":"1","persistent":{"disk-773d31f6-5245-468f-a324-8698b07330a8":"2"}},
					"env":{}
				}
			`))
			Expect(actualAgentEnv).To(Equal(expectedAgentEnv))
	
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
