# docker-starter

Jednoduchá Go REST API aplikace s jednou POST routou `/start`.

## Požadavky
- Go 1.22 nebo novější

## Instalace závislostí
```
go mod tidy
```

## Spuštění aplikace
```
go run main.go
```

## Spuštění aplikace s parametry
```
go run main.go --login-script /cesta/k/login.sh --compose-dir /cesta/k/compose_souborum [--port 8080]
```
- `--port` je volitelný, výchozí je 8080.

## API
### POST /start
Přijímá JSON:
```json
{
  "project": "nazev_projektu",
  "image": "nazev_image"
}
```

### Co se stane po zavolání:
1. Spustí se login script (sh) z parametru `--login-script`.
2. V adresáři z `--compose-dir` se najde soubor `<project>.yaml`.
3. V tomto souboru se najde a přepíše řádek `image: ...` na hodnotu z payloadu.
4. Spustí se `docker compose up -d --force-recreate -f <soubor>`.

### Odpověď
Vrací potvrzení nebo chybovou hlášku.