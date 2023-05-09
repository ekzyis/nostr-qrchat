package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/namsral/flag"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

var (
	NostrPubKey string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		err = fmt.Errorf("error loading .env: %w", err)
		log.Fatal(err)
	}
	flag.StringVar(&NostrPubKey, "NOSTR_PUBKEY", "", "Nostr pubkey of recipient")
	flag.Parse()
	if NostrPubKey == "" {
		log.Fatal("NOSTR_PUBKEY not set")
	}
}

func GenerateKeyPair() ([2]string, error) {
	sk := nostr.GeneratePrivateKey()
	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		err = fmt.Errorf("error getting pubkey from %s: %w", sk, err)
		return [2]string{"", ""}, err
	}
	nsec, err := nip19.EncodePrivateKey(sk)
	if err != nil {
		err = fmt.Errorf("error encoding private key %s: %w", sk, err)
		return [2]string{"", ""}, err
	}
	npub, err := nip19.EncodePublicKey(pk)
	if err != nil {
		err = fmt.Errorf("error encoding pubkey %s: %w", pk, err)
		return [2]string{"", ""}, err
	}
	return [2]string{nsec, npub}, nil
}
