include_recipe 'build-essential'
include_recipe 'git'

bash 'build-lvm2' do
  cwd '/usr/local/lvm2'
  code <<-EOF
    ./configure --enable-static_link
    make device-mapper
    make install_device-mapper
  EOF
  action :nothing
end

git 'sync-lvm2' do
  destination '/usr/local/lvm2'
  repository 'https://git.fedorahosted.org/git/lvm2.git'
  reference 'v2_02_103'
  action :sync
  notifies :run, 'bash[build-lvm2]', :immediately
end
