package main

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/PlainDAG/go-PlainDAG/p2p"
	"github.com/PlainDAG/go-PlainDAG/sign"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/spf13/viper"
)

func Gen_keys(port int) {

	privateKey, _, _ := crypto.GenerateKeyPair(0, 2048)
	// create host
	h := p2p.MakeHost(port, privateKey)
	Pubkey := h.ID().Pretty() // pubkey *
	privateKeyString, _ := crypto.MarshalPrivateKey(privateKey)
	Prvkey := crypto.ConfigEncodeKey(privateKeyString) // prikey *
	fmt.Println(Pubkey)
	fmt.Println(Prvkey)
	h.Close()
}

func judgeNodeType(i int, b []int) bool {
	for _, v := range b {
		if i == v {
			return true
		}
	}
	return false
}

func main() {

	viperRead := viper.New()

	// for environment variables
	viperRead.SetEnvPrefix("")
	viperRead.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viperRead.SetEnvKeyReplacer(replacer)

	viperRead.SetConfigName("configs")
	//viperRead.SetConfigName("config_ecs")
	viperRead.AddConfigPath("../benchmark/config")

	err := viperRead.ReadInConfig()
	if err != nil {
		panic(err)
	}
	// shardNum := viperRead.GetInt("shard_num")

	ProcessCount := viperRead.GetInt("processcount")

	// fmt.Println(tsPubkey)
	//idipmap
	idIPMapInterface := viperRead.GetStringMap("host_ip")

	hostIPMap := make(map[int]string, len(idIPMapInterface)*ProcessCount)

	for idString, ipInterface := range idIPMapInterface {
		id, _ := strconv.Atoi(idString)

		for j := 0; j < ProcessCount; j++ {
			if id == 0 {
				hostIPMap[0] = ipInterface.(string)
				break
			}
			hostIPMap[(id-1)*ProcessCount+j+1] = ipInterface.(string)
		}
	}
	// fmt.Println(hostIPMap)
	p2p_port := viperRead.GetInt("begin_port")
	//idipmap
	idPortMap := make(map[int]int, len(hostIPMap))
	for i := 0; i < len(idIPMapInterface); i++ {
		for j := 0; j < ProcessCount; j++ {
			if i == 0 {

				idPortMap[0] = p2p_port
				break
			}
			idPortMap[(i-1)*ProcessCount+j+1] = p2p_port + i*10 + j
		}
	}

	privateKeysRsa := make(map[int]string)
	publicKeysRsa := make(map[int]string)

	for i := 0; i < len(idIPMapInterface); i++ {
		for j := 0; j < ProcessCount; j++ {
			privateKey, _, _ := crypto.GenerateKeyPair(0, 2048)
			if i == 0 {

				privateKeyString, _ := crypto.MarshalPrivateKey(privateKey)
				privateKeysRsa[0] = crypto.ConfigEncodeKey(privateKeyString)
				h := p2p.MakeHost(idPortMap[0], privateKey)
				publicKeysRsa[0] = h.ID().Pretty()
				break
			}
			subScript := (i-1)*ProcessCount + j + 1

			privateKeyString, _ := crypto.MarshalPrivateKey(privateKey)
			privateKeysRsa[subScript] = crypto.ConfigEncodeKey(privateKeyString)
			h := p2p.MakeHost(idPortMap[subScript], privateKey)
			publicKeysRsa[subScript] = h.ID().Pretty()

		}

	}

	bynum := viperRead.GetInt("byzantine")
	bynodes := make([]int, bynum)
	for i := 0; i < bynum; i++ {
		bynodes[i] = 5*i + 1
	}

	batchsize := viperRead.GetInt("batchsize")
	batchtimeout := viperRead.GetInt("batchtimeout")

	TotalNodeNum := (len(idIPMapInterface)-1)*ProcessCount + 1

	numT := (TotalNodeNum-1)/5 + 1
	shares, pubPoly := sign.GenTSKeys(numT, TotalNodeNum)
	// generate config file for each shard

	// fmt.Println(idPubkeymap)
	// fmt.Println(idPortMap)

	for machineid := 0; machineid < len(idIPMapInterface); machineid++ {

		for processid := 0; processid < ProcessCount; processid++ {
			//generate private key and public key
			nodename := fmt.Sprintf("node%d_%d", machineid, processid)
			viperWrite := viper.New()
			viperWrite.SetConfigFile(fmt.Sprintf("../benchmark/config/%s.yaml", nodename))
			var id int
			if machineid == 0 {
				id = 0
			} else {
				id = (machineid-1)*ProcessCount + processid + 1
			}

			shareAsBytes, err := sign.EncodeTSPartialKey(shares[id])
			if err != nil {
				panic("encode the share")
			}

			tsPubKeyAsBytes, err := sign.EncodeTSPublicKey(pubPoly)
			if err != nil {
				panic("encode the share")
			}
			viperWrite.Set("id", id)
			viperWrite.Set("nodename", nodename)
			viperWrite.Set("ip", hostIPMap[id])
			viperWrite.Set("p2p_port", idPortMap[id])
			viperWrite.Set("private_key", privateKeysRsa[id])
			viperWrite.Set("public_key", publicKeysRsa[id])

			// info of all nodes

			viperWrite.Set("id_ip", hostIPMap)
			viperWrite.Set("id_p2p_port", idPortMap)
			viperWrite.Set("id_public_key", publicKeysRsa)

			viperWrite.Set("tsShare", hex.EncodeToString(shareAsBytes))
			viperWrite.Set("tsPubKey", hex.EncodeToString(tsPubKeyAsBytes))
			viperWrite.Set("batchsize", batchsize)
			viperWrite.Set("batchtimeout", batchtimeout)
			viperWrite.Set("byzantine", bynum)
			if judgeNodeType(id, bynodes) {
				viperWrite.Set("nodetype", 1)
			} else {
				viperWrite.Set("nodetype", 0)
			}

			err = viperWrite.WriteConfig()
			if err != nil {
				panic(err)
			}
			if machineid == 0 {
				break
			}

		}
	}
}

