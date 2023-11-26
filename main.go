package main

import (
	"flag"
	"time"

	"github.com/PlainDAG/go-PlainDAG/core"
)

func main() {
	optype := flag.Int("o", 1, "specify the operation type")
	// cycle := flag.Int("c", 500, "specify the number of tx sending cycles")
	// rate := flag.Int("r", 2000, "specify the sending rate per round")

	config_path := flag.String("f", "config", "config path")

	// nodeid := flag.Int("n", 0, "config nodeid")
	flag.Parse()

	switch *optype {

	case 1:
		n, err := core.StartandConnect(*config_path)
		if err != nil {
			panic(err)
		}

		time.Sleep(5 * time.Second)
		go n.SendForever()
		go n.HandleMsgForever()

		select {}

	}
}

// shares, pubPoly := sign.GenTSKeys(2, 6)
// for i, share := range shares {
// 	viperWrite := viper.New()
// 	viperWrite.SetConfigFile(fmt.Sprintf("config%d.yaml", i))
// 	shareAsBytes, err := sign.EncodeTSPartialKey(share)
// 	if err != nil {
// 		panic(err)
// 	}
// 	tsPubKeyAsBytes, err := sign.EncodeTSPublicKey(pubPoly)
// 	if err != nil {
// 		panic(err)
// 	}
// 	viperWrite.Set("TSShare", hex.EncodeToString(shareAsBytes))
// 	viperWrite.Set("TSPubKey", hex.EncodeToString(tsPubKeyAsBytes))

// }
// mbr.Msgs[1] = append(mbr.Msgs[1], &messagesbyindex[0])
// prvkeystring := "CAASpwkwggSjAgEAAoIBAQCg40gvh0q9OwNOC31LoIGnqjFh9tzsLbUGBaPeVLfb1BqB7EWt5Ya8M3yK3CIvFhFdio6IBRFu6jCg0rFcK4FOc8qlw//SBwPgeyW56FOfjKI/WTIe6FR1O0EFjSqE5Oubiy1RvFVdmHuyJeekkFQLFknUntNucZKbn3gDNOOUR6eV82PR6Q39ttyQpX+hMzBzzv7K/isoqpcwg9CAfPZGJz8AfgGBfVgVo9yxer/6zghBjzdf1QlH2jdgotGLmaIrnj3sVai2gC5PNUqeg3Imd8Ow9ftbG35QZVET/QEKbrnaDu9SxaJ9LEJHiMh0hZsQKqco4IJEFMka0Q5XtvZZAgMBAAECggEAbxM7PQUMxoQ9jd5EzLeti9Hmchn7AFu2BMhUECUxImXXPyeeG6bBVKG/NCcyuotjxc7pBGNrW8X3eLC9nkKy7TToDXW54ojRVmPu8eDCCv8O7OlpvwjrdlxIUcraNhHNz/9QdIOv9ARYMfAVcvnp7BWhN1fH5RIoA6UfOCeFj1Ko1maiHqzX0q2zItz/dXTIISLk//W7Bi5oOQd8YehNsX/CwIdDrRgcdj9sscPb+4ZtuRuPUZ1n86E3q5rn8g1Qy3tXhIj+OM7N2eJYvVEDysgvPkuTcKHa8pjEtIzqPjtMwy5NGLPZb7P8xf033oPq4q99A7hf27fc9XrRWOatmQKBgQDMdxzBslAqoXL6YWGsQ+XMQP2E0INmorSe58/WPfAWIym+gm886HnUAfAT57sydS4WpBINGatIecu8DW2yFsuixKyiDS/DmpmiZllnHHXSnhhEPkNNbaq0ikERvBkJCzDBJUpiZC7y+5534oozbJC2vCtxfs6/76ttThXIGtFT/wKBgQDJcGJHs6v/deX/FRFdu6FkTQJCNXFTGrY2tK5hKLZREhZckA+sDSw+kpO1rhmvfu9LlzZwMEBafeke26bOHK402qApnZSphT73IqXxM8KAeL7yjQOfqC4Jt7oKSW5QCioeJdSVZuTjm2n7/gfuDg6dyucHeiwsPwarLF0WWbrVpwKBgHj+a5f76xCvJZkhE2mbbFsogl2b/oY39maqiwUe9fpDqKpCCY2jjKR22RkOYmqDiViAkuYJsKBc4sFPuQBQGQUjGX10DDXWQOAnbPRllRujznxj0/P317KqtcLG6pG9e4ZwiMocuuOzHp2XA21W63QXeiXZgoN2Up2GPcGCjSkFAoGBAML7+tXm7+1GZQvli7rMXSumczI9YuWLbKdFe6nma5vLw0Nz2weydIpY/YuV65z5ZI4p33L28cPmLtpEyZCnKGVW3kOKGhWBOfKkYjY44OPUfRhxMPnBJFcZtYYxkAr/28b03XKEd7htfkiCm5BtoO5SMhEFzG5Dz6OvPKfe0T/vAoGATx/uRkPxsVR1GkXFb6Q8e9HT5frMIhULxMoTSc7T8sF/HwGjzCSp18Xdx9X9GQAmSVyGUfIg2om82FHq8a9cNc79kdZXJKUy657Ts729xXeqenuGk+NN/N6BnykKIODP1eEqMGCmmE6z9YHufHObud2MR3dS81L4ilIYS/b6CV8="
// privkeybytes, _ := crypto.ConfigDecodeKey(prvkeystring)
// privkeyobj, _ := crypto.UnmarshalPrivateKey(privkeybytes)
// h := p2p.MakeHost(9005, privkeyobj)
// fmt.Pntln(h.ID().Pretty())
