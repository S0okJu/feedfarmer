# FeedFarmer Helm 배포

## 1) 이미지 준비
이 차트는 기본값으로 `feedfarmer:latest` 이미지를 사용합니다.

```bash
docker build -t <REGISTRY>/feedfarmer:<TAG> .
docker push <REGISTRY>/feedfarmer:<TAG>
```

## 2) 설치

```bash
helm upgrade --install feedfarmer ./charts/feedfarmer \
  --namespace feedfarmer --create-namespace \
  --set image.repository=<REGISTRY>/feedfarmer \
  --set image.tag=<TAG>
```

## 3) Ingress 활성화 (옵션)

```bash
helm upgrade --install feedfarmer ./charts/feedfarmer \
  --namespace feedfarmer --create-namespace \
  --set image.repository=<REGISTRY>/feedfarmer \
  --set image.tag=<TAG> \
  --set ingress.enabled=true \
  --set ingress.className=nginx \
  --set ingress.hosts[0].host=feedfarmer.example.com
```

## 4) 주요 values

- `replicaCount`: 기본 `1` (SQLite 단일 쓰기 특성으로 권장)
- `persistence.enabled`: 기본 `true`
- `persistence.size`: 기본 `5Gi`
- `env.dbPath`: 기본 `/data/feedfarmer.db`
- `service.type`: 기본 `ClusterIP`

## 5) 값 파일로 운영 환경 분리

```bash
helm upgrade --install feedfarmer ./charts/feedfarmer \
  --namespace feedfarmer \
  -f values-prod.yaml
```
