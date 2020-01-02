package install

import (
	"bosh-vmrun-cpi/config"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func cpiSourcePath(cpiConfig config.Config) string {
	hypervisorPlatform := cpiConfig.Cloud.Properties.Vmrun.Ssh_Tunnel.Platform
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return localPathJoinFunc()(dir, fmt.Sprintf("cpi-%s", hypervisorPlatform))
}

func cpiDestPath(cpiConfig config.Config) string {
	vmStorePath := cpiConfig.Cloud.Properties.Vmrun.Vm_Store_Path
	hypervisorPlatform := cpiConfig.Cloud.Properties.Vmrun.Ssh_Tunnel.Platform
	remotePathJoin := remotePathJoinFunc(cpiConfig)

	if hypervisorPlatform == "windows" {
		return remotePathJoin(vmStorePath, "cpi.exe")
	}

	return remotePathJoin(vmStorePath, "cpi")
}

func stemcellStorePath(cpiConfig config.Config) string {
	return cpiConfig.Cloud.Properties.Vmrun.Stemcell_Store_Path
}

func stemcellMappingsPath(cpiConfig config.Config) string {
	remotePathJoin := remotePathJoinFunc(cpiConfig)
	return remotePathJoin(cpiConfig.Cloud.Properties.Vmrun.Stemcell_Store_Path, "mappings")
}

func localPathJoinFunc() func(...string) string {
	return filepath.Join
}

func remotePathJoinFunc(cpiConfig config.Config) func(...string) string {
	sep := cpiConfig.Cloud.Properties.Vmrun.PlatformPathSeparator()

	return func(pathParts ...string) string {
		return strings.Join(pathParts, sep)
	}
}

func sshCredentials(cpiConfig config.Config) (sshHostname, sshPort, sshUsername, sshRawPrivateKey string) {
	sshHostname = cpiConfig.Cloud.Properties.Vmrun.Ssh_Tunnel.Host
	sshPort = cpiConfig.Cloud.Properties.Vmrun.Ssh_Tunnel.Port
	sshUsername = cpiConfig.Cloud.Properties.Vmrun.Ssh_Tunnel.Username
	sshRawPrivateKey = cpiConfig.Cloud.Properties.Vmrun.Ssh_Tunnel.Private_Key
	return sshHostname, sshPort, sshUsername, sshRawPrivateKey
}
