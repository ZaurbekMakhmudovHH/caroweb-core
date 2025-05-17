#!/bin/bash

# === Аргументы ===
STAGE="$1"
CORE_BRANCH="$2"
WEB_BRANCH="$3"
ADMIN_BRANCH="$4"
DOCKER_BRANCH="$5"

if [ -z "$STAGE" ]; then
  echo "❌ Использование: ./deploy.sh <STAGE> [core_branch] [web_branch] [admin_branch] [docker_branch]"
  exit 1
fi

# === Настройки ===
TARGET_PATH="/var/www/$STAGE"
GIT_HOST="teblogelsouy.beget.app"
GIT_GROUP="carowebapp"

# === Подстановка master, если ветки не указаны ===
CORE_BRANCH="${CORE_BRANCH:-master}"
WEB_BRANCH="${WEB_BRANCH:-master}"
ADMIN_BRANCH="${ADMIN_BRANCH:-master}"
DOCKER_BRANCH="${DOCKER_BRANCH:-master}"

# === Репозитории и ветки ===
REPOS=("service-core" "web" "admin-panel" "docker")
BRANCHES=("$CORE_BRANCH" "$WEB_BRANCH" "$ADMIN_BRANCH" "$DOCKER_BRANCH")

mkdir -p "$TARGET_PATH"

# === Функция клонирования/обновления ===
sync_repo() {
  local name="$1"
  local branch="$2"
  local path="$TARGET_PATH/$name"

  if [ ! -d "$path/.git" ]; then
    echo "📥 Клонируем $name ($branch)..."
    git clone -b "$branch" "git@$GIT_HOST:$GIT_GROUP/$name.git" "$path"
  else
    echo "🔁 Обновляем $name ($branch)..."
    cd "$path"
    git fetch origin
    git checkout "$branch"
    git pull origin "$branch"
  fi
}

# === Клонирование / обновление всех репозиториев ===
for i in "${!REPOS[@]}"; do
  sync_repo "${REPOS[$i]}" "${BRANCHES[$i]}"
done

# === Копируем .env файл для docker ===
ENV_SOURCE="/var/www/env/.env.docker-$STAGE"
ENV_TARGET="$TARGET_PATH/docker/.env"

if [ -f "$ENV_SOURCE" ]; then
  echo "📦 Копируем $ENV_SOURCE → $ENV_TARGET"
  cp "$ENV_SOURCE" "$ENV_TARGET"
else
  echo "⚠️  Файл $ENV_SOURCE не найден. Деплой остановлен."
  exit 1
fi

# === Устанавливаем PORT_CORE по STAGE ===
case "$STAGE" in
  d1) PORT_CORE=8081 ;;
  d2) PORT_CORE=8082 ;;
  d3) PORT_CORE=8083 ;;
  stage) PORT_CORE=8084 ;;
  *) PORT_CORE=8090 ;;
esac

# === Путь до .env внутри service-core ===
CORE_ENV_PATH="$TARGET_PATH/service-core/.env"
DB_NAME="$STAGE"
DB_USER="$STAGE"
DB_PASSWORD="your_password_$STAGE"

# === Убедиться, что .env существует ===
if [ ! -f "$CORE_ENV_PATH" ]; then
  echo "⚠️  .env в $CORE_ENV_PATH не найден — создаём новый"
  touch "$CORE_ENV_PATH"
fi

# === Обновляем значения подключения к БД ===
sed -i "s/^DB_HOST=.*/DB_HOST=172.17.0.1/" "$CORE_ENV_PATH"
sed -i "s/^DB_PORT=.*/DB_PORT=5432/" "$CORE_ENV_PATH"
sed -i "s/^DB_USER=.*/DB_USER=$DB_USER/" "$CORE_ENV_PATH"
sed -i "s/^DB_PASSWORD=.*/DB_PASSWORD=$DB_PASSWORD/" "$CORE_ENV_PATH"
sed -i "s/^DB_NAME=.*/DB_NAME=$DB_NAME/" "$CORE_ENV_PATH"
sed -i "s/^SMTP_HOST=.*/SMTP_HOST=172.17.0.1/" "$CORE_ENV_PATH"


echo "🔧 Обновлены переменные БД в $CORE_ENV_PATH"

# === Запуск docker compose ===
cd "$TARGET_PATH/docker"
echo "🧹 Остановка предыдущих контейнеров..."
docker compose -f docker-compose.stage.yml down || true

echo "🚀 Запуск docker-compose.stage.yml..."
docker compose -f docker-compose.stage.yml pull
docker compose -f docker-compose.stage.yml up -d --build


echo "✅ Деплой $STAGE завершён"
