# 🤖 Telegram bot 

[![CC BY-NC-SA 4.0][cc-by-nc-sa-shield]][cc-by-nc-sa]

[cc-by-nc-sa]: http://creativecommons.org/licenses/by-nc-sa/4.0/
[cc-by-nc-sa-image]: https://licensebuttons.net/l/by-nc-sa/4.0/88x31.png
[cc-by-nc-sa-shield]: https://img.shields.io/badge/License-CC%20BY--NC--SA%204.0-lightgrey.svg

[![stable](https://img.shields.io/badge/-stable-brightgreen?style=flat-square)](https://go-faster.org/docs/projects/status#stable)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/boeing666/telegram_bot/.github%2Fworkflows%2Fgo.yml?style=flat-square)
[![Release](https://img.shields.io/github/release/boeing666/telegram_bot.svg?style=flat-square)](https://github.com/boeing666/telegram_bot/releases)
![GitHub last commit](https://img.shields.io/github/last-commit/boeing666/telegram_bot?style=flat-square)

## 📘 Описание
Телеграм бот, написанный на golang, для сохранения важных сообщений среди чатов и каналов. Путем пересылки их в личный диалог с пользователем.
## 🎯 Цель
Облегчить поиск необходимых сообщений среди множества чатов и каналов в телеграме пользователя. 
## 📋 Функции
- ❗ **Определение важных сообщений** – По заданному пользователем набору ключевых слов бот определяет важные сообщения.
- 📨 **Направление сообщений в личный диалог** - Бот направляет важные сообщения в личный диалог с пользователем, предоставляя ключевую информацию об оригинальном сообщении и ссылку на него.
- 🛠️ **Настройка ключевых слов** - Бот сохраняет ключевые слова для конкретных каналов, позволяя пользователю настраивать.
## 🌟 Использование
1. Скачайте последнюю версию бота и разархивируйте [Releases](https://github.com/boeing666/telegram_bot/releases)
2. Создайте API ключи [Telegram API]([https://t.me/BotFather](https://my.telegram.org/auth)) api_id и api_hash
3. Создайте api_token [BotFather](https://t.me/BotFather)
4. Заполните конфиг данными ```configs/config.json```:
```yml
{
    "app_id": 0,
    "app_hash": "",
    "api_token": "",
    "phone_number": "",
    "db_settings": {
        "host": "",
        "username": "",
        "password": "",
        "database": ""
    }
}
```
5. Запустите бота
