package dnspod

import (
	_ "cert/configs"

	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/dnspod"
	"github.com/go-acme/lego/v4/registration"
	"github.com/spf13/viper"
)

// You'll need a user or account type that implements acme.User
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func Dnspod_cert(account string, domain_name string) {

	// Create a user. New accounts need an email and private key to start.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	myUser := MyUser{
		Email: viper.GetString("user_email"),
		key:   privateKey,
	}

	config := lego.NewConfig(&myUser)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	config.CADirURL = viper.GetString("letsencrypt_url")
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// We specify an HTTP port of 5002 and an TLS port of 5001 on all interfaces
	// because we aren't running as root and can't bind a listener to port 80 and 443
	// (used later when we attempt to pass challenges). Keep in mind that you still
	// need to proxy challenge traffic to port 5002 and 5001.

	// Config is used to configure the creation of the DNSProvider.
	dnspodconfig := dnspod.Config{
		LoginToken:         viper.GetString("DNSPOD." + account + ".key"),
		TTL:                600,
		PropagationTimeout: dns01.DefaultPropagationTimeout,
		PollingInterval:    dns01.DefaultPollingInterval,
		HTTPClient:         http.DefaultClient,
	}

	p, err := dnspod.NewDNSProviderConfig(&dnspodconfig)
	err = client.Challenge.SetDNS01Provider(p)
	if err != nil {
		log.Fatal(err)
	}

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
	}
	myUser.Registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{domain_name, "*." + domain_name},
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}

	// Each certificate comes back with the cert bytes, the bytes of the client's
	// private key, and a certificate URL. SAVE THESE TO DISK.
	fmt.Printf("%#v\n", certificates)
	err = os.MkdirAll("./cert/"+request.Domains[0], os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	f, _ := os.Create("./cert/" + request.Domains[0] + "/" + request.Domains[0] + ".key")
	f.Write(certificates.PrivateKey)
	f.Close()

	f1, _ := os.Create("./cert/" + request.Domains[0] + "/" + request.Domains[0] + ".crt")
	f1.Write(certificates.Certificate)
	f1.Close()
	// ... all done.
}
