// Experiment A: Azure Database for PostgreSQL Flexible Server vs Azure HorizonDB.
// Case C1 = 8 vCore / Same-zone HA / no read replica (HorizonDB mapped to vCores=8, replicaCount=1 = one standby-capable replica; documented as best-effort mapping).
// Case C5 = 16 vCore (scale in place: `az postgres flexible-server update --sku-name Standard_D16ds_v5`, HorizonDB vCores 16).
targetScope = 'resourceGroup'

param location string = 'australiaeast'
param prefix string = 'expa'
param administratorLogin string = 'benchadmin'
@secure()
param administratorLoginPassword string
param vCores int = 8
param pgSkuName string = 'Standard_D8ds_v5'
param pgHaMode string = 'SameZone'          // 'Disabled' | 'SameZone' | 'ZoneRedundant'
param horizonReplicaCount int = 1
param benchVmSize string = 'Standard_D16ds_v5'
param benchAdminSshKey string
param operatorIp string
param tags object = { workload: 'expA-pg-vs-horizondb', owner: 'euson' }

var pgName = '${prefix}-pg'
var hzName = '${prefix}-hz'

resource vnet 'Microsoft.Network/virtualNetworks@2024-01-01' = {
  name: '${prefix}-vnet'
  location: location
  tags: tags
  properties: {
    addressSpace: { addressPrefixes: ['10.30.0.0/16'] }
    subnets: [ { name: 'bench', properties: { addressPrefix: '10.30.1.0/24', networkSecurityGroup: { id: nsg.id } } } ]
  }
}
resource nsg 'Microsoft.Network/networkSecurityGroups@2024-01-01' = {
  name: '${prefix}-nsg'
  location: location
  tags: tags
  properties: {
    securityRules: [
      { name: 'ssh-operator', properties: { priority: 100, direction: 'Inbound', access: 'Allow', protocol: 'Tcp', sourceAddressPrefix: '${operatorIp}/32', sourcePortRange: '*', destinationAddressPrefix: '*', destinationPortRange: '22' } }
    ]
  }
}
resource pip 'Microsoft.Network/publicIPAddresses@2024-01-01' = {
  name: '${prefix}-bench-pip'
  location: location
  tags: tags
  sku: { name: 'Standard' }
  properties: { publicIPAllocationMethod: 'Static' }
}
resource nic 'Microsoft.Network/networkInterfaces@2024-01-01' = {
  name: '${prefix}-bench-nic'
  location: location
  tags: tags
  properties: {
    enableAcceleratedNetworking: true
    ipConfigurations: [ { name: 'ipconfig1', properties: { subnet: { id: vnet.properties.subnets[0].id }, publicIPAddress: { id: pip.id }, privateIPAllocationMethod: 'Dynamic' } } ]
  }
}
resource vm 'Microsoft.Compute/virtualMachines@2024-07-01' = {
  name: '${prefix}-bench-vm'
  location: location
  tags: tags
  zones: ['1']
  properties: {
    hardwareProfile: { vmSize: benchVmSize }
    osProfile: {
      computerName: '${prefix}-bench'
      adminUsername: 'benchadmin'
      linuxConfiguration: { disablePasswordAuthentication: true, ssh: { publicKeys: [ { path: '/home/benchadmin/.ssh/authorized_keys', keyData: benchAdminSshKey } ] } }
      // The DB password is delivered ONLY to the bench VM (cloud-init), never to the operator.
      customData: base64(join([
        '#cloud-config'
        'write_files:'
        '  - path: /home/benchadmin/.bench/pg.env'
        '    permissions: "0600"'
        '    content: |'
        '      PG_USER=${administratorLogin}'
        '      PG_PASSWORD=${administratorLoginPassword}'
        '      PG_HOST=${pgName}.postgres.database.azure.com'
        'runcmd:'
        '  - chown -R benchadmin:benchadmin /home/benchadmin/.bench'
      ], '\n'))
    }
    storageProfile: {
      imageReference: { publisher: 'Canonical', offer: 'ubuntu-24_04-lts', sku: 'server', version: 'latest' }
      osDisk: { createOption: 'FromImage', diskSizeGB: 64, managedDisk: { storageAccountType: 'Premium_LRS' } }
    }
    networkProfile: { networkInterfaces: [ { id: nic.id } ] }
  }
}

resource postgres 'Microsoft.DBforPostgreSQL/flexibleServers@2024-08-01' = {
  name: pgName
  location: location
  tags: tags
  sku: { name: pgSkuName, tier: 'GeneralPurpose' }
  properties: {
    version: '17'
    administratorLogin: administratorLogin
    administratorLoginPassword: administratorLoginPassword
    availabilityZone: '1'
    backup: { backupRetentionDays: 7, geoRedundantBackup: 'Disabled' }
    highAvailability: pgHaMode == 'Disabled' ? { mode: 'Disabled' } : { mode: pgHaMode, standbyAvailabilityZone: pgHaMode == 'ZoneRedundant' ? '2' : null }
    network: { publicNetworkAccess: 'Enabled' }
    storage: { storageSizeGB: 256, autoGrow: 'Enabled', tier: 'P15' }
  }
}
resource pgFw 'Microsoft.DBforPostgreSQL/flexibleServers/firewallRules@2024-08-01' = {
  name: 'bench-vm'
  parent: postgres
  properties: { startIpAddress: pip.properties.ipAddress, endIpAddress: pip.properties.ipAddress }
}
resource pgDb 'Microsoft.DBforPostgreSQL/flexibleServers/databases@2024-08-01' = {
  name: 'benchmark'
  parent: postgres
  properties: { charset: 'UTF8', collation: 'en_US.utf8' }
}

#disable-next-line BCP081
resource horizon 'Microsoft.HorizonDB/clusters@2026-01-20-preview' = {
  name: hzName
  location: location
  tags: tags
  properties: {
    createMode: 'Create'
    version: '17'
    administratorLogin: administratorLogin
    administratorLoginPassword: administratorLoginPassword
    vCores: vCores
    replicaCount: horizonReplicaCount
    zonePlacementPolicy: 'BestEffort'
  }
}
#disable-next-line BCP081
resource hzFw 'Microsoft.HorizonDB/clusters/pools/firewallRules@2026-01-20-preview' = {
  name: '${horizon.name}/DefaultPool/bench-vm'
  properties: { startIpAddress: pip.properties.ipAddress, endIpAddress: pip.properties.ipAddress }
}

output pgFqdn string = postgres.properties.fullyQualifiedDomainName
output hzFqdn string = horizon.properties.fullyQualifiedDomainName
output benchIp string = pip.properties.ipAddress
output pgId string = postgres.id
output hzId string = horizon.id
