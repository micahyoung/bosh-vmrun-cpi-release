$env:SSH_USERNAME=$env:USERNAME
$env:SSH_HOSTNAME="localhost"
$env:SSH_PORT="22"
Set-Content env:SSH_PRIVATE_KEY -Value (Get-Content -Raw ~\.ssh\id_rsa_vmrun)
$env:SSH_PLATFORM="windows"
$env:PATH+=';C:\Program Files (x86)\VMware\VMware Workstation'
$env:PATH+=';C:\Program Files (x86)\VMware\VMware Workstation\OVFTool'