// func Gen_config() {
// 	viperRead := viper.New()

// 	// for environment variables
// 	viperRead.SetEnvPrefix("")
// 	viperRead.AutomaticEnv()
// 	replacer := strings.NewReplacer(".", "_")
// 	viperRead.SetEnvKeyReplacer(replacer)

// 	viperRead.SetConfigName("config")
// 	viperRead.AddConfigPath("./config")

// 	err := viperRead.ReadInConfig()
// 	if err != nil {
// 		panic(err)
// 	}
// 	idNameMapInterface := viperRead.GetStringMap("id_name")
// 	nodeNumber := len(idNameMapInterface)
// 	idNameMap := make(map[int]string, nodeNumber)
// 	for idString, nodenameInterface := range idNameMapInterface {
// 		if nodename, ok := nodenameInterface.(string); ok {
// 			id, err := strconv.Atoi(idString)
// 			if err != nil {
// 				panic(err)
// 			}
// 			idNameMap[id] = nodename
// 		} else {
// 			panic("id_name in the config file cannot be decoded correctly")
// 		}
// 	}
// 	fmt.Println("idNameMap: ", idNameMap)

// 	idP2PPortMapInterface := viperRead.GetStringMap("id_p2p_port")
// 	if nodeNumber != len(idP2PPortMapInterface) {
// 		panic("id_p2p_port does not match with id_name")
// 	}
// 	idP2PPortMap := make(map[int]int, nodeNumber)
// 	for idString, portInterface := range idP2PPortMapInterface {
// 		id, err := strconv.Atoi(idString)
// 		if err != nil {
// 			panic(err)
// 		}
// 		if port, ok := portInterface.(int); ok {
// 			idP2PPortMap[id] = port
// 		} else {
// 			panic("id_p2p_port in the config file cannot be decoded correctly")
// 		}
// 	}
// 	fmt.Println("idP2PPortMap: ", idP2PPortMap)

// 	//idipmap
// 	idIPMapInterface := viperRead.GetStringMap("id_ip")
// 	if nodeNumber != len(idIPMapInterface) {
// 		panic("id_ip does not match with id_name")
// 	}
// 	idIPMap := make(map[int]string, nodeNumber)
// 	for idString, ipInterface := range idIPMapInterface {
// 		id, err := strconv.Atoi(idString)
// 		if err != nil {
// 			panic(err)
// 		}
// 		if ip, ok := ipInterface.(string); ok {
// 			idIPMap[id] = ip
// 		} else {
// 			panic("id_ip in the config file cannot be decoded correctly")
// 		}
// 	}

// 	// generate private keys and public keys for each node
// 	// generate config file for each node
// 	idPrvkeyMap := make(map[int][]byte, nodeNumber)

// 	//idPubkeyMapHex := make(map[int]string, nodeNumber)
// 	for id, _ := range idNameMap {
// 		privateKey, _, _ := crypto.GenerateKeyPair(0, 2048)
// 		privateKeyString, _ := crypto.MarshalPrivateKey(privateKey)

// 		//publicKeyString, _ := crypto.MarshalPublicKey(publicKey)

// 		idPrvkeyMap[id] = privateKeyString

// 		//idPubkeyMapHex[id] = crypto.ConfigEncodeKey(publicKeyString)

// 		// }
// 	}
// 	for id, nodename := range idNameMap {
// 		//generate private key and public key

// 		//generate config file
// 		viperWrite := viper.New()
// 		viperWrite.SetConfigFile(fmt.Sprintf("%s.yaml", nodename))
// 		viperWrite.Set("id", id)
// 		viperWrite.Set("nodename", nodename)
// 		viperWrite.Set("privkey_sig", crypto.ConfigEncodeKey(idPrvkeyMap[id]))
// 		// viperWrite.Set("pubkey_sig", idPubkeyMapHex[id])
// 		// viperWrite.Set("id_pubkey_sig", idPubkeyMapHex)
// 		viperWrite.Set("p2p_port", idP2PPortMap[id])
// 		viperWrite.Set("ip", idIPMap[id])
// 		viperWrite.Set("id_name", idNameMap)
// 		viperWrite.Set("id_p2p_port", idP2PPortMap)
// 		viperWrite.Set("id_ip", idIPMap)
// 		viperWrite.Set("node_number", nodeNumber)
// 		//viperWrite.SetConfigName(nodename)
// 		//viperWrite.AddConfigPath("./")
// 		err := viperWrite.WriteConfig()
// 		if err != nil {
// 			panic(err)
// 		}
// 	}

// }
