package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

func loadConfig(configFile string) (string, string, string, error) {
	cfg, err := ini.Load(configFile)
	if err != nil {
		return "", "", "", fmt.Errorf("falha ao carregar o arquivo de configuração: %v", err)
	}

	REMOTE_IP := strings.TrimSpace(cfg.Section("backup").Key("remote_ip").String())
	REMOTE_DIR := strings.TrimSpace(cfg.Section("backup").Key("remote_dir").String())
	TARGET_BACKUP := strings.TrimSpace(cfg.Section("backup").Key("backup_name").String())

	return REMOTE_IP, REMOTE_DIR, TARGET_BACKUP, nil
}

func saveConfig(configFile, ip, dir, name string) error {
	cfg, err := ini.Load(configFile)
	if err != nil {
		return fmt.Errorf("falha ao carregar o arquivo de configuração: %v", err)
	}
	cfg.Section("backup").Key("remote_ip").SetValue(ip)
	cfg.Section("backup").Key("remote_dir").SetValue(dir)
	cfg.Section("backup").Key("backup_name").SetValue(name)
	return cfg.SaveTo(configFile)
}

func createLocalBackupDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}

func printHelp() {
	fmt.Println(`Backup Remoto via SSH:
Esta aplicação realiza backups remotos de forma automatizada via SSH.
É recomendado criar uma chave pública RSA para evitar a necessidade de senhas.

Exemplo para criar uma chave:
$ ssh-keygen -t rsa -b 4096 -C "seu-email@example.com"
$ ssh-copy-id user@remote_host

O backup local é compactado com tar.gz e está armazenado no diretório padrão de usuários no Windows.

Para quaisquer dúvidas, entre em contato: Wanderlei Silva do Carmo <wander.silva@gmail.com>`)
}

func main() {
	configFile := "config.ini"
	REMOTE_IP, REMOTE_DIR, TARGET_BACKUP, err := loadConfig(configFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		fmt.Println("\nEscolha uma opção:")
		fmt.Println("1. Iniciar Backup")
		fmt.Println("2. Configuração")
		fmt.Println("3. Ajuda")
		fmt.Println("4. Sair")

		var choice int
		fmt.Scan(&choice)

		switch choice {
		case 1:
			fmt.Println("Iniciando backup...")

			DATE := time.Now().Format("02-01-2006")
			localBackupDir := filepath.Join(os.Getenv("USERPROFILE"), "backup")

			if err := createLocalBackupDir(localBackupDir); err != nil {
				fmt.Printf("Erro ao criar diretório local: %v\n", err)
				continue
			}

			sshCmd := exec.Command("ssh", "root@"+REMOTE_IP, "cd "+REMOTE_DIR+" && tar zcf /tmp/"+TARGET_BACKUP+"-"+DATE+".tar.gz ./")
			if err := sshCmd.Run(); err != nil {
				fmt.Println("Erro ao criar backup remoto:", err)
				continue
			}

			scpCmd := exec.Command("scp", "root@"+REMOTE_IP+":/tmp/"+TARGET_BACKUP+"-"+DATE+".tar.gz", localBackupDir)
			if err := scpCmd.Run(); err != nil {
				fmt.Println("Erro ao transferir o backup:", err)
				continue
			}

			fmt.Println("Backup concluído com sucesso!")

		case 2:
			var ip, dir, name string
			fmt.Print("Insira o IP Remoto: ")
			fmt.Scan(&ip)
			fmt.Print("Insira o Diretório Remoto: ")
			fmt.Scan(&dir)
			fmt.Print("Insira o Nome do Backup: ")
			fmt.Scan(&name)

			err := saveConfig(configFile, ip, dir, name)
			if err != nil {
				fmt.Println("Erro ao salvar configuração:", err)
			} else {
				fmt.Println("Configuração salva com sucesso.")
			}

		case 3:
			printHelp()

		case 4:
			fmt.Println("Saindo...")
			return

		default:
			fmt.Println("Opção inválida. Tente novamente.")
		}
	}
}
