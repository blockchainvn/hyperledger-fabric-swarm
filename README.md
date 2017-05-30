# hyperledger_on_swarm
---

I'm going to explain how to deploy hyperledger fabric on Docker swarm in this respository.
## 0. Pre-reqs
 - 3 physical or virtual machines
   : 1 NFS server + 2 NFS client
   https://www.howtoforge.com/tutorial/setting-up-an-nfs-server-and-client-on-centos-7/
     NFS is mounted on /nfs-share
 - Download and Extract artifacts and binaries
   cd /nfs-share
   curl -sSL https://goo.gl/NIKLiU | bash
 - port already used....
   https://github.com/moby/moby/issues/31249

## 1. Install Docker >= 1.13

https://docs.docker.com/cs-engine/1.13/#install-on-centos-7172--rhel-70717273-yum-based-systems

systemctl start docker && systemctl enable docker

## 2. Generate the artifacts

## 3. 
TBC.....
