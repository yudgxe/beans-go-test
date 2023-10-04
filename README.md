# beans-go-test
### Описание

Допушения, которые сделаны.
 1. Версии указаны без знаков сравнения.
 2. `targets` имеет единую структуру т.к это сильно упрощает парсинг.
 3. При `update` уже существующих пакетов, они будут перезаписаны.
 4. Авторизация только по паролю.
 5. Пустые папки игнорируются.
 6. Не реализована возможность исключения из шаблона. Т.к единственное [решение](https://github.com/gobwas/glob), которое я смог найти не работает с слешами.

Пример packet.json  
```
{
	"name": "packet-1",
	"ver": "1.10",
	"targets": [
		{"path": "./archive_this1/*.txt"},
		{"path": "./archive_this2/*.txt"}
	],
	"packets": [
		{ "name": "packet-2", "ver": "1.10" },
		{ "name": "packet-3", "ver": "1.10" }
	]
}
``` 
Где `packets` - это зависимости пакета. Другими словами, при вызове `update` для `packet-1` будут скачаны зависости этого пакета, а так же зависимости зависимых пакетов, если они есть. В данном случаи, это пакеты `packet-2` и `packet-3`.

Пример packages.json
```
{
	"packages": [
	 {"name": "packet-1", "ver": "1.10"},
	 {"name": "packet-4", "ver": "1.10"},
	 {"name": "packet-5", "ver": "1.10"}
	]
}
```

-----
### Пример вложености пакетов
При создание пакета 
```
{
	"name": "packet-1",
	"ver": "1.10",
	"targets": [
		{"path": "./archive_this1/*.txt"},
		{"path": "./archive_this2/*.txt"}
	],
	"packets": [
		{ "name": "packet-2", "ver": "1.10" },
		{ "name": "packet-3", "ver": "1.10" }
	]
}
``` 
и пакета
```
{
	"name": "packet-3",
	"ver": "1.10",
	"targets": [
		{"path": "./archive_this1/*.txt"},
		{"path": "./archive_this2/*.txt"}
	],
	"packets": [
		{ "name": "packet-4", "ver": "1.10" }
	]
}
``` 

И последующим иобновлении `packet-1` будут скачаны 4 пакета `packet-1` `packet-2` `packet-3` `packet-4`.   
Т.к `pakcet-1` зависит от `packet-2` и `packet-3`, а пакет `packet-3` зависит от `packet-4`.
```
{
	"packages": [
	 {"name": "packet-1", "ver": "1.10"}
	]
}
```

-----
### Запуск

В папке `server` есть Dockerfile с простым shh сервером на ubuntu, тестировалось все на нем.  
По умолчанию все настроено под него.
 
Для запуска: 
 1. docker build -t ssh_server .\server\.
 2. docker run -p 22:22 ssh_server


Для изменения поведения по умолчанию реализованы следующие флаги:
```
Usage:
  -a string
        Адрес сервера (default "localhost:22")
  -p string
        Пароль пользователя (default "123")
  -u string
        Имя пользователя (default "dev")
```
Всего существует две сабкоманды:
 1. create - отпавка пакета.
 2. update - скачивание пакета/пакетов.

Примеры: 
 1. go run .\cmd\app\main.go create ./packet.json
 2. go run .\cmd\app\main.go update ./packages.json
 3. 
-----
### Структруа хранения.
Пакеты хранятся в следующем виде.
```bash
├── packet-1
│   ├── 1.10
│   │	 ├── deps - зависимости, если есть
│   │	 └── packet.zip - упаковыный архив
│   └── 1.20
│   	 └── packet.zip	 
└── pakcet-2
    ├── 1.10
    │	 ├── deps - зависимости, если есть
    │	 └── packet.zip - упаковыный архив
    └── 1.20
         └── packet.zip	
```

