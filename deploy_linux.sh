#!/bin/bash

# Скрипт автоматического развертывания

# Установка зависимостей
sudo apt update
sudo apt install -y golang git

# Клонирование репозитория
git clone https://github.com/DoktorAssering/LoadTestingGO.git
cd LoadTestingGO

# Установка зависимостей для каждого модуля
cd ../agent
go mod tidy

cd ../controller
go mod tidy

cd ../load_service
go mod tidy

cd ..
go mod tidy

# Компиляция и запуск основоного файла проекта
go build -o main .
./main &