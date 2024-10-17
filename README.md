# Backup Remoto via SSH:
> 
> Autor: Wanderlei Silva do Carmo
> Eng. Arquiteto de Software
> 

## Esta aplicação realiza backups remotos de forma automatizada via SSH.
É recomendado criar uma chave pública RSA para evitar a necessidade de senhas.

### Exemplo para criar uma chave:

<code> 

$ ssh-keygen -t rsa -b 4096 -C "seu-email@example.com"
$ ssh-copy-id user@remote_host

</code>

<p>O backup local é compactado com tar.gz e está armazenado no diretório padrão de usuários no Windows.
</p>

<p> Os arquivos executáveis para Linux e Windows podem ser baixados e juntamente com o arquivo ini de exemplo <strong>config.ini.exemplo</strong>. 
</p>

<p> O arquivo <strong>config.ini.exemplo</strong> deve ser renomeado para <strong>config.ini</strong>
</p>



Para quaisquer dúvidas, entre em contato: wande.silva@gmail.com

