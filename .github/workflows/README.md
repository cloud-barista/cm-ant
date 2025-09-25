# GitHub Actions Workflows

이 디렉토리는 CM-ANT 프로젝트의 GitHub Actions 워크플로우를 포함합니다.

## 기존 워크플로우

### 1. Continuous Integration (CI)
- **파일**: `continuous-integration.yaml`
- **트리거**: Pull Request 생성/업데이트
- **기능**: 소스 코드 빌드 및 컨테이너 이미지 빌드 테스트

### 2. Continuous Delivery (CD)
- **파일**: `continuous-delivery.yaml`
- **트리거**: 
  - `main` 브랜치 푸시
  - `v*.*.*` 형식의 태그 푸시
- **기능**: Docker Hub 및 GHCR에 컨테이너 이미지 배포

## 새로운 워크플로우

### 3. Retag Release
- **파일**: `retag-release.yaml`
- **트리거**: 수동 실행 (`workflow_dispatch`)
- **기능**: 기존 태그를 삭제하고 최신 커밋으로 재생성

#### 사용법:
1. GitHub 리포지토리의 **Actions** 탭으로 이동
2. **Retag Release** 워크플로우 선택
3. **Run workflow** 버튼 클릭
4. 입력값 설정:
   - `tag_name`: 재생성할 태그명 (예: `v0.4.0`)
   - `confirm_deletion`: `DELETE` 입력 (확인용)

#### 주의사항:
- ⚠️ **기존 태그가 완전히 삭제됩니다**
- Docker Hub에서 해당 태그의 모든 이미지가 제거됩니다
- 삭제 후 30초 대기 후 새 이미지 빌드 시작

### 4. Force Rebuild Tag
- **파일**: `force-rebuild.yaml`
- **트리거**: 수동 실행 (`workflow_dispatch`)
- **기능**: 기존 태그를 유지하면서 최신 코드로 강제 재빌드

#### 사용법:
1. GitHub 리포지토리의 **Actions** 탭으로 이동
2. **Force Rebuild Tag** 워크플로우 선택
3. **Run workflow** 버튼 클릭
4. 입력값 설정:
   - `tag_name`: 재빌드할 태그명 (예: `v0.4.0`)
   - `force_rebuild`: `REBUILD` 입력 (확인용)

#### 특징:
- 기존 태그는 유지됩니다
- 캐시를 사용하지 않고 완전히 새로 빌드합니다
- Docker Hub의 태그 충돌 문제를 해결합니다

## 문제 해결 시나리오

### 시나리오 1: Docker Hub에서 같은 태그에 여러 digest 존재
```bash
# 문제: v0.4.0 태그가 두 개의 다른 digest를 가리킴
# 해결: Retag Release 워크플로우 사용
```

### 시나리오 2: 최신 코드가 반영되지 않은 이미지
```bash
# 문제: 태그는 최신이지만 이미지 내용이 오래됨
# 해결: Force Rebuild Tag 워크플로우 사용
```

### 시나리오 3: 일반적인 새 릴리즈
```bash
# 방법: 기존 CD 워크플로우 사용
git tag v0.4.1
git push origin v0.4.1
```

## 필요한 GitHub Secrets

다음 Secrets이 설정되어 있어야 합니다:

- `DOCKER_USERNAME`: Docker Hub 사용자명
- `DOCKER_PASSWORD`: Docker Hub 비밀번호 또는 액세스 토큰
- `CR_PAT`: GitHub Container Registry Personal Access Token

## 워크플로우 실행 권한

- **cloud-barista** 조직의 멤버만 워크플로우를 실행할 수 있습니다
- 포크된 리포지토리에서는 실행할 수 없습니다 (보안상의 이유)

## 트러블슈팅

### 워크플로우가 실행되지 않는 경우
1. GitHub Secrets 설정 확인
2. 리포지토리 권한 확인
3. 입력값 형식 확인 (태그명은 `v0.4.0` 형식)

### Docker Hub API 오류
1. Docker Hub 계정 권한 확인
2. API 레이트 리미트 확인
3. 네트워크 연결 상태 확인
