# hyperledger_on_swarm

This repository is for deploying Hyperledger Fabric on Swarm cluster easily.

## Limitation
- This works WITHOUT TLS only.
  Whenever enable TLS, grpc error code 14 occurs.

### Pre-reqs
- 2 or more machines
- Install Docker >= 1.13

### Get Hyperledger Fabric artifacts and binaries
* As all containers share same cryption keys and artifacts, you need to put them on same location.
   - In this case, the location is '/nfs-share' (I personally use NFS)
   - For example,
   ```
    cd /nfs-share
    curl -sSL https://goo.go/NIKLiU | bash
    ```
    
### Generate the artifacts


## 3. 
TBC.....
