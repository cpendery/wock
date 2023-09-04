package cert

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/adrg/xdg"
	"github.com/cpendery/mkcert"
)

var (
	enabledStores  = []string{"system", "nss"}
	WockCertFile   = filepath.Join(xdg.CacheHome, "wock", "cert.pem")
	WockKeyFile    = filepath.Join(xdg.CacheHome, "wock", "key.pem")
	logger         = log.New(os.Stdout, "", 0)
	verboseLogging = false
)

const (
	mkcertLogPrefix           = "[mkcert] "
	wockUnsafeInstallVariable = "WOCK_UNSAFE_INSTALL"
)

func SetVerbose(verbose bool) {
	verboseLogging = verbose
}

func setupLogging() {
	log.Default().SetPrefix(mkcertLogPrefix)
	if !verboseLogging {
		log.Default().SetOutput(io.Discard)
	}
}

func tearDownLogging() {
	log.Default().SetPrefix("")
	log.Default().SetOutput(os.Stdout)
}

func IsInstalled() (isInstalled bool) {
	cert := mkcert.MKCert{
		EnabledStores: enabledStores,
	}
	if err := cert.Load(); err != nil {
		isInstalled = false
		return
	}
	nss := cert.CheckNSS()
	platform := cert.CheckPlatform()
	switch runtime.GOOS {
	case "darwin", "linux":
		isInstalled = nss && platform
	default:
		isInstalled = platform
	}
	return
}

func Install() error {
	setupLogging()
	defer tearDownLogging()
	b, err := strconv.ParseBool(os.Getenv(wockUnsafeInstallVariable))
	unsafeInstall := b && err == nil

	cert := mkcert.MKCert{
		EnabledStores:                      enabledStores,
		UnsafeWindowsAdminCertInstallation: unsafeInstall,
	}
	if err := cert.Load(); err != nil {
		return err
	}
	isInstalled := IsInstalled()
	if isInstalled {
		logger.Println("Local CA is already installed")
		return nil
	}
	if err := cert.Install(); err != nil {
		return fmt.Errorf("failed to install local CA: %w", err)
	}
	logger.Println("Successfully installed/verified local CA")
	return nil
}

func Uninstall() error {
	setupLogging()
	defer tearDownLogging()
	cert := mkcert.MKCert{
		EnabledStores: enabledStores,
	}
	cert.Load()
	if err := cert.Uninstall(); err != nil {
		return fmt.Errorf("failed to uninstall local CA: %w", err)
	}
	logger.Println("Successfully uninstalled local CA")
	return nil
}

func CreateCert(hosts []string) error {
	cert := mkcert.MKCert{
		CertFile:      WockCertFile,
		KeyFile:       WockKeyFile,
		EnabledStores: enabledStores,
	}
	return cert.Run(hosts)
}
