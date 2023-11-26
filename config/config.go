package config

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/PlainDAG/go-PlainDAG/sign"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/spf13/viper"
	"go.dedis.ch/kyber/v3/share"
)

type Config struct {
	Ipaddress string
	Port      int
	Id        int

	Nodetype int

	Prvkey      crypto.PrivKey
	Pubkey      crypto.PubKey
	Pubkeyraw   []byte
	IdPubkeymap map[int]string
	// this map references the id.pretty() to id

	IdportMap map[int]int
	IdaddrMap map[int]string // ip

	//the first  map is to store the public key of of each node, the key string is the string(pubkey) field

	StringpubkeyMap map[string]crypto.PubKey
	// the second map is to store the index of each node and reference the public key to id. the key string is the string(pubkey) field
	StringIdMap map[string]int

	TSPubKey     *share.PubPoly
	TSPrvKey     *share.PriShare
	Batchsize    int
	Batchtimeout int
	Byzantine    int
}

func Loadconfig(config_path string) *Config {

	// // find the number index in string
	// var fileindex int
	// for i := 0; i < len(filepath); i++ {
	// 	if filepath[i] >= '0' && filepath[i] <= '9' {

	// 		//convert byte to int
	// 		fileindex, _ = strconv.Atoi(string(filepath[i]))
	// 		break
	// 	}
	// }

	viperRead := viper.New()

	// for environment variables
	viperRead.SetEnvPrefix("")
	viperRead.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viperRead.SetEnvKeyReplacer(replacer)

	viperRead.SetConfigName(config_path)

	viperRead.AddConfigPath("./")

	err := viperRead.ReadInConfig()
	if err != nil {
		panic(err)
	}

	// id_name_map
	nodeid := viperRead.GetInt("id")
	idipMap := make(map[int]string)
	idipMapInterface := viperRead.GetStringMap("id_ip")
	machineNumber := len(idipMapInterface)
	for idString, nodeipInterface := range idipMapInterface {
		if nodeip, ok := nodeipInterface.(string); ok {
			id, err := strconv.Atoi(idString)
			if err != nil {
				panic(err)
			}
			idipMap[id] = nodeip
		} else {
			panic("id_ip in the config file cannot be decoded correctly")
		}
	}
	// id_p2p_port_map
	idP2PPortMapInterface := viperRead.GetStringMap("id_p2p_port")
	idP2PPortMap := make(map[int]int, machineNumber)

	for idString, portInterface := range idP2PPortMapInterface {
		id, err := strconv.Atoi(idString)
		if err != nil {
			panic(err)
		}
		if port, ok := portInterface.(int); ok {
			idP2PPortMap[id] = port

		} else {
			panic("id_p2p_port in the config file cannot be decoded correctly")
		}
	}

	// id_ip_map

	// id_public_key_map
	privkey := viperRead.GetString("private_key")

	pubkeyothersmap := viperRead.GetStringMap("id_public_key")
	// convert the strings obove into bytes
	privkeybytes, err := crypto.ConfigDecodeKey(privkey)
	if err != nil {
		panic(err)
	}

	//fmt.Println(privkey)
	// convert the bytes into private key and public key
	pubkeysmap := make(map[int]string, machineNumber)
	privkeyobj, err := crypto.UnmarshalPrivateKey(privkeybytes)
	if err != nil {
		panic(err)
	}

	// convert the map into map[int]crypto.PubKey

	for idString, pubkeyothersInterface := range pubkeyothersmap {
		if pubkeyothers, ok := pubkeyothersInterface.(string); ok {
			id, err := strconv.Atoi(idString)
			if err != nil {
				panic(err)
			}

			pubkeysmap[id] = pubkeyothers
		} else {
			panic("public_key_others in the config file cannot be decoded correctly")
		}
	}
	pubkeyidmap := make(map[string]int, machineNumber)
	for id, pubkeyothers := range pubkeysmap {
		pubkeyidmap[pubkeyothers] = id
	}

	// tspubkey
	tsPubKeyAsString := viperRead.GetString("tspubkey")
	tsPubKeyAsBytes, err := hex.DecodeString(tsPubKeyAsString)
	if err != nil {
		panic(err)
	}
	tsPubKey, err := sign.DecodeTSPublicKey(tsPubKeyAsBytes)
	if err != nil {
		panic(err)
	}
	// tsshare
	tsShareAsString := viperRead.GetString("tsshare")
	// fmt.Println("tsShareAsString: ", tsShareAsString)
	tsShareAsBytes, err := hex.DecodeString(tsShareAsString)
	if err != nil {
		panic(err)
	}
	tsShareKey, err := sign.DecodeTSPartialKey(tsShareAsBytes)
	if err != nil {
		panic(err)
	}

	// simlatency

	// shardNum
	batchsize := viperRead.GetInt("batchsize")
	batchtimeout := viperRead.GetInt("batchtimeout")
	fmt.Println(viperRead.GetInt("byzantine"))
	return &Config{

		Id:        nodeid,
		Nodetype:  viperRead.GetInt("nodetype"),
		Byzantine: viperRead.GetInt("byzantine"),

		Ipaddress: idipMap[nodeid],
		Port:      idP2PPortMap[nodeid],

		Prvkey: privkeyobj,

		IdaddrMap: idipMap,
		IdportMap: idP2PPortMap,

		IdPubkeymap: pubkeysmap,
		StringIdMap: pubkeyidmap,

		TSPubKey:     tsPubKey,
		TSPrvKey:     tsShareKey,
		Batchsize:    batchsize,
		Batchtimeout: batchtimeout,
	}
}
