# Environment File Configuration

Цей контролер підтримує конфігурацію через env файли та змінні середовища.

## Використання

### 1. Через env файл

Створіть файл `.env` з конфігурацією:

```bash
# Required parameters
POD_NAME=redis-master
POD_CONTAINER_NAME=redis
POD_IMAGE=redis
POD_TAG=1.2.3
POD_PORT=30555

# Optional parameters
POD_NAMESPACE=databases
POD_CONFIGMAP=redis-config
POD_OUTPUT=pod.yaml  # or "stdout" for console output
POD_VERBOSE=false    # true for debug output
```

Запустіть команду з env файлом:

```bash
./controller generate-pod-yaml --env-file=.env
```

### 2. Через змінні середовища

Встановіть змінні середовища:

```bash
export POD_NAME=redis-master
export POD_CONTAINER_NAME=redis
export POD_IMAGE=redis
export POD_TAG=1.2.3
export POD_PORT=30555
export POD_NAMESPACE=databases
export POD_VERBOSE=true
```

Запустіть команду:

```bash
./controller generate-pod-yaml
```

### 3. Комбінація флагів та env файлу

Флаги командного рядка мають пріоритет над env файлом та змінними середовища:

```bash
# Перевизначення вербозності
./controller generate-pod-yaml --env-file=.env --verbose

# Перевизначення вихідного файлу
./controller generate-pod-yaml --env-file=.env --output=my-pod.yaml

# Вивід в консоль замість файлу
./controller generate-pod-yaml --env-file=.env --output=stdout
```

## Керування виводом

| Значення `POD_OUTPUT` | Результат |
|----------------------|-----------|
| `pod.yaml` | Зберігає в файл `pod.yaml` |
| `stdout` | Виводить в консоль |
| Порожнє або не встановлене | Виводить в консоль |

## Керування вербозністю

| Значення `POD_VERBOSE` | Результат |
|----------------------|-----------|
| `true`, `1`, `yes` | Увімкнено детальне логування |
| `false`, `0`, `no` | Тільки основні повідомлення |

**Примітка:** Флаг `--verbose` перевизначає налаштування з env файлу.

## Змінні середовища

| Змінна | Опис | Обов'язкова |
|--------|------|-------------|
| `POD_NAME` | Назва Pod | ✅ |
| `POD_CONTAINER_NAME` | Назва контейнера | ✅ |
| `POD_IMAGE` | Образ контейнера | ✅ |
| `POD_TAG` | Тег образу | ✅ |
| `POD_PORT` | Порт контейнера | ✅ |
| `POD_NAMESPACE` | Namespace (за замовчуванням: default) | ❌ |
| `POD_CONFIGMAP` | Назва ConfigMap | ❌ |
| `POD_OUTPUT` | Шлях до вихідного файлу або "stdout" | ❌ |
| `POD_VERBOSE` | Увімкнути детальне логування | ❌ |

## Правила іменування

- Всі змінні середовища мають префікс `POD_`
- Дефіси (`-`) в назвах флагів замінюються на підкреслення (`_`) в змінних середовища
- Приклад: `--pod-name` → `POD_NAME`

## Приклади використання

### Вивід в консоль з детальним логуванням:
```bash
./controller generate-pod-yaml --env-file=.env --output=stdout --verbose
```

### Збереження в файл без детального логування:
```bash
./controller generate-pod-yaml --env-file=.env --output=my-pod.yaml
```

### Використання змінних середовища:
```bash
export POD_OUTPUT=stdout
export POD_VERBOSE=true
./controller generate-pod-yaml --env-file=.env
``` 