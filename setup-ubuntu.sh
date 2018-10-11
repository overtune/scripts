# Update and install software
sudo apt update
sudo add-apt-repository universe
sudo apt install vim tmux git zsh mosh build-essential cmake python3-dev -y

# Install Oh-my-shell
curl -fsSL https://raw.githubusercontent.com/robbyrussell/oh-my-zsh/master/tools/install.sh
sh install.sh
rm install.sh

# Get the dotfiles
cd ~
git clone https://github.com/overtune/dotfiles.git
cd dotfiles
bash makesymlinks.sh

# Install Vim plugs
vim +PlugInstall +qall +silent
cd ~/.vim/plugged/YouCompleteMe && python3 install.py 
