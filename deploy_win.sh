REM Скрипт автоматического развертывания

REM Установка зависимостей
REM Для Windows управление пакетами может отличаться
REM Пример:
REM choco install golang -y

REM Клонирование репозитория
git clone https://github.com/DoktorAssering/LoadTestingGO.git
cd LoadTestingGO

REM Установка зависимостей для каждого модуля
cd ..\agent
go mod tidy

cd ..\controller
go mod tidy

cd ..\load_service
go mod tidy

cd ..

go mod tidy

REM Компиляция и запуск основного файла проекта
go build -o main .
start main.exe
