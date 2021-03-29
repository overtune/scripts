#!/usr/bin/env python3
"""
Author : Johan Runesson
Purpose: Setup a mac 
"""

import argparse
import os


# Terminal color class
class bcolors:
    OK = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'


# --------------------------------------------------
def get_args():
    """Get command-line arguments"""

    parser = argparse.ArgumentParser(
        description='Setup a mac',
        formatter_class=argparse.ArgumentDefaultsHelpFormatter)

    parser.add_argument(
        '-s',
        '--skip',
        help='A comma separated string of installations to skip',
        metavar='str',
        type=str,
        default='')

    parser.add_argument('-e',
                        '--email',
                        help='Email for Apple ID',
                        metavar='str',
                        type=str,
                        default='')

    return parser.parse_args()


# --------------------------------------------------
def shouldRun(name):
    """Validate if we should run the installation"""

    args = get_args()
    skip = args.skip.split(',')

    if name in skip:
        print(f'{bcolors.WARNING}Skipping {name}')
        return False
    else:
        print(f'{bcolors.OK}Installing {name}')
        return True


# --------------------------------------------------
def main():
    """Let's do this!"""

    args = get_args()
    email = args.email

    # Setup folder structure
    if shouldRun('folder-structure'):
        os.system('mkdir -p  ~/Development/work')
        os.system('mkdir -p  ~/Development/private')

    # Install homebrew
    if shouldRun('homebrew'):
        os.system(
            'curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh'
        )

    # Install mas
    if shouldRun('mas'):
        os.system('brew install mas')
        if email:
            os.system('mas signin ' + email)
        else:
            print(f'{bcolors.WARNING}No email for Apple ID provided')

    # Install alacritty
    if shouldRun('alacritty'):
        os.system('brew install alacritty')

    # Install iA Writer
    if shouldRun('iawriter'):
        os.system('mas install 775737590 ')

    # Install docker
    if shouldRun('iawriter'):
        os.system('brew install docker')

    # Install qmk-toolbox
    if shouldRun('qmktoolbox'):
        os.system('brew install qmk-toolbox')

    # Install vim
    if shouldRun('vim'):
        os.system('brew install vim')

    # Install tmux
    if shouldRun('tmux'):
        os.system('brew install tmux')

    # Install thefuck
    if shouldRun('thefuck'):
        os.system('brew install thefuck')

    # Install ack
    if shouldRun('ack'):
        os.system('brew install ack')

    # Install fff
    if shouldRun('fff'):
        os.system('brew install fff')

    # Install curl
    if shouldRun('curl'):
        os.system('brew install curl')

    # Install git
    if shouldRun('git'):
        os.system('brew install git')

    # Install asdf
    # if shouldRun('asdf'):
    #     os.system('brew install coreutils')
    #     os.system('brew install asdf')

    # Install node
    if shouldRun('nodejs'):
        os.system('brew install nodejs')


# --------------------------------------------------
if __name__ == '__main__':
    main()
