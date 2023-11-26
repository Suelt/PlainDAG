## Description
An implementation for PlainDAG - The first non-RBC DAG-based BFT protocol.


## Usage
### 1. Machine types
Machines are divided into two types:
- *Workcomputer*: configure `servers` and `clients`  particularly via `ansible` tool 
- *Servers*: run daemons of `PlainDAG`, communicate with each other via P2P model
- *Clients*: run `client`, communicate with `server` via RPC model 

### 2. Environment requirements
- Recommended OS releases: Ubuntu 20.04
- Go version: 1.19
- Python version: 3.9.4
- Ansible version: 2.5.1

### 3. Steps to run PlainDAG

#### 3.1 Install ansible on the work computer
Commands below are run on the *work computer*.
```shell script
sudo apt install python3-pip
sudo pip3 install --upgrade pip
pip3 install ansible
# add ~/.local/bin to your $PATH
echo 'export PATH=$PATH:~/.local/bin:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### 3.2 Login without passwords
Enable workcomputer to login in servers and clients without passwords.

#### 3.3 Install Go-related modules/packages

install go modules/packages on the *work computer*.


#### 3.4 Generate configurations

Generate configurations for each server.

Operations below are done on the *work computer*.

- Change `ips`, `port`,  and other parameters  in file `benchmark/configs.yaml`
-  Run `go run ./config_gen/main.go`

#### 3.5 Configure servers via ansible tool
Change the `hosts` file in the directory `ansible-remote`, the hostnames and IPs should be consistent with `benchmark/configs.yaml`.

And commands below are run on the *work computer*.

```shell script
# Enter the directory `PlainDAG`
go build
# Enter the directory `ansible`
ansible-playbook -i ./hosts conf-server.yaml
```

#### 3.6 Run PlainDAG servers via ansible tool
```shell script
# run PlainDAG servers
ansible-playbook -i ./hosts run-server.yaml
```

You can stop servers by using command

```shell
# stop PlainDAG servers
ansible-playbook -i ./hosts kill-server.yaml
```

#### 3.7 Run Client

For stable testing, we let the primary node to package a batch by itself if it does not receive any request from the client within the `batchtimeout` period. And you can also run a client to send request.

You can use the workserver as the client or run a client on a remote host. The following command will start a client and the client will keep sending requests to the primary node until you stop it:

```shell
#Enter the directory `client`
go run main.go -rpcaddress $IP_ADDR_OF_Linked sendrequest
```

If you want to send exact number of requests, you can use：

```shell
#Enter the directory `client`
go run main.go -rpcaddress $IP_ADDR_OF_Linked sendrequest -n $number
```

For the full list of flags, run `go run main.go -rpcaddress $IP_ADDR_OF_Leader hlep sendrequest`.



