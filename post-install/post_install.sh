#!/bin/bash

_install_nvim()
{
    echo "install nvim"
    sudo apt-get install ninja-build gettext cmake unzip curl -y 1&>/dev/null
    git clone https://github.com/neovim/neovim
    (cd neovim && make CMAKE_BUILD_TYPE=Release -j `nproc` && sudo make install)
}

_install_base()
{
    echo "install base"
    sudo apt update && sudo apt full-upgrade -y 1&>/dev/null
    sudo apt install git zsh htop tmux apt-transport-https ca-certificates curl gnupg2 software-properties-common -y 1&>/dev/null
}

_install_containers()
{
    echo "install docker"
    sudo curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
    sudo echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list
    sudo apt-get update
    sudo apt-get install docker-ce docker-ce-cli containerd.io docker docker-compose
    sudo systemctl enable docker
}

_manage_kawaegle()
{
    echo "manage kawaegle"
    sudo useradd -U -m -G sudo kawaegle
    pass="$(</dev/urandom tr -dc 'A-Za-z0-9!#$%&()*+,-./:;<=>?@[\]^_{|}' | head -c 15; echo)"
    echo -e "$pass\n$pass" | sudo passwd kawaegle
    sudo chsh -s /bin/zsh kawaegle
    (su kawaegle && mkdir -p /home/kawaegle/.ssh/ /home/kawaegle/.local/{bin,share} /home/kawaegle/.cache/ && \
        echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDjD/5+bJ3jR/OqkJMrolos4vgf2Xlo3uShaD5OBRxXWH1vPjzRWWyM3rDd3k4b+lz1GRcaASMmpSSk9nIT9MYI/qQsNoWFCh2KAhxgK+igB4eYh0Yn+jIBRiqPcY1LEdcmW5+dw/mOPq4rycFD0hTJ47vcwc1aUmJAO8xC5pdkiK3MeaALhlO4uRkztpdxk5idMudNQC30EB2jgAaTlaB7dQqF4pfPKyYSPCHwP1SvP1MbXXrOja3pJzUoqMnCpEa3dt+HYhKYWCPy6MvKiMrsxNTMluXAw5+B7vxVGatUWedaEkHpGV9lybebl1jRL/COWSJPtx1QJs6Q56KIFXR5EypK8b+JOkD2tSFImgsxOohXuqGznnZ1TXDBcyv7jn0g9N4jy8ZC3c8+bLiOLS7sZ8UwOfeGS8GIkMcKnNbkCj+9ukiskXvXM8swfh/DIuAPwAg3154ofZcrlBR028lTOXVqtZEABDIY2NmZqhfyY6BH07q75Uy7NPudIxQuy80EAGtHc0gz2FVxM+OYmnGKPPCHw9yb0x8IfXdeYmvaBSTr40ZOs6cBYK/9pOGQA4Fiw+o7HspW4UrrUvKM2s+VAvwOZEIJOMiuEO9+NRGJ0kkZgbuiiMCbAjkFBZPmmwKAeh3X9gECfAmN8/Cd11rRyHOUorvgJgfERtmOyqBr7w== kawaegle@OppaiLaptop" >> /home/kawaegle/.ssh/authorized_keys)
}

_manage_users()
{
    echo "manage user and default"
    pass=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 13 ; echo '')
    echo -e "$pass\n$pass" | sudo passwd debian
    echo "$pass" | sudo tee -a /home/debian/psw
    if [[ ! -d /home/kawaegle/ ]]; then
        _manage_kawaegle
    fi

    pass=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 13 ; echo '')
    echo -e "$pass\n$pass" | sudo passwd root
    echo "$pass" | sudo tee -a /root_psw
}
_setup_ssh()
{
    echo "setup ssh"
    file=$(mktemp)
    cat /etc/ssh/sshd_config | sed "s/#Port 22/Port 2022/" | sed "s/#PermitRootLogin.*/PermitRootLogin no/" | sed "s/#PubkeyAuthentication.*/PubkeyAuthentication yes/" > $file
    sudo rm /etc/ssh/sshd_config
    cat $file | sudo tee -a /etc/ssh/sshd_config > /dev/null
    sudo systemctl restart sshd
}


_setup_server()
{
    _setup_ssh
    sudo rm -rf /etc/sudoers.d/90-cloud-init-users
}

_install_base
_install_nvim
_manage_users
_setup_server
_install_containers
reboot
