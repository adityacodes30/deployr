package utils

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
	"golang.org/x/crypto/ssh"
)

func Keygen() (string, string, string) {
	deployrdir := os.Getenv("HOME") + "/.deployr"
	tempKeyDir := deployrdir + "/temp"
	savePrivateFileTo := deployrdir + "/temp/key"
	savePublicFileTo := deployrdir + "/temp/key.pub"
	bitSize := 4096

	err := os.MkdirAll(deployrdir, 0755)
	if err != nil {
		log.Fatal("Failed to create .deployr directory: ", err)
	}

	errr := os.MkdirAll(tempKeyDir, 0755)
	if errr != nil {
		log.Fatal("Failed to create .deployr directory: ", err)
	}

	privateKey, err := generatePrivateKey(bitSize)
	if err != nil {
		log.Fatal(err.Error())
	}

	publicKeyBytes, err := generatePublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	privateKeyBytes := encodePrivateKeyToPEM(privateKey)

	err = writeKeyToFile(privateKeyBytes, savePrivateFileTo)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = writeKeyToFile([]byte(publicKeyBytes), savePublicFileTo)
	if err != nil {
		log.Fatal(err.Error())
	}

	// fmt.Println("public key " + string(publicKeyBytes))
	// fmt.Println("private key " + string(privateKeyBytes))

	return string(publicKeyBytes), string(privateKeyBytes), savePrivateFileTo
}

func generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	log.Println("Private Key generated")
	return privateKey, nil
}

func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}

// returns in the format "ssh-rsa ..."
func generatePublicKey(privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	log.Println("Public key generated")
	return pubKeyBytes, nil
}

func writeKeyToFile(keyBytes []byte, saveFileTo string) error {
	err := ioutil.WriteFile(saveFileTo, keyBytes, 0600)
	if err != nil {
		return err
	}

	log.Printf("Key saved to: %s", saveFileTo)
	return nil
}

func AddHostToKnownHosts(host string, key ssh.PublicKey) error {
	knownHostsPath := os.Getenv("HOME") + "/.ssh/known_hosts"

	hostKeyEntry := fmt.Sprintf("%s %s %s\n",
		host,
		key.Type(),
		base64.StdEncoding.EncodeToString(key.Marshal()))

	return ioutil.WriteFile(knownHostsPath, []byte(hostKeyEntry), 0644)
}

func ResolveConfimation(domain string, ip string) {

	fmt.Print("\n\033[32mPlease type 'confirm' if your IP has propagated:\033[0m\n\n")
	fmt.Printf("You can check that by if the IP on this website matches --> \033[33m%s\033[0m\n \n", ip)
	fmt.Printf("Paste this in your browser --> \033[34m: https://www.nslookup.io/domains/%s/dns-records/\033[0m\n", domain)
	fmt.Printf("\033[34m\033]8;;https://www.nslookup.io/domains/%s/dns-records/\033\\Or Click here to directly go to the website\033]8;;\033\\\033[0m\n\n", domain)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "confirm" {
			fmt.Println("DNS-IP propagation confirmed. Proceeding in...")
			break
		} else {
			fmt.Println("Invalid input. Please type 'confirm'.")
		}
	}

	for i := 10; i >= 1; i-- {

		fmt.Printf("\033[32m\r%d\033[0m\n", i)
		time.Sleep(1 * time.Second)
	}

	fmt.Println() // Move to the next line when done
}

///// -------------- Under Development ------------------

func PollForIpPoint(domain string, ipAddr string) bool {
	counter := 0
	dnsResolver := "8.8.8.8:53"

	for {
		ip, err := resolveWithDNS(domain, dnsResolver)
		if err != nil {
			fmt.Printf("Error looking up host: %v\n", err)
			return false
		}

		if len(ip) > 0 && ip[0] == ipAddr {
			fmt.Println("Domain resolved to the correct IP!")
			return true
		}

		counter++
		if counter >= 1000 {
			fmt.Println("Reached maximum retries. DNS propagation failed.")
			return false
		}

		fmt.Println("Waiting for public DNS to be propagated...")
		time.Sleep(20 * time.Second)
	}
}

func resolveWithDNS(domain string, resolver string) ([]string, error) {
	c := dns.Client{}
	m := dns.Msg{}
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, _, err := c.ExchangeContext(ctx, &m, resolver)
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, ans := range resp.Answer {
		if a, ok := ans.(*dns.A); ok {
			ips = append(ips, a.A.String())
		}
	}
	return ips, nil
}

func PrintSucesss(domain string) {

	fmt.Printf("\033[32m"+`
╭──────────────────────────────────────────────────────────╮
│                   Congratulations! 🎉                     │
│                                                          │
│ Your app is deployed on https://%s                       │
│                                                          │
│ Do not worry if you do not see your application up      │
│ instantly, it might be building. It also takes some time │
│ for DNS and SSL certificates to propagate across the     │
│ global network. Please be patient.                       │
│                                                          │
│ If you do not see it deployed even after a few hours,   │
│ please paste the logs you got above at:                 │
│ https://github.com/adityacodes30/deployr/issues         │
╰──────────────────────────────────────────────────────────╯
`+"\033[0m", domain)

}
