#!/usr/bin/env python3
import json
import urllib.request
import hashlib
import sys
import platform
import os
import time
import sys

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

if sys.platform.startswith("win"):
    import msvcrt
else:
    import tty, termios

def input_password(prompt="Mot de passe: "):
    print(prompt, end="", flush=True)
    password = ""
    if sys.platform.startswith("win"):
        while True:
            ch = msvcrt.getch()
            if ch in {b'\r', b'\n'}:
                print()
                break
            elif ch == b'\x08':
                if len(password) > 0:
                    password = password[:-1]

                    sys.stdout.write('\b \b')
                    sys.stdout.flush()
            elif ch == b'\x03':
                raise KeyboardInterrupt
            else:
                password += ch.decode("utf-8")
                sys.stdout.write("*")
                sys.stdout.flush()
    else:
        fd = sys.stdin.fileno()
        old_settings = termios.tcgetattr(fd)
        try:
            tty.setraw(fd)
            while True:
                ch = sys.stdin.read(1)
                if ch in ('\r', '\n'):
                    print()
                    break
                elif ch == '\x7f': 
                    if len(password) > 0:
                        password = password[:-1]
                        sys.stdout.write('\b \b')
                        sys.stdout.flush()
                elif ch == '\x03':
                    raise KeyboardInterrupt
                else:
                    password += ch
                    sys.stdout.write("*")
                    sys.stdout.flush()
        finally:
            termios.tcsetattr(fd, termios.TCSADRAIN, old_settings)
    return password

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
    print(f"{Colors.GREEN}Se connecter (1)")
    print(f"{Colors.RED}Créer un compte (2){Colors.RESET}")
    action = input("").strip()

    if action == "1":
        URL = "https://chat.switchcodeur.com/login"
        is_register = False
    elif action == "2":
        URL = "https://chat.switchcodeur.com/register"
        is_register = True
    else:
        print(f"{Colors.YELLOW}Veuillez choisir un nombre entre 1 et 2.{Colors.RESET}")
        time.sleep(2.25)
        main()
        return

    username = input(f"Nom d'utilisateur: {Colors.BLUE}").strip()
    password = input_password(f"{Colors.RESET}Mot de passe: {Colors.BLUE}")

    if is_register and (len(username) < 1 or len(password) < 1):
        print(f"{Colors.RED}Erreur : le pseudo et le mot de passe doivent faire au moins 1 caractère.{Colors.RESET}")
        time.sleep(1)
        main()
        return

    data = {
        "username": username,
        "password": hash_password(password)
    }

    req = urllib.request.Request(
        URL,
        data=json.dumps(data).encode("utf-8"),
        headers={"Content-Type": "application/json"},
        method="POST"
    )

    try:
        with urllib.request.urlopen(req) as resp:
            status = resp.status
            body = resp.read().decode()
            if is_register:
                if status == 201:
                    print(f"{Colors.GREEN}Compte créé avec succès ! Vous pouvez maintenant vous connecter.{Colors.RESET}")
                else:
                    print(f"{Colors.RED}Erreur lors de la création du compte : {body}{Colors.RESET}")
            else:
                if status == 200:
                    response = json.loads(body)
                    token = response.get("key", "")
                    print(f"{Colors.GREEN}Connexion réussie ! Votre clé est : {Colors.YELLOW}{token}{Colors.RESET}")
                else:
                    print(f"{Colors.RED}Erreur lors de la connexion : {body}{Colors.RESET}")

    except urllib.error.HTTPError as e:
        if is_register and e.code == 500:
            print(f"{Colors.RED}Erreur : Ce pseudo est déjà pris. Veuillez en choisir un autre.{Colors.RESET}")
            time.sleep(3)
            main()
        elif not is_register and e.code == 500 or e.code == 400:
            print(f"{Colors.RED}Pseudo ou mot de passe incorrect. Réessayez...{Colors.RESET}")
            time.sleep(3)
            main()
        else:
            print(f"{Colors.RED}HTTP error: {e.code}{Colors.RESET}")
    except Exception as e:
        print(f"{Colors.RED}Request failed: {e}{Colors.RESET}")
    except KeyboardInterrupt:
        print(f"\n{Colors.YELLOW}Déconnexion...{Colors.RESET}")
        sys.exit(0)

main()
