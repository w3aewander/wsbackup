package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
	"gopkg.in/ini.v1"
)

type Config struct {
	RemoteIP   string `ini:"remote_ip"`
	RemotePort string `ini:"remote_port"`
	RemoteDir  string `ini:"remote_dir"`
	Username   string `ini:"username"`
	Password   string `ini:"password"`
	LocalDir   string `ini:"local_dir"`
	BackupName string `ini:"backup_name"`
}

func main() {
	// Lê o arquivo de configuração
	var cfg Config

	// Lê o arquivo de configuração
	iniFile, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalf("Erro ao ler config.ini: %v", err)
	}

	// Exibe os parâmetros de configuração
	printHeader("Configurações do Backup")
	backupSection := iniFile.Section("backup")
	cfg.RemoteIP = backupSection.Key("remote_ip").String()
	cfg.RemotePort = backupSection.Key("remote_port").String()
	cfg.RemoteDir = backupSection.Key("remote_dir").String()
	cfg.Username = backupSection.Key("username").String()
	cfg.Password = backupSection.Key("password").String()
	cfg.LocalDir = backupSection.Key("local_dir").String()
	cfg.BackupName = backupSection.Key("backup_name").String()

	fmt.Printf("IP Remoto: %s\n", cfg.RemoteIP)
	fmt.Printf("Porta Remota: %s\n", cfg.RemotePort)
	fmt.Printf("Diretório Remoto: %s\n", cfg.RemoteDir)
	fmt.Printf("Nome do backup: %s\n", cfg.BackupName)
	fmt.Printf("Diretório Local para Backup: %s\n", cfg.LocalDir)
	fmt.Println()

	// Verifica se os parâmetros foram carregados corretamente
	if cfg.RemoteIP == "" || cfg.RemotePort == "" || cfg.RemoteDir == "" || cfg.LocalDir == "" || cfg.BackupName == "" {
		log.Fatalf("Erro: Um ou mais parâmetros de configuração estão vazios.")
	}

	// Cria o diretório de backup se não existir
	err = os.MkdirAll(cfg.LocalDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Erro ao criar diretório de backup: %v", err)
	}

	// Realiza o backup
	err = performBackup(cfg)
	if err != nil {
		log.Fatalf("Erro ao realizar backup: %v", err)
	}

	printSuccess("Backup realizado com sucesso!")
}

func performBackup(cfg Config) error {
	printStatus("Preparando...")

	// Configurações do cliente SSH
	sshConfig := &ssh.ClientConfig{
		User: cfg.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(cfg.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Conexão SSH
	printStatus("Acessando servidor remoto...")
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", cfg.RemoteIP, cfg.RemotePort), sshConfig)
	if err != nil {
		return fmt.Errorf("falha na conexão SSH: %w", err)
	}
	defer client.Close()

	printStatus("Preparando arquivos para o backup...")
	// Executa o comando de backup no servidor remoto
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("falha ao criar sessão SSH: %w", err)
	}
	defer session.Close()

	// Executa o comando para compactar os arquivos no diretório remoto
	printStatus("Realizando o backup...")
	tarCmd := fmt.Sprintf("tar -czf /tmp/%s.tar.gz -C %s .", cfg.BackupName, cfg.RemoteDir)
	if err := session.Run(tarCmd); err != nil {
		return fmt.Errorf("erro ao executar comando de backup: %w", err)
	}

	// Copia o arquivo de backup para o diretório local
	backupName := "/tmp/%s.tar.gz"
	remoteBackupPath := fmt.Sprintf(backupName, cfg.BackupName)
	localBackupPath := filepath.Join(cfg.LocalDir, fmt.Sprintf("%s.tar.gz", cfg.BackupName))

	printStatus("Fazendo download do backup...")
	err = downloadFile(client, remoteBackupPath, localBackupPath)
	if err != nil {
		return fmt.Errorf("erro ao baixar o backup: %w", err)
	}

	printStatus("Apagando o arquivo temporário do backup...")
	// Remove o arquivo de backup remoto
	removeSession, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("falha ao criar sessão para remoção: %w", err)
	}
	defer removeSession.Close()

	if err := removeSession.Run("rm " + remoteBackupPath); err != nil {
		return fmt.Errorf("erro ao remover arquivo de backup remoto: %v", err)
	}
	return nil
}

func downloadFile(client *ssh.Client, remotePath, localPath string) error {
	// Cria uma nova sessão SSH
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Cria um pipe para receber os dados do servidor
	remoteFile, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	// Inicia a sessão para executar o comando de cópia
	if err := session.Start(fmt.Sprintf("cat %s", remotePath)); err != nil {
		return err
	}

	// Cria o arquivo local
	localFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	// Lê os dados do arquivo remoto e escreve no arquivo local
	if _, err := io.Copy(localFile, remoteFile); err != nil {
		return err
	}

	// Aguarda o término da sessão
	if err := session.Wait(); err != nil {
		return err
	}

	return nil
}

func printHeader(title string) {
	fmt.Println("===========================================")
	fmt.Printf("            %s\n", title)
	fmt.Println("===========================================")
}

func printSuccess(message string) {
	fmt.Printf("\033[32m%s\033[0m\n", message) // Texto verde
}

func printStatus(message string) {
	fmt.Printf("\r\033[36m%s\033[0m", message) // Texto ciano
	for i := 0; i < 3; i++ {
		fmt.Print(".")
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println()
}
