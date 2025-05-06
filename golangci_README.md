# Версия golangci-lint v2.0.1

```
Локально должна быть создана папка **golangci-lint** 
(она добавлена в .gitignore и пушиться не будет)
В ней будут храниться кэщ и отчеты.
```

Можно добавить флаг **-v** в команду запуска, тогда будут выводиться логи golangci-lint.

# Запуск под Linux  

Запускается из корня проекта

```shell
docker run --rm \
    -v $(pwd):/app \
    -v $(pwd)/golangci-lint/.cache/golangci-lint/v2.0.1:/root/.cache \
    -w /app \
    golangci/golangci-lint:v2.0.1 \
        golangci-lint run \
            -c .golangci.yml 
```
---
# Запуск под windows

**Не забыть запустить docker desktop.**

Запускается из корня проекта через **PowerShell**

```shell
docker run --rm ` 
    -v ${pwd}:/app `
    -v ${pwd}/golangci-lint/.cache/golangci-lint/v2.0.1:/root/.cache `
    -w /app `
    golangci/golangci-lint:v2.0.1 `
        golangci-lint run `
            -c .golangci.yml 
```
---
# Запуск под Windows через WSL

Запускается из терминала wsl так же как и под линукс. 
  
### Если не запускается:


* Если повезёт, то хватит [этой статьи](https://docs.docker.com/desktop/features/wsl/)
* Если не повезло - придется накатить docker отдельно на WSL

---

# В случае проблем с линтерами

1. Попробовать обновить образ: `docker pull golangci/golangci-lint:v2.0.1` (LINUX/WSL)