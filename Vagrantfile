# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
 
  config.vm.hostname = "stocker"
  config.vm.box = "baseline-precise"
  config.vm.box_url = "https://s3.amazonaws.com/int-ops/baseline/vagrant/virtualbox/precise.box"

  config.berkshelf.berksfile_path = "./Berksfile"
  config.berkshelf.enabled = true

  config.vm.provision :chef_solo do |chef|
    chef.json = {
      redisio: {
        servers: [
          {
            port: '6379'
          }
        ]
      },
      go: {
        version: '1.2',
        packages: [
          'github.com/garyburd/redigo/redis',
          'github.com/dotcloud/docker'
        ]
      }
    }

    chef.run_list = [
      'recipe[baseline]',
      'recipe[redisio::install]',
      'recipe[redisio::enable]',
      'recipe[sqlite3]',
      'recipe[lvm2]',
      'recipe[golang::packages]'
    ]
  end
end
