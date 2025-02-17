// https://gist.githubusercontent.com/miguelmota/ee0fd9756e1651f38f4cd38c6e99b8bf/raw/7c470e0b84caddb5aed3e8cfe761333d642a6471/hdwallet.go
package main

import (
	"bufio"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"os"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

// Wallet ...
type Wallet struct {
	mnemonic    string
	path        string
	root        *hdkeychain.ExtendedKey
	extendedKey *hdkeychain.ExtendedKey
	privateKey  *ecdsa.PrivateKey
	publicKey   *ecdsa.PublicKey
}

// Config ...
type Config struct {
	Mnemonic string
	Path     string
}

// New ...
func New(config *Config) (*Wallet, error) {
	if config.Path == "" {
		config.Path = `m/44'/60'/0'/0`
	}

	if config.Mnemonic == "" {
		return nil, errors.New("mnemonic is required")
	}

	seed := bip39.NewSeed(config.Mnemonic, "")
	dpath, err := accounts.ParseDerivationPath(config.Path)
	if err != nil {
		return nil, err
	}

	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}

	key := masterKey

	for _, n := range dpath {
		key, err = key.Child(n)
		if err != nil {
			return nil, err
		}
	}

	privateKey, err := key.ECPrivKey()
	privateKeyECDSA := privateKey.ToECDSA()
	if err != nil {
		return nil, err
	}

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("failed ot get public key")
	}

	wallet := &Wallet{
		mnemonic:    config.Mnemonic,
		path:        config.Path,
		root:        masterKey,
		extendedKey: key,
		privateKey:  privateKeyECDSA,
		publicKey:   publicKeyECDSA,
	}

	return wallet, nil
}

// Derive ...
func (s Wallet) Derive(index interface{}) (*Wallet, error) {
	var idx uint32
	switch v := index.(type) {
	case int:
		idx = uint32(v)
	case int8:
		idx = uint32(v)
	case int16:
		idx = uint32(v)
	case int32:
		idx = uint32(v)
	case int64:
		idx = uint32(v)
	case uint:
		idx = uint32(v)
	case uint8:
		idx = uint32(v)
	case uint16:
		idx = uint32(v)
	case uint32:
		idx = v
	case uint64:
		idx = uint32(v)
	default:
		return nil, errors.New("unsupported index type")
	}

	address, err := s.extendedKey.Child(idx)
	if err != nil {
		return nil, err
	}

	privateKey, err := address.ECPrivKey()
	privateKeyECDSA := privateKey.ToECDSA()
	if err != nil {
		return nil, err
	}

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("failed ot get public key")
	}

	path := fmt.Sprintf("%s/%v", s.path, idx)

	wallet := &Wallet{
		path:        path,
		root:        s.extendedKey,
		extendedKey: address,
		privateKey:  privateKeyECDSA,
		publicKey:   publicKeyECDSA,
	}

	return wallet, nil
}

// PrivateKey ...
func (s Wallet) PrivateKey() *ecdsa.PrivateKey {
	return s.privateKey
}

// PrivateKeyBytes ...
func (s Wallet) PrivateKeyBytes() []byte {
	return crypto.FromECDSA(s.PrivateKey())
}

// PrivateKeyHex ...
func (s Wallet) PrivateKeyHex() string {
	return hexutil.Encode(s.PrivateKeyBytes())[2:]
}

// PublicKey ...
func (s Wallet) PublicKey() *ecdsa.PublicKey {
	return s.publicKey
}

// PublicKeyBytes ...
func (s Wallet) PublicKeyBytes() []byte {
	return crypto.FromECDSAPub(s.PublicKey())
}

// PublicKeyHex ...
func (s Wallet) PublicKeyHex() string {
	return hexutil.Encode(s.PublicKeyBytes())[4:]
}

// Address ...
func (s Wallet) Address() common.Address {
	return crypto.PubkeyToAddress(*s.publicKey)
}

// AddressHex ...
func (s Wallet) AddressHex() string {
	return s.Address().Hex()
}

// Path ...
func (s Wallet) Path() string {
	return s.path
}

// Mnemonic ...
func (s Wallet) Mnemonic() string {
	return s.mnemonic
}

// NewMnemonic ...
func NewMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}

// NewSeed ...
func NewSeed() ([]byte, error) {
	return bip32.NewSeed()
}

func main() {
	fmt.Println("DISCLAIMER: THE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.")
	fmt.Println("This piece of code can be used to recover your private key from your mnemonic, if you're facing trouble accessing it via the Harmony One Wallet or Metamask. Please note that you alone are responsible for the safety of your private key and mnemonic phrase, and take precautions that you deem appropriate before running this application.")
	in := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter your mnemonic phrase: ")
	line, err := in.ReadString('\n')
	if err != nil {
		fmt.Println("Could not read mnemonic phrase due to", err)
		os.Exit(1)
	}
	_, err = bip39.EntropyFromMnemonic(line)
	if err != nil {
		fmt.Println("Invalid mnemonic entered")
		os.Exit(1)
	}
	derivationPaths := []string{"m/44'/1023'/0'/0/0", "m/44'/60'/0'/0/0"}
	for _, derivationPath := range derivationPaths {
		config := &Config{Mnemonic: line, Path: derivationPath}
		wallet, err := New(config)
		if err != nil {
			fmt.Printf("Could not determine key with path %s due to %s\n", derivationPath, err)
			continue
		}
		fmt.Printf("Your private key with derivation path %s is\n%s\n", derivationPath, wallet.PrivateKeyHex())
	}
}
