# LinguaTest
LinguaTest — это веб-решение для изучения 4 иностранных языков (русский, английский, итальянский, испанский, монгольский) в формате тестов. Проект построен на микросервисной архитектуре с использованием gRPC для взаимодействия между сервисами.
## Архитектура проекта
- Auth Service (Go): Управление аутентификацией (регистрация, логин, токены).
- User Service (Python): Управление профилями пользователей и уровнями знаний.
- Test Service (Python): Предоставление тестов и обработка ответов.
- Frontend (Vite): Многостраничное приложение с интерфейсом (стартовая страница, тесты).
- Database: PostgreSQL для хранения данных.
- Envoy: Прокси для gRPC-web.

## Требования
**Зависимости**

- Go: 1.23.9+
- Python: 3.9+
- Node.js: 14+
- PostgreSQL: 14+
- Envoy: v1.29.8
- protoc: Для генерации gRPC-кода
- Docker: Для контейнеризации


## Установка

Клонируйте репозиторий:
```
git clone https://github.com/yourusername/lingua-test.git
cd lingua-test
```


**Установите зависимости**:
- Auth Service:
```
cd auth-service
go mod tidy
```

- User Service:
```
cd user-service
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

- Test Service:
```
cd test-service
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```


- Frontend:
```
cd frontend
npm install
```



**Настройте базу данных**:

Запустите PostgreSQL.
Создайте базы данных:
```
CREATE DATABASE auth_db;
CREATE DATABASE user_db;
CREATE DATABASE test_db;
```

Примените миграции:
```
psql -U postgres -d auth_db -f auth-service/db/schema.sql
psql -U postgres -d user_db -f user-service/db/schema.sql
psql -U postgres -d test_db -f test-service/db/schema.sql
```



Сгенерируйте gRPC-код (все 3):
```
protoc --go_out=auth-service/gen/go --go_opt=paths=source_relative \
--go-grpc_out=auth-service/gen/go --go-grpc_opt=paths=source_relative \
proto/sso.proto
```
```
protoc --python_out=user-service/gen/python --grpc_python_out=user-service/gen/python \
proto/user.proto
```
```
protoc --python_out=test-service/gen/python --grpc_python_out=test-service/gen/python \
proto/test.proto
```


## Запуск

Запустите Envoy:
```
docker run -d -p 8080:8080 -v $(pwd)/docker/envoy.yaml:/etc/envoy/envoy.yaml envoyproxy/envoy:v1.29.8
```


**Запустите сервисы:**

- Auth Service:
```
cd auth-service
go run cmd/main.go
```

- User Service:
```
cd user-service
source venv/bin/activate
python src/main.py
```

- Test Service:
```
cd test-service
source venv/bin/activate
python src/main.py
```



- Запустите фронтенд:
```
cd frontend
npm run dev
```

Откройте http://localhost:3000/main для стартовой страницы.



## Использование

- Перейдите на стартовую страницу (/main) и зарегистрируйтесь/войдите.
- Перейдите на страницу тестов (/tests), выберите язык и начните тестирование.
- В профиле пользователя можно просмотреть уровень знаний и историю тестов.


## Команда

- **Аналитик/системный аналитик:** Дарья Исаева
- **Frontend-разработчик:** Сергей Телегин
- **Backend-разработчики:** Вероника Машталер и Сергей Телегин

## Лицензия
[MIT License](LICENSE)
