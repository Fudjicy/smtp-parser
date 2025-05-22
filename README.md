# 🔍 SMTP Log Analyzer

Утилита для быстрого поиска в SMTP-логах с поддержкой параллельной обработки и цветовым выводом

## 🌟 Особенности

| Функция               | Описание                                  | Иконка |
|-----------------------|-------------------------------------------|--------|
| **Поиск по email**    | Точный поиск по адресу получателя/отправителя | 🔍    |
| **Фильтр по дате**    | Поиск записей за конкретную дату          | 📅    |
| **Рекурсивный поиск** | Обработка вложенных папок                 | 📂    |
| **Параллелизм**       | Обработка 50+ файлов одновременно         | ⚡    |
| **Цветовой вывод**    | Подсветка совпадений в терминале          | 🎨    |

## 🚀 Быстрый старт

### Требования
- Установленный Go 1.20+
- Терминал с поддержкой ANSI-цветов

### Установка

#Клонировать репозиторий
git clone https://github.com/Fudjicy/smtp-parser/smtp-log-analyzer.git
cd smtp-log-analyzer

# Компиляция (выберите вашу ОС)
# Linux/macOS
go build -o smtp-analyzer main.go

# Windows
go build -o smtp-analyzer.exe main.go

🛠 Использование
# Базовый синтаксис:
./smtp-analyzer -folder <путь_к_папке> -email <email> [-date <дата>]

# Поиск по email
./smtp-analyzer -folder ./logs -email user@example.com

# Поиск с фильтром по дате
./smtp-analyzer -folder /var/logs -email admin@company.com -date 2023-10-05

# Windows (PowerShell)
.\smtp-analyzer.exe -folder "C:\Logs" -email "john@doe.com"
