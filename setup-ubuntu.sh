# Update and install software
sudo apt update
sudo apt install vim tmux git zsh mosh python cmake -y

# Install Oh-my-shell
sh -c "$(curl -fsSL https://raw.githubusercontent.com/robbyrussell/oh-my-zsh/master/tools/install.sh)"

# Get the dotfiles
cd ~
git clone https://github.com/overtune/dotfiles.git
cd dotfiles
bash makesymlinks.sh

vim +PlugInstall +qall
cd ~/.vim/plugged/YouCompleteMe && ./install.py
