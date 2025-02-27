package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/filecoin-project/bacalhau/pkg/storage/util"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/rs/zerolog/log"
)

func IsDebug() bool {
	return os.Getenv("LOG_LEVEL") == "debug"
}

func DevstackGetShouldPrintInfo() bool {
	return os.Getenv("DEVSTACK_PRINT_INFO") != ""
}

func DevstackSetShouldPrintInfo() {
	os.Setenv("DEVSTACK_PRINT_INFO", "1")
}

func ShouldKeepStack() bool {
	return os.Getenv("KEEP_STACK") != ""
}

func GetStoragePath() string {
	storagePath := os.Getenv("BACALHAU_STORAGE_PATH")
	if storagePath == "" {
		storagePath = os.TempDir()
	}
	return storagePath
}

func GetAPIHost() string {
	return os.Getenv("BACALHAU_HOST")
}

func GetAPIPort() string {
	return os.Getenv("BACALHAU_PORT")
}

// by default we wait 2 minutes for the IPFS network to resolve a CID
// tests will override this using config.SetVolumeSizeRequestTimeout(2)
var getVolumeSizeRequestTimeoutSeconds int64 = 120

// how long do we wait for a volume size request to timeout
// if a non-existing cid is asked for - the dockerIPFS.IPFSClient.GetCidSize(ctx, volume.Cid)
// function will hang for a long time - so we wrap that call in a timeout
// for tests - we only want to wait for 2 seconds because everything is on a local network
// in prod - we want to wait longer because we might be running a job that is
// using non-local CIDs
// the tests are expected to call SetVolumeSizeRequestTimeout to reduce this timeout
func GetVolumeSizeRequestTimeout() time.Duration {
	return time.Duration(getVolumeSizeRequestTimeoutSeconds) * time.Second
}

func SetVolumeSizeRequestTimeout(seconds int64) {
	getVolumeSizeRequestTimeoutSeconds = seconds
}

// by default we wait 5 minutes for the IPFS network to download a CID
// tests will override this using config.SetVolumeSizeRequestTimeout(2)
var downloadCidRequestTimeoutSeconds int64 = 300

// how long do we wait for a cid to download
func GetDownloadCidRequestTimeout() time.Duration {
	return time.Duration(downloadCidRequestTimeoutSeconds) * time.Second
}

func SetDownloadCidRequestTimeout(seconds int64) {
	downloadCidRequestTimeoutSeconds = seconds
}

// by default we wait 5 minutes for a URL to download
// tests will override this using config.SetDownloadURLRequestTimeoutSeconds(2)
var downloadURLRequestTimeoutSeconds int64 = 300

// how long do we wait for a URL to download
func GetDownloadURLRequestTimeout() time.Duration {
	return time.Duration(downloadURLRequestTimeoutSeconds) * time.Second
}

func SetDownloadURLRequestTimeoutSeconds(seconds int64) {
	downloadURLRequestTimeoutSeconds = seconds
}

func GetConfigPath() string {
	suffix := "/.bacalhau"
	env := os.Getenv("BACALHAU_PATH")
	var d string
	if env == "" {
		// e.g. /home/francesca/.bacalhau
		dirname, err := os.UserHomeDir()
		if err != nil {
			log.Fatal().Err(err)
		}
		d = dirname + suffix
	} else {
		// e.g. /data/.bacalhau
		d = env + suffix
	}
	// create dir if not exists
	if err := os.MkdirAll(d, util.OS_USER_RWX); err != nil {
		log.Fatal().Err(err)
	}
	return d
}

const BitsForKeyPair = 2048

func GetPrivateKey(keyName string) (crypto.PrivKey, error) {
	configPath := GetConfigPath()

	// We include the port in the filename so that in devstack multiple nodes
	// running on the same host get different identities
	privKeyPath := fmt.Sprintf("%s/%s", configPath, keyName)

	if _, err := os.Stat(privKeyPath); errors.Is(err, os.ErrNotExist) {
		// Private key does not exist - create and write it

		// Creates a new RSA key pair for this host.
		prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, BitsForKeyPair, rand.Reader)
		if err != nil {
			log.Error().Err(err)
			return nil, err
		}

		keyOut, err := os.OpenFile(privKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, util.OS_USER_RW)
		if err != nil {
			return nil, fmt.Errorf("failed to open key.pem for writing: %v", err)
		}
		privBytes, err := crypto.MarshalPrivateKey(prvKey)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal private key: %v", err)
		}
		// base64 encode privBytes
		b64 := base64.StdEncoding.EncodeToString(privBytes)
		_, err = keyOut.WriteString(b64 + "\n")
		if err != nil {
			return nil, fmt.Errorf("failed to write to key file: %v", err)
		}
		if err := keyOut.Close(); err != nil {
			return nil, fmt.Errorf("error closing key file: %v", err)
		}
		log.Printf("wrote %s", privKeyPath)
	}

	// Now that we've ensured the private key is written to disk, read it! This
	// ensures that loading it works even in the case where we've just created
	// it.

	// read the private key
	keyBytes, err := os.ReadFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %v", err)
	}
	// base64 decode keyBytes
	b64, err := base64.StdEncoding.DecodeString(string(keyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %v", err)
	}
	// parse the private key
	prvKey, err := crypto.UnmarshalPrivateKey(b64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	return prvKey, nil
}
