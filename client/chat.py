#!/usr/bin/env python3
import json
import urllib.request
import hashlib
import sys
import platform
import os
import time

IS_WINDOWS = platform.system() == "Windows"
IS_TERMUX = os.getenv("TERMUX_VERSION") is not None
IS_ANDROID = IS_TERMUX or "ANDROID" in os.getenv("PATH", "").upper()

if IS_WINDOWS:
    try:
        import ctypes
        kernel32 = ctypes.windll.kernel32
        kernel32.SetConsoleMode(kernel32.GetStdHandle(-11), 7)
    except:
        pass

class Colors:
    SUPPORTS_COLOR = (
        hasattr(sys.stdout, 'isatty') and sys.stdout.isatty()
    ) or IS_TERMUX or os.getenv('TERM') in ['xterm', 'xterm-color', 'xterm-256color', 'screen', 'linux']
    
    if SUPPORTS_COLOR:
        RESET = "\033[0m"
        BOLD = "\033[1m"
        DIM = "\033[2m"
        RED = "\033[31m"
        GREEN = "\033[32m"
        YELLOW = "\033[33m"
        BLUE = "\033[34m"
        MAGENTA = "\033[35m"
        CYAN = "\033[36m"
        YELLOW = "\033[0;38;2;255;255;0;49m"
        PURPLE = "\033[0;38;2;144;0;255;49m"
        WHITE= "\033[0;38;2;255;255;255;49m"
    else:
        RESET = BOLD = DIM = ""
        RED = GREEN = YELLOW = BLUE = MAGENTA = CYAN = ""

def hash_password(password: str) -> str:
    return hashlib.sha256(password.encode("utf-8")).hexdigest()

def clear_screen():
    if IS_WINDOWS:
        os.system('cls')
    else:
        os.system('clear')

def print_banner():
    banner = f"""
{Colors.YELLOW}{Colors.BOLD}=====================================================
                    {Colors.RED}CHAT
        {Colors.PURPLE}MADE BY SWITCHCODEUR & STYLOBOW
                {Colors.GREEN}VERSION 0.1
      {Colors.RED}{Colors.DIM}Github : https://github.com/swtchcoder/chat {Colors.RESET}
{Colors.YELLOW}{Colors.BOLD}====================================================={Colors.RESET}
"""
    print(banner)

def main():
    clear_screen()
    print_banner()

    print(f"{Colors.BLUE}Bienvenue dans le Chat.")
    print(f"{Colors.WHITE}Veuillez choisir une des options suivante pour rejoindre le chat:")
    print(f"\033[0;38;2;0;255;0;49mSe connecter (1)")
    print(f"\033[0;38;2;255;0;0;49mCréer un compte (2){Colors.RESET}")
    action = input("").strip()
    if action == "1":
        URL = "https://chat.switchcodeur.com/login"
    elif action == "2":
        URL = "https://chat.switchcodeur.com/register"
    else:
        main()
        time.spleep(1)
        print("Veuillez choisir un nombre entre 1 et 2.")

    data = {
        "username": input("Nom d'utilisateur: "),
        "password": hash_password(input("Mot de passe: "))
    }

    req = urllib.request.Request(
        URL,
        data=json.dumps(data).encode("utf-8"),
        headers={"Content-Type": "application/json"},
        method="POST"
    )

    try:
        with urllib.request.urlopen(req) as resp:
            print(f"Status: {resp.status}")
            body = resp.read()
            if body:
                print(body.decode())
    except urllib.error.HTTPError as e:
        print(f"HTTP error: {e.code}")
    except Exception as e:
        print("Request failed:", e)
    except KeyboardInterrupt:
        print(f"\n{Colors.YELLOW}Deconnexion...{Colors.RESET}")
        sys.exit(0)

main()
