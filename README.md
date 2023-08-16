# Telegram Messenger

## Introduction

This program enables you to manage your SMS via a Telegram bot, including tasks such as receiving and sending SMS. To use this program, you will need a Telegram account and a Telegram bot. You can create a Telegram bot by interacting with [BotFather](https://t.me/botfather) and following the provided instructions.

I have thoroughly tested this program and found it to work well. However, its compatibility with your system may vary. Should you encounter any issues, please do not hesitate to inform me.

## Tested Devices

* Qualcomm 410 WiFi Stick
* Quectel EM12-G

## Installation & Usage

You can obtain the latest release from the [releases page](https://github.com/damonto/telegram-messenger/releases).

Sometimes, you might need to set executable permissions for the binary file using the following command:

```bash
chmod +x telegram-messenger
```

Once done, you can run the program with root privileges:

```bash
sudo ./telegram-messenger -token=YourTelegramToken -chat-id=YourTelegramChatID -sim=YourSIMId
```

If you wish to run the program in the background, you can utilize the `systemctl` command. Here is an example of how to achieve this:

1. Start by creating a service file in the `/etc/systemd/system` directory. For instance, you can name the file `telegram-messenger.service` and include the following content:

```plaintext
[Unit]
Description=Telegram Messenger SMS Manager
After=network.target

[Service]
Type=simple
User=root
Restart=on-failure
ExecStart=/your/binary/path/here/telegram-messenger -token=YourTelegramToken -chat-id=YourTelegramChatID -sim=YourSIMId
RestartSec=10s
TimeoutStopSec=30s

[Install]
WantedBy=multi-user.target
```

2. Then, use the following command to start the service:

```bash
sudo systemctl start telegram-messenger
```

3. If you want the service to start automatically upon system boot, use the following command:

```bash
sudo systemctl enable telegram-messenger
```
