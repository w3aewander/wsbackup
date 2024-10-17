package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"gopkg.in/ini.v1"
)

var myApp fyne.App // Declare myApp at the package level

func loadConfig(configFile string) (string, string, string, error) {
	cfg, err := ini.Load(configFile)
	if err != nil {
		return "", "", "", fmt.Errorf("Falha ao carregar o arquivo de configuração: %v", err)
	}

	REMOTE_IP := strings.TrimSpace(cfg.Section("backup").Key("remote_ip").String())
	REMOTE_DIR := strings.TrimSpace(cfg.Section("backup").Key("remote_dir").String())
	TARGET_BACKUP := strings.TrimSpace(cfg.Section("backup").Key("backup_name").String())

	return REMOTE_IP, REMOTE_DIR, TARGET_BACKUP, nil
}

func saveConfig(configFile, ip, dir, name string) error {
	cfg, err := ini.Load(configFile)
	if err != nil {
		return fmt.Errorf("Falha ao carregar o arquivo de configuração: %v", err)
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

func createHelpContent() string {
	return `Backup Remoto via SSH:
Esta aplicação realiza backups remotos de forma automatizada via SSH.
É recomendado criar uma chave pública RSA para evitar a necessidade de senhas.

Exemplo para criar uma chave:
$ ssh-keygen -t rsa -b 4096 -C "seu-email@example.com"
$ ssh-copy-id user@remote_host

O backup local é compactado com tar.gz e está armazenado no diretório ~/backup,
dentro do diretório padrão de usuários no Windows.

Para quaisquer dúvidas, entre em contato: Wanderlei Silva do Carmo <wander.silva@gmail.com>`
}

func showHelpWindow(parent fyne.Window) {
	helpContent := widget.NewLabel(createHelpContent())
	helpContent.Wrapping = fyne.TextWrapWord

	scroll := container.NewScroll(helpContent)
	scroll.SetMinSize(fyne.NewSize(700, 500)) // Tamanho mínimo para a janela de ajuda

	helpWindow := myApp.NewWindow("Ajuda")
	helpWindow.SetContent(container.NewVBox(
		scroll,
		layout.NewSpacer(),
		widget.NewButton("Fechar", func() {
			helpWindow.Close() // Fechar a janela de ajuda
		}),
	))
	helpWindow.Resize(fyne.NewSize(700, 500)) // Definir tamanho da janela de ajuda
	helpWindow.Show()
}

func showConfigWindow(parent fyne.Window) {
	REMOTE_IP, REMOTE_DIR, TARGET_BACKUP, err := loadConfig("config.ini")
	if err != nil {
		dialog.ShowError(err, parent)
		return
	}

	ipEntry := widget.NewEntry()
	ipEntry.SetText(REMOTE_IP)

	dirEntry := widget.NewEntry()
	dirEntry.SetText(REMOTE_DIR)

	nameEntry := widget.NewEntry()
	nameEntry.SetText(TARGET_BACKUP)

	form := widget.NewForm(
		widget.NewFormItem("IP Remoto", ipEntry),
		widget.NewFormItem("Diretório Remoto", dirEntry),
		widget.NewFormItem("Nome do Backup", nameEntry),
	)

	// Move configWindow declaration here, and close it in its scope
	configWindow := myApp.NewWindow("Configuração")

	saveButton := widget.NewButton("Salvar", func() {
		err := saveConfig("config.ini", ipEntry.Text, dirEntry.Text, nameEntry.Text)
		if err != nil {
			dialog.ShowError(err, parent)
		} else {
			dialog.ShowInformation("Configuração", "Configuração salva com sucesso.", parent)
		}
	})

	closeButton := widget.NewButton("Fechar", func() {
		configWindow.Close() // Fechar a janela de configuração
	})

	content := container.NewVBox(
		widget.NewLabel("Alterar Configurações"),
		form,
		saveButton,
		closeButton,
	)

	configWindow.SetContent(content)
	configWindow.Resize(fyne.NewSize(400, 300)) // Tamanho da janela de configuração
	configWindow.Show()
}

func confirmExit(win fyne.Window) {
	dialog.ShowConfirm("Sair", "Você realmente deseja sair?", func(confirm bool) {
		if confirm {
			win.Close()
		}
	}, win)
}

func main() {
	myApp = app.New() // Inicializar myApp
	myWindow := myApp.NewWindow("Backup Automatizado")

	logoImage := canvas.NewImageFromFile("./logo.png")
	logoImage.FillMode = canvas.ImageFillOriginal

	statusLabel := widget.NewLabel("Status: Aguardando...")
	progressBar := widget.NewProgressBar()
	spinner := widget.NewProgressBarInfinite()
	spinner.Hide()

	statusBar := widget.NewLabel("Versão: 1.0.0 | Data de Criação: 17/10/2024")

	startBackup := func() {
		statusLabel.SetText("Iniciando backup...")
		spinner.Show()

		REMOTE_IP, REMOTE_DIR, TARGET_BACKUP, err := loadConfig("config.ini")
		if err != nil {
			dialog.ShowError(err, myWindow)
			spinner.Hide()
			return
		}

		DATE := time.Now().Format("02-01-2006")
		localBackupDir := filepath.Join(os.Getenv("USERPROFILE"), "backup") // Mudei para usar USERPROFILE no Windows

		if err := createLocalBackupDir(localBackupDir); err != nil {
			dialog.ShowError(fmt.Errorf("Erro ao criar diretório local: %v", err), myWindow)
			spinner.Hide()
			return
		}

		sshCmd := exec.Command("ssh", "root@"+REMOTE_IP, "cd "+REMOTE_DIR+" && tar zcf /tmp/"+TARGET_BACKUP+"-"+DATE+".tar.gz ./")

		go func() {
			if err := sshCmd.Run(); err != nil {
				statusLabel.SetText("Erro ao criar backup remoto.")
				spinner.Hide()
				return
			}
			progressBar.SetValue(0.5)

			scpCmd := exec.Command("scp", "root@"+REMOTE_IP+":/tmp/"+TARGET_BACKUP+"-"+DATE+".tar.gz", localBackupDir)
			if err := scpCmd.Run(); err != nil {
				statusLabel.SetText("Erro ao transferir o backup.")
				spinner.Hide()
				return
			}

			progressBar.SetValue(1)
			spinner.Hide()
			statusLabel.SetText("Backup concluído com sucesso!")
		}()
	}

	startButton := widget.NewButton("Iniciar Backup", startBackup)

	menu := fyne.NewMainMenu(
		fyne.NewMenu("Arquivo",
			fyne.NewMenuItem("Configuração", func() { showConfigWindow(myWindow) }),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Sair", func() { confirmExit(myWindow) }),
		),
		fyne.NewMenu("Ajuda", fyne.NewMenuItem("Sobre", func() { showHelpWindow(myWindow) })),
	)

	myWindow.SetMainMenu(menu)

	content := container.NewVBox(
		logoImage,
		widget.NewLabel("Backup Automatizado v1.0.0"),
		startButton,
		statusLabel,
		spinner,
		progressBar,
		statusBar,
	)

	myWindow.SetContent(content)
	//myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.SetFixedSize(true)
	myWindow.SetCloseIntercept(func() { confirmExit(myWindow) })
	myWindow.ShowAndRun()
}
