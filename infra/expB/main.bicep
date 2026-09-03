// Experiment B: Azure Database for MySQL Flexible Server — Premium SSD v1 vs v2 pair.
// v2 storage (PremiumV2_LRS) requires the preview API (2025-12-01-preview) and v6 compute
// (platform-enforced: UnsupportedStorageSkuForV6VmSku — v6 only with SSDv2, and SSDv2 needs v6).
// Buffer pool is pinned small (8 GiB) so the ~24 GiB dataset exceeds cache (validity gate G1).
targetScope = 'resourceGroup'

param location string = 'koreacentral'
param namePrefix string = 'mysqlbm2'
param administratorLogin string = 'mysqladmin'
@secure()
param administratorLoginPassword string
param v1SkuName string = 'Standard_E8ds_v5'   // C5: E16ds_v5
param v2SkuName string = 'Standard_E8ds_v6'   // C5: E16ds_v6
param storageSizeGB int = 128
param iops int = 5000                          // same-IOPS cell; at 128GiB v1 free IOPS is ~5037 anyway
param highAvailabilityMode string = 'Disabled' // C1/C5: 'SameZone' (requires autoGrow Enabled)
param availabilityZone string = '1'
param bufferPoolBytes string = '8589934592'
param benchVmSize string = 'Standard_D16ds_v5'
param benchAdminSshKey string
param operatorIp string
param tags object = { workload: 'expB-mysql-ssd-v1v2', owner: 'euson' }

var servers = [
  { name: '${namePrefix}-v1', sku: v1SkuName, storageSku: 'Premium_LRS' }
  { name: '${namePrefix}-v2', sku: v2SkuName, storageSku: 'PremiumV2_LRS' }
]

// ---- network + bench VM -----------------------------------------------------
resource nsg 'Microsoft.Network/networkSecurityGroups@2024-01-01' = {
  name: '${namePrefix}-nsg'
  location: location
  tags: tags
  properties: {
    securityRules: [
      { name: 'ssh-operator', properties: { priority: 100, direction: 'Inbound', access: 'Allow', protocol: 'Tcp', sourceAddressPrefix: '${operatorIp}/32', sourcePortRange: '*', destinationAddressPrefix: '*', destinationPortRange: '22' } }
    ]
  }
}
resource vnet 'Microsoft.Network/virtualNetworks@2024-01-01' = {
  name: '${namePrefix}-vnet'
  location: location
  tags: tags
  properties: {
    addressSpace: { addressPrefixes: ['10.40.0.0/16'] }
    subnets: [ { name: 'bench', properties: { addressPrefix: '10.40.1.0/24', networkSecurityGroup: { id: nsg.id } } } ]
  }
}
resource pip 'Microsoft.Network/publicIPAddresses@2024-01-01' = {
  name: '${namePrefix}-bench-pip'
  location: location
  tags: tags
  sku: { name: 'Standard' }
  properties: { publicIPAllocationMethod: 'Static' }
}
resource nic 'Microsoft.Network/networkInterfaces@2024-01-01' = {
  name: '${namePrefix}-bench-nic'
  location: location
  tags: tags
  properties: {
    enableAcceleratedNetworking: true
    ipConfigurations: [ { name: 'ipconfig1', properties: { subnet: { id: vnet.properties.subnets[0].id }, publicIPAddress: { id: pip.id }, privateIPAllocationMethod: 'Dynamic' } } ]
  }
}
resource vm 'Microsoft.Compute/virtualMachines@2024-07-01' = {
  name: '${namePrefix}-bench-vm'
  location: location
  tags: tags
  zones: [availabilityZone]
  properties: {
    hardwareProfile: { vmSize: benchVmSize }
    osProfile: {
      computerName: '${namePrefix}-bench'
      adminUsername: 'benchadmin'
      linuxConfiguration: { disablePasswordAuthentication: true, ssh: { publicKeys: [ { path: '/home/benchadmin/.ssh/authorized_keys', keyData: benchAdminSshKey } ] } }
      customData: base64(join([
        '#cloud-config'
        'write_files:'
        '  - path: /home/benchadmin/.bench/mysql.env'
        '    permissions: "0600"'
        '    content: |'
        '      MYSQL_USER=${administratorLogin}'
        '      MYSQL_PASSWORD=${administratorLoginPassword}'
        '      V1_HOST=${namePrefix}-v1.mysql.database.azure.com'
        '      V2_HOST=${namePrefix}-v2.mysql.database.azure.com'
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

// ---- MySQL pair (preview API so storageSku is writable for the v2 arm) ------
#disable-next-line BCP081
resource mysql 'Microsoft.DBforMySQL/flexibleServers@2025-12-01-preview' = [for s in servers: {
  name: s.name
  location: location
  tags: union(tags, { storageSku: s.storageSku })
  sku: { name: s.sku, tier: 'MemoryOptimized' }
  properties: {
    createMode: 'Default'
    version: '8.4'
    administratorLogin: administratorLogin
    administratorLoginPassword: administratorLoginPassword
    availabilityZone: availabilityZone
    backup: { backupRetentionDays: 7, geoRedundantBackup: 'Disabled' }
    highAvailability: { mode: highAvailabilityMode, standbyAvailabilityZone: highAvailabilityMode == 'SameZone' ? availabilityZone : null }
    network: { publicNetworkAccess: 'Enabled' }
    storage: {
      storageSizeGB: storageSizeGB
      iops: iops
      autoGrow: highAvailabilityMode == 'Disabled' ? 'Disabled' : 'Enabled' // HA requires autoGrow; beware unbounded growth under sustained writes (observed 128GiB→1.6TB in a 14h soak)
      autoIoScaling: 'Disabled'
      logOnDisk: 'Disabled'
      storageRedundancy: 'LocalRedundancy'
      storageSku: s.storageSku
    }
    dataEncryption: { type: 'SystemManaged' }
  }
}]

#disable-next-line BCP081
resource dbs 'Microsoft.DBforMySQL/flexibleServers/databases@2025-12-01-preview' = [for (s, i) in servers: {
  parent: mysql[i]
  name: 'benchmark'
  properties: { charset: 'utf8mb4', collation: 'utf8mb4_0900_ai_ci' }
}]

#disable-next-line BCP081
resource bufferPool 'Microsoft.DBforMySQL/flexibleServers/configurations@2025-12-01-preview' = [for (s, i) in servers: {
  parent: mysql[i]
  name: 'innodb_buffer_pool_size'
  properties: { value: bufferPoolBytes, source: 'user-override' }
  dependsOn: [dbs[i]]
}]

#disable-next-line BCP081
resource fw 'Microsoft.DBforMySQL/flexibleServers/firewallRules@2025-12-01-preview' = [for (s, i) in servers: {
  parent: mysql[i]
  name: 'bench-vm'
  properties: { startIpAddress: pip.properties.ipAddress, endIpAddress: pip.properties.ipAddress }
  dependsOn: [bufferPool[i]]
}]

output v1Fqdn string = mysql[0].properties.fullyQualifiedDomainName
output v2Fqdn string = mysql[1].properties.fullyQualifiedDomainName
output benchIp string = pip.properties.ipAddress
