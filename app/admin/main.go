package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"service2/business/data/schema"
	"service2/foundations/database"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
)

func main() {
	// keyGen()\
	// fmt.Println("test")
	// tokenGen()
	migrate()
}

func migrate() {

	dbConfig := database.Config{
		User:       "admin",
		Name:       "postgres",
		Host:       "127.0.0.0:5432",
		Password:   "admin",
		DisableTLS: true,
	}
	db, err := database.Open(dbConfig)

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := schema.Migrate(db); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("migration complet")

	if err := schema.Seed(db); err != nil {
		log.Fatalln(err)
	}

}
func keyGen() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalln(err)
	}

	privateFile, err := os.Create("private.pem")

	if err != nil {
		log.Fatalln(err)

	}
	defer privateFile.Close()

	privateBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	if err := pem.Encode(privateFile, &privateBlock); err != nil {
		log.Fatalln(err)
	}

	// public files
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)

	if err != nil {
		log.Println(err)
	}

	publicFile, err := os.Create("public.pem")
	if err != nil {
		return errors.Wrap(err, "creating public file")
	}
	defer privateFile.Close()

	// construct a PEM block for public key.
	publicBlock := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	if err := pem.Encode(publicFile, &publicBlock); err != nil {
		log.Fatalln(err)
	}
	return nil

}

func tokenGen() error {

	privatePEM, err := os.ReadFile("private.pem")
	if err != nil {
		return errors.Wrap(err, "reading PEM private key file")
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)

	if err != nil {
		log.Fatalln(err)
	}

	type JwtClaim struct {
		jwt.StandardClaims
		Authorize []string
	}
	claims := JwtClaim{jwt.StandardClaims{Issuer: "service project", Subject: "12324", ExpiresAt: time.Now().Add(8760 * time.Hour).Unix(), IssuedAt: time.Now().Unix()}, []string{"ADMIN"}}

	method := jwt.GetSigningMethod("RS256")
	tkn := jwt.NewWithClaims(method, claims)
	tkn.Header["kid"] = "3f433e9a-1bbc-4925-98f8-f4e119cd6bce"
	srt, err := tkn.SignedString(privateKey)

	if err != nil {
		return errors.Wrap(err, "generate token")
	}
	fmt.Println(srt)

	return err

}